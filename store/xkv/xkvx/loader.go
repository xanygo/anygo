package xkvx

import (
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
)

var globalConfigFile *ConfigFile
var configErr error
var configOnce sync.Once

func loadConfig() {
	globalConfigFile = &ConfigFile{}
	configErr = xcfg.Parse("xkv", &globalConfigFile)
}

func Load[V any](name string) (xkv.Storage[V], error) {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return nil, configErr
	}
	return globalConfigFile.Load[V](name)
}

type instanceKey[V any] struct {
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
}

func (cf *ConfigFile) Load[V any](name string) (xkv.Storage[V], error) {
	key := instanceKey[V]{
		N: name,
	}
	if val, ok := cf.instance.Load(key); ok {
		v := val.(*instanceValue)
		if v.E != nil {
			return nil, v.E
		}
		return v.C.(xkv.Storage[V]), nil
	}
	c, err := cf.createKV[V](name)
	old, loaded := cf.instance.LoadOrStore(key, &instanceValue{C: c, E: err})
	if !loaded {
		return c, err
	}
	v := old.(*instanceValue)
	if v.E != nil {
		return nil, v.E
	}
	return v.C.(xkv.Storage[V]), nil
}

func (cf *ConfigFile) createKV[V any](name string) (xkv.Storage[V], error) {
	for _, item := range cf.Items {
		str, _ := xmap.GetString(item, "Name")
		if name != str {
			continue
		}
		c, err := cf.newKV[V](name, item)
		if err != nil {
			err = fmt.Errorf("xkv Name=%q: %w", name, err)
		}
		return c, err
	}
	return nil, fmt.Errorf("%w for xkv %q", xerror.NotFound, name)
}

func (cf *ConfigFile) newKV[V any](name string, item map[string]any) (xkv.Storage[V], error) {
	tp, _ := xmap.GetString(item, "Type")
	if tp == "" {
		return nil, fmt.Errorf("invalid xkv type: %v", item)
	}
	switch tp {
	case "File":
		return cf.newFile[V](name, item)
	case "Nop":
		return &xkv.Nop[V]{}, nil
	case "Memory":
		return xkv.NewMemoryAny[V](xcodec.JSON), nil
	case "Redis":
		return cf.newRedis[V](name, item)
	case "DB":
		return cf.newDB[V](name, item)
	default:
		return nil, fmt.Errorf("unsupport xkv type %q", tp)
	}
}

func (cf *ConfigFile) newFile[V any](name string, item map[string]any) (xkv.Storage[V], error) {
	fc := &xkv.File{}
	if err := fc.Init(item); err != nil {
		return nil, err
	}
	typeID := zreflect.TypeID1[V]()
	fc.Dir = filepath.Join(fc.Dir, strconv.FormatUint(uint64(typeID), 10))
	tr := &xkv.Transformer[V]{
		Storage: fc,
	}
	if err := tr.Init(item); err != nil {
		return nil, err
	}

	return tr, nil
}

func (cf *ConfigFile) newRedis[V any](name string, item map[string]any) (xkv.Storage[V], error) {
	rs := &Redis{}
	if err := rs.Init(item); err != nil {
		return nil, err
	}
	tr := &xkv.Transformer[V]{
		Storage: rs,
	}
	if err := tr.Init(item); err != nil {
		return nil, err
	}

	return tr, nil
}

func (cf *ConfigFile) newDB[V any](name string, item map[string]any) (xkv.Storage[V], error) {
	ds := &Database{
		ValueTypeID: zreflect.TypeID1[V](),
	}
	if err := ds.Init(item); err != nil {
		return nil, err
	}
	tr := &xkv.Transformer[V]{
		Storage: ds,
	}
	if err := tr.Init(item); err != nil {
		return nil, err
	}

	return tr, nil
}
