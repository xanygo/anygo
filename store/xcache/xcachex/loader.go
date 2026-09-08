package xcachex

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xcontainer"
	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xslice"
	"github.com/xanygo/anygo/xtime"
)

var globalConfigFile *ConfigFile
var configErr error
var configOnce sync.Once

func loadConfig() {
	globalConfigFile = &ConfigFile{}
	err := xcfg.Parse("store/xcache", &globalConfigFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		configErr = err
	}
}

// Load 依据名字初始化并加载 Cache 对象。使用配置文件 {confDir}/store/xcache.{json|yml|toml}
func Load[K comparable, V any](name string) (xcache.MCache[K, V], error) {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return nil, configErr
	}
	return globalConfigFile.Load[K, V](name)
}

func MustLoad[K comparable, V any](name string) xcache.MCache[K, V] {
	c, err := Load[K, V](name)
	if err != nil {
		panic(err)
	}
	return c
}

// CheckConfig 检查 store/xcache 配置文件是否正确
func CheckConfig() error {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return configErr
	}
	return globalConfigFile.Check()
}

// MustCheckConfig 检查 store/xcache 配置文件是否正确，若不正确则panic
func MustCheckConfig() {
	err := CheckConfig()
	if err != nil {
		panic(fmt.Errorf("check store/xcache: %w", err))
	}
}

type instanceKey[K comparable, V any] struct {
	K K
	V V
	N string
}

type instanceValue struct {
	C any
	E error
}

type ConfigFile struct {
	Items    []map[string]any `json:"Items" yaml:"Items"`
	instance sync.Map
	refs     xcontainer.DepGraph[string]
}

func (cf *ConfigFile) MustLoad[K comparable, V any](name string) xcache.MCache[K, V] {
	kv, err := cf.Load[K, V](name)
	if err != nil {
		panic(err)
	}
	return kv
}

func (cf *ConfigFile) Load[K comparable, V any](name string) (xcache.MCache[K, V], error) {
	key := instanceKey[K, V]{
		N: name,
	}
	if val, ok := cf.instance.Load(key); ok {
		v := val.(*instanceValue)
		if v.E != nil {
			return nil, v.E
		}
		return v.C.(xcache.MCache[K, V]), nil
	}
	c, err := cf.createCache[K, V](name)
	old, loaded := cf.instance.LoadOrStore(key, &instanceValue{C: c, E: err})
	if !loaded {
		rn := xcache.NewName[K, V](name)
		xcache.Registry().TryRegister(rn, c)
		return c, err
	}
	v := old.(*instanceValue)
	if v.E != nil {
		return nil, v.E
	}
	return v.C.(xcache.MCache[K, V]), nil
}

func (cf *ConfigFile) Check() error {
	clone := &ConfigFile{
		Items: slices.Clone(cf.Items),
	}
	var errs []error
	for idx, item := range clone.Items {
		name, _ := xmap.GetString(item, "Name")
		if name == "" {
			err := fmt.Errorf("[%d].Name is empty: %v", idx, item)
			errs = append(errs, err)
			continue
		}
		_, err := clone.newCache1[string, string](name, item)
		if err != nil {
			err = fmt.Errorf("[%d]=%q: %w", idx, name, err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (cf *ConfigFile) createCache[K comparable, V any](name string) (xcache.MCache[K, V], error) {
	for _, item := range cf.Items {
		str, _ := xmap.GetString(item, "Name")
		if name != str {
			continue
		}
		c, err := cf.newCache1[K, V](name, item)
		if err != nil {
			err = fmt.Errorf("load xcache %q: %w", name, err)
		}
		return c, err
	}
	return nil, fmt.Errorf("load xcache %q: %w, Items.len=%d", name, xerror.NotFound, len(cf.Items))
}

func (cf *ConfigFile) newCache1[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	tp, _ := xmap.GetString(item, "Type")
	if tp == "" {
		return nil, fmt.Errorf("invalid cache type: %v", item)
	}
	switch tp {
	case "Nop":
		return &xcache.Nop[K, V]{}, nil
	case "Chains":
		return cf.newChains[K, V](name, item)
	case "Wrap":
		return cf.newWrap[K, V](name, item)
	default:
		life, err := cf.parserLife(item)
		if err != nil {
			return nil, err
		}
		cc, err := cf.newCache2[K, V](tp, name, item)
		if err != nil || life.empty() {
			return cc, err
		}
		return life.wrap(cc), nil
	}
}

func (cf *ConfigFile) newCache2[K comparable, V any](tp string, name string, item map[string]any) (xcache.MCache[K, V], error) {
	switch tp {
	case "File":
		fc := &xcache.File[K, V]{}
		return fc, fc.Init(item)
	case "MemoryLRU":
		mc := xcache.NewLRU[K, V](1) // 容量设置为1，最终采用配置中的值
		return mc, mc.Init(item)
	case "MemoryFIFO":
		mc := xcache.NewMemoryFIFO[K, V](1)
		return mc, mc.Init(item)
	case "MemoryLIFO":
		mc := xcache.NewMemoryLIFO[K, V](1)
		return mc, mc.Init(item)
	case "Redis":
		return cf.newRedis[K, V](name, item)
	case "DB":
		return cf.newDB[K, V](name, item)
	default:
		return nil, fmt.Errorf("newCache wth unsupport Type=%q", tp)
	}
}

func (cf *ConfigFile) newRedis[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	rc := &Redis{}
	if err := rc.Init(item); err != nil {
		return nil, err
	}
	tr := &xcache.Transformer[K, V]{
		Cache: rc,
	}

	sp := make(map[string]any, 1)
	codec, err := xmap.GetString(item, "Codec")
	if err != nil {
		return nil, err
	}
	if codec != "" {
		sp["ValueCodec"] = codec
	}
	return tr, tr.Init(sp)
}

func (cf *ConfigFile) newDB[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	dc := &Database{
		TypeID: zreflect.TypeID2[K, V](),
	}
	if err := dc.Init(item); err != nil {
		return nil, err
	}
	tr := &xcache.Transformer[K, V]{
		Cache: dc,
	}
	sp := make(map[string]any, 1)
	codec, err := xmap.GetString(item, "Codec")
	if err != nil {
		return nil, err
	}
	if codec != "" {
		sp["ValueCodec"] = codec
	}
	return tr, tr.Init(sp)
}

func (cf *ConfigFile) newChains[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	val, ok := xmap.Get(item, "Chains")
	if !ok || xslice.Len(val) == 0 {
		return nil, errors.New("missing [Chains] section")
	}
	var chains []*xcache.Chain[K, V]
	var errs []error
	xslice.Range[any](val, func(val any) bool {
		zc, err := xcodec.ConvertAs[*chainItemConfig](val)
		if err != nil {
			errs = append(errs, err)
			return false
		}
		if zc.Ref == "" {
			errs = append(errs, fmt.Errorf("required Ref in %v", val))
			return false
		}
		if zc.Life.empty() {
			errs = append(errs, fmt.Errorf("required Life in %v", val))
			return false
		}

		if zc.Life.Default <= 0 && zc.Life.Min <= 0 && zc.Life.Force <= 0 {
			errs = append(errs, fmt.Errorf("required Life.[Default|Min|Force] in %v", val))
			return false
		}

		if err := cf.refs.Add(name, zc.Ref); err != nil {
			errs = append(errs, err)
			return false
		}

		c, err := cf.Load[K, V](zc.Ref)
		if err != nil {
			errs = append(errs, fmt.Errorf("load Chains cache %q failed: %w", zc.Ref, err))
			return false
		}
		ci := &xcache.Chain[K, V]{
			Cache:        c,
			NewLifeFn:    zc.Life.newLifeFn[K],
			WriteTimeout: zc.WriteTimeout.Duration(),
		}

		chains = append(chains, ci)
		return true
	})
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	cc := xcache.NewChains[K, V](chains...)
	return xcache.AsMCache(cc, 10), nil
}

func (cf *ConfigFile) newWrap[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	ref, err := xmap.GetString(item, "Ref")
	if err != nil {
		return nil, err
	}
	if ref == "" {
		return nil, errors.New("missing 'Ref'")
	}
	c, err := cf.Load[K, V](ref)
	if err != nil {
		return nil, err
	}

	if err := cf.refs.Add(name, ref); err != nil {
		return nil, err
	}

	life, err := cf.parserLife(item)
	if err != nil {
		return nil, err
	}

	keyTransform, _ := xmap.GetMap(item, "KeyTransform")
	if len(keyTransform) == 0 && life.empty() {
		// 两者同时为空，返回原始的
		return c, nil
	}
	wp := &xcache.Wrapper[K, V]{
		Cache: c,
	}
	if len(keyTransform) > 0 {
		if err = cf.parserWrapKeyTransform(wp, keyTransform); err != nil {
			return nil, err
		}
	}

	if life != nil {
		wp.NewLifeFn = life.newLifeFn[K]
	}
	return wp, nil
}

func (cf *ConfigFile) parserWrapKeyTransform[K comparable, V any](wp *xcache.Wrapper[K, V], cfg map[string]any) error {
	kt := reflect.TypeFor[K]()
	rule := kt.String()
	param, err := xmap.GetMap(cfg, rule)
	if err != nil {
		return err
	}
	if len(param) > 0 {
		rp, ok := keyTransformFns[rule]
		if !ok {
			return fmt.Errorf("%w: KeyTransform %q", xerror.NotFound, rule)
		}
		rpFn, ok := rp.(func(p map[string]any) (func(K) K, error))
		if !ok {
			return fmt.Errorf("KeyTransform %q type not match, got %T, expect %T", rule, rp, rpFn)
		}
		fn, err := rpFn(param)
		if err != nil {
			return err
		}
		wp.NewKeyFn = fn
		return nil
	}

	if kt.Kind() == reflect.String {
		// 用于支持 type MyString string 这种自定义 string 类型的 key
		param, err := xmap.GetMap(cfg, "string")
		if err != nil {
			return err
		}
		fn := stringKindTransform(param)
		if fn == nil {
			return nil
		}
		wp.NewKeyFn = func(k K) K {
			krv := reflect.ValueOf(k)
			newKrv := fn(krv)
			if krv.Equal(newKrv) {
				return k
			}
			return newKrv.Interface().(K)
		}
	}
	child, err := xmap.GetMap(cfg, "Default")
	if err != nil {
		return err
	}
	if len(child) > 0 {
		if child["Panic"] == true {
			err := fmt.Errorf("%w KeyTransform func for type %q, pls use xcache.RegisterKeyTransform first", xerror.NotFound, rule)
			panic(err)
		} else if child["Refuse"] == true {
			return fmt.Errorf("%w KeyTransform func for type %q, refused, pls use xcache.RegisterKeyTransform first", xerror.NotFound, rule)
		}
	}
	return nil
}

func (cf *ConfigFile) parserLife(cfg map[string]any) (*lifeConfig, error) {
	param, err := xmap.GetMap(cfg, "Life")
	if err != nil || len(param) == 0 {
		return nil, err
	}
	return xcodec.ConvertAs[*lifeConfig](param)
}

type lifeConfig struct {
	Default xtime.Duration // 默认值，当 ttl 为 0 时生效
	Force   xtime.Duration // 优先级最高
	Min     xtime.Duration
	Max     xtime.Duration
}

func (lc *lifeConfig) empty() bool {
	return lc == nil || (lc.Default <= 0 && lc.Force <= 0 && lc.Min <= 0 && lc.Max <= 0)
}

func (lc *lifeConfig) newLifeFn[K comparable](k K, t time.Duration) time.Duration {
	if t == 0 && lc.Default > 0 {
		return lc.Default.Duration()
	}
	if lc.Force > 0 {
		return lc.Force.Duration()
	}
	if lc.Min > 0 && t < lc.Min.Duration() {
		return lc.Min.Duration()
	}

	if lc.Max > 0 && t > lc.Max.Duration() {
		return lc.Max.Duration()
	}

	return t
}

func (lc *lifeConfig) wrap[K comparable, V any](c xcache.MCache[K, V]) xcache.MCache[K, V] {
	return &xcache.Wrapper[K, V]{
		Cache:     c,
		NewLifeFn: lc.newLifeFn[K],
	}
}

type chainItemConfig struct {
	Ref          string
	Life         *lifeConfig
	WriteTimeout xtime.Duration
}

var keyTransformFns = map[string]any{}

func RegisterKeyTransform[K comparable](fn func(p map[string]any) (func(K) K, error)) error {
	tp := reflect.TypeFor[K]().String()
	if fn == nil {
		return fmt.Errorf("cannot RegisterKeyTransform %q with nil func", tp)
	}
	if _, ok := keyTransformFns[tp]; ok {
		return fmt.Errorf("key transform function %q is already registered", tp)
	}
	keyTransformFns[tp] = fn
	return nil
}

func MustRegisterKeyTransform[K comparable](fn func(p map[string]any) (func(K) K, error)) {
	err := RegisterKeyTransform(fn)
	if err != nil {
		panic(err)
	}
}

func init() {
	MustRegisterKeyTransform(stringKeyTransform)
}

func stringKeyTransform(p map[string]any) (func(string) string, error) {
	prefix, _ := xmap.GetString(p, "Prefix")
	prefix = strings.TrimSpace(prefix)
	suffix, _ := xmap.GetString(p, "Suffix")
	suffix = strings.TrimSpace(suffix)
	if prefix == "" && suffix == "" {
		return nil, nil
	}
	return func(k string) string {
		return prefix + k + suffix
	}, nil
}

func stringKindTransform(p map[string]any) func(reflect.Value) reflect.Value {
	if len(p) == 0 {
		return nil
	}
	prefix, _ := xmap.GetString(p, "Prefix")
	prefix = strings.TrimSpace(prefix)
	suffix, _ := xmap.GetString(p, "Suffix")
	suffix = strings.TrimSpace(suffix)
	if prefix == "" && suffix == "" {
		return nil
	}
	return func(k reflect.Value) reflect.Value {
		s := prefix + k.String() + suffix
		v := reflect.New(k.Type()).Elem()
		v.SetString(s)
		return v
	}
}
