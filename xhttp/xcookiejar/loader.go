package xcookiejar

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http/cookiejar"
	"slices"
	"sync"

	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/xkvx"
	"github.com/xanygo/anygo/xmap"
)

var globalConfigFile *ConfigFile
var configErr error
var configOnce sync.Once

func loadConfig() {
	globalConfigFile = &ConfigFile{}
	err := xcfg.Parse("store/cookiejar", &globalConfigFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		configErr = err
	}
}

// Load 依据名字初始化并加载 xkv 对象。使用配置文件 {confDir}/store/xkv.{json|yml|toml}
func Load(name string) (*Jar, error) {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return nil, configErr
	}
	return globalConfigFile.Load(name)
}

func MustLoad(name string) *Jar {
	c, err := Load(name)
	if err != nil {
		panic(fmt.Errorf("load %q: %w", name, err))
	}
	return c
}

// CheckConfig 检查 store/xkv 配置文件是否正确
func CheckConfig() error {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return configErr
	}
	return globalConfigFile.Check()
}

// MustCheckConfig 检查 store/xkv 配置文件是否正确，若不正确则panic
func MustCheckConfig() {
	err := CheckConfig()
	if err != nil {
		panic(fmt.Errorf("check store/cookiejar: %w", err))
	}
}

type ConfigFile struct {
	Items    []map[string]any `json:"Items" yaml:"Items"`
	instance sync.Map
}

func (cf *ConfigFile) MustLoad(name string) *Jar {
	kv, err := cf.Load(name)
	if err != nil {
		panic(err)
	}
	return kv
}

type instanceValue struct {
	Value *Jar
	E     error
}

func (cf *ConfigFile) Load(name string) (*Jar, error) {
	if val, ok := cf.instance.Load(name); ok {
		v := val.(*instanceValue)
		if v.E != nil {
			return nil, v.E
		}
		return v.Value, nil
	}
	c, err := cf.createJar(name)
	old, loaded := cf.instance.LoadOrStore(name, &instanceValue{Value: c, E: err})
	if !loaded {
		return c, err
	}
	v := old.(*instanceValue)
	if v.E != nil {
		return nil, v.E
	}
	return v.Value, nil
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
		_, err := clone.newJar(item)
		if err != nil {
			err = fmt.Errorf("[%d]=%q: %w", idx, name, err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (cf *ConfigFile) createJar(name string) (*Jar, error) {
	for _, item := range cf.Items {
		str, _ := xmap.GetString(item, "Name")
		if name != str {
			continue
		}
		c, err := cf.newJar(item)
		if err != nil {
			err = fmt.Errorf("load xcookiejar %q: %w", name, err)
		}
		return c, err
	}
	return nil, fmt.Errorf("load xcookiejar %q: %w, Items.len=%d", name, xerror.NotFound, len(cf.Items))
}

func (cf *ConfigFile) newJar(item map[string]any) (*Jar, error) {
	tp, _ := xmap.GetString(item, "Type")
	if tp == "" {
		return nil, fmt.Errorf("invalid type: %v", item)
	}
	switch tp {
	case "No":
		return nil, nil
	case "Nop":
		return cf.newNop()
	case "Memory":
		return cf.newMemory(item)
	case "DB":
		return cf.newDB(item)
	case "KV":
		return cf.newKV(item)
	default:
		return nil, fmt.Errorf("unsupport type %q", tp)
	}
}

func (cf *ConfigFile) newDB(item map[string]any) (*Jar, error) {
	ps, err := cf.loadPSList(item)
	if err != nil {
		return nil, err
	}

	ds := &Database{}
	if err := ds.Init(item); err != nil {
		return nil, err
	}

	jar := &Jar{
		Storage: ds,
		PSList:  ps,
	}
	return jar, nil
}

func (cf *ConfigFile) loadPSList(item map[string]any) (cookiejar.PublicSuffixList, error) {
	name, err := xmap.GetString(item, "PSList")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, nil
	}
	return FindPsList(name)
}

func (cf *ConfigFile) newKV(item map[string]any) (*Jar, error) {
	ps, err := cf.loadPSList(item)
	if err != nil {
		return nil, err
	}
	ds := &KV{}
	if err := ds.Init(item); err != nil {
		return nil, err
	}
	jar := &Jar{
		Storage: ds,
		PSList:  ps,
	}
	return jar, nil
}

func (cf *ConfigFile) newMemory(item map[string]any) (*Jar, error) {
	ps, err := cf.loadPSList(item)
	if err != nil {
		return nil, err
	}
	memDB := xkvx.NewMemoryAny[Entry](xcodec.JSON)
	ds := &KV{
		EntryStore: func(key string) xkv.Hash[Entry] {
			return memDB.Hash(key)
		},
	}
	jar := &Jar{
		Storage: ds,
		PSList:  ps,
	}
	return jar, nil
}

func (cf *ConfigFile) newNop() (*Jar, error) {
	memDB := &xkvx.Nop[Entry]{}
	ds := &KV{
		EntryStore: func(key string) xkv.Hash[Entry] {
			return memDB.Hash(key)
		},
	}
	jar := &Jar{
		Storage: ds,
	}
	return jar, nil
}
