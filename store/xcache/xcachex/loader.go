package xcachex

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xcontainer"
	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xtype"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
)

var globalConfigFile *ConfigFile
var configErr error
var configOnce sync.Once

func loadConfig() {
	globalConfigFile = &ConfigFile{}
	configErr = xcfg.Parse("xcache", &globalConfigFile)
}

func Load[K comparable, V any](name string) (xcache.MCache[K, V], error) {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return nil, configErr
	}
	return globalConfigFile.Load[K, V](name)
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

	refs xcontainer.DepGraph[string]
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
		return c, err
	}
	v := old.(*instanceValue)
	if v.E != nil {
		return nil, v.E
	}
	return v.C.(xcache.MCache[K, V]), nil
}

func (cf *ConfigFile) createCache[K comparable, V any](name string) (xcache.MCache[K, V], error) {
	for _, item := range cf.Items {
		str, _ := xmap.GetString(item, "Name")
		if name != str {
			continue
		}
		c, err := cf.newCache[K, V](name, item)
		if err != nil {
			err = fmt.Errorf("cache Name=%q: %w", name, err)
		}
		return c, err
	}
	return nil, fmt.Errorf("%w for cache %q", xerror.NotFound, name)
}

func (cf *ConfigFile) newCache[K comparable, V any](name string, item map[string]any) (xcache.MCache[K, V], error) {
	tp, _ := xmap.GetString(item, "Type")
	if tp == "" {
		return nil, fmt.Errorf("invalid cache type: %v", item)
	}
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
	case "Nop":
		return &xcache.Nop[K, V]{}, nil
	case "Chains":
		return cf.newChains[K, V](name, item)
	case "Wrap":
		return cf.newWrap[K, V](name, item)
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
	if codec, ok := xmap.GetString(item, "Codec"); ok {
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
	if codec, ok := xmap.GetString(item, "Codec"); ok {
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
		zc := chainConfigItem{}
		if err := xcodec.Convert(val, &zc); err != nil {
			errs = append(errs, err)
			return false
		}
		if zc.Ref == "" || zc.Life == 0 {
			errs = append(errs, fmt.Errorf("invalid config: Ref=%q Life=%s", zc.Ref, zc.Life.String()))
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
			Cache: c,
			LifeFn: func(ctx context.Context, key K, value V) time.Duration {
				return zc.Life.Duration()
			},
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
	ref, ok := xmap.GetString(item, "Ref")
	if !ok || ref == "" {
		return nil, errors.New("missing 'Ref'")
	}
	c, err := cf.Load[K, V](ref)
	if err != nil {
		return nil, err
	}

	if err := cf.refs.Add(name, ref); err != nil {
		return nil, err
	}

	ttlStr, _ := xmap.GetString(item, "Life")
	var ttl time.Duration
	if ttlStr != "" {
		ts, err := time.ParseDuration(ttlStr)
		if err != nil {
			return nil, err
		}
		ttl = ts
	}

	keyTransform, _ := xmap.GetMap(item, "KeyTransform")
	if len(keyTransform) == 0 && ttl == 0 {
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

	if ttl > 0 {
		wp.NewLifeFn = func(k K, v V, t time.Duration) time.Duration {
			return ttl
		}
	}
	return wp, nil
}

func (cf *ConfigFile) parserWrapKeyTransform[K comparable, V any](wp *xcache.Wrapper[K, V], cfg map[string]any) error {
	kt := reflect.TypeFor[K]()
	rule := kt.String()
	param, ok := xmap.GetMap(cfg, rule)
	if ok {
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
		param, ok := xmap.GetMap(cfg, "string")
		if ok {
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
	}
	child, ok := xmap.GetMap(cfg, "Default")
	if ok && len(child) > 0 {
		if child["Panic"] == true {
			err := fmt.Errorf("%w KeyTransform func for type %q, pls use xcache.RegisterKeyTransform first", xerror.NotFound, rule)
			panic(err)
		} else if child["Refuse"] == true {
			return fmt.Errorf("%w KeyTransform func for type %q, refused, pls use xcache.RegisterKeyTransform first", xerror.NotFound, rule)
		}
	}
	return nil
}

type chainConfigItem struct {
	Ref          string
	Life         xtype.Duration
	WriteTimeout xtype.Duration
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
