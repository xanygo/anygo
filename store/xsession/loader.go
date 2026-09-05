package xsession

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/store/xcache/xcachex"
	"github.com/xanygo/anygo/store/xkv/xkvx"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xerror"
)

var globalConfigFile *ConfigFile
var configErr error
var configOnce sync.Once

func loadConfig() {
	globalConfigFile = &ConfigFile{}
	configErr = xcfg.Parse("store/xsession", &globalConfigFile)

	log.Println("configErr=", configErr)
}

// LoadStorageFunc 依据名字初始化并加载 HTTPHandler 所需要的 NewStorageFunc。
//
// 使用配置文件 {confDir}/store/xsession.{json|yml|toml}
func LoadStorageFunc(name string) (NewStorageFunc, error) {
	configOnce.Do(loadConfig)
	if configErr != nil {
		return nil, configErr
	}
	return globalConfigFile.Load(name)
}

func MustLoadStorageFunc(name string) NewStorageFunc {
	c, err := LoadStorageFunc(name)
	if err != nil {
		panic(err)
	}
	return c
}

type ConfigFile struct {
	Items    []map[string]any `json:"Items" yaml:"Items"`
	instance sync.Map
}

type instanceValue struct {
	V NewStorageFunc
	E error
}

func (cf *ConfigFile) Load(name string) (NewStorageFunc, error) {
	if val, ok := cf.instance.Load(name); ok {
		v := val.(*instanceValue)
		if v.E != nil {
			return nil, v.E
		}
		return v.V, nil
	}
	fn, err := cf.createFn(name)
	old, loaded := cf.instance.LoadOrStore(name, &instanceValue{V: fn, E: err})
	if !loaded {
		return fn, err
	}
	v := old.(*instanceValue)
	return v.V, v.E
}

func (cf *ConfigFile) MustLoad(name string) NewStorageFunc {
	fn, err := cf.Load(name)
	if err != nil {
		panic(err)
	}
	return fn
}

func (cf *ConfigFile) createFn(name string) (NewStorageFunc, error) {
	for _, item := range cf.Items {
		str, _ := xmap.GetString(item, "Name")
		if name != str {
			continue
		}
		c, err := cf.newFn(name, item)
		if err != nil {
			err = fmt.Errorf("load xsession %q: %w", name, err)
		}
		return c, err
	}
	return nil, fmt.Errorf("load xsession %q: %w, Items.len=%d", name, xerror.NotFound, len(cf.Items))
}

func (cf *ConfigFile) newFn(name string, item map[string]any) (NewStorageFunc, error) {
	tp, _ := xmap.GetString(item, "Type")
	if tp == "" {
		return nil, fmt.Errorf("invalid Store type: %v", item)
	}
	switch tp {
	case "XKV":
		return cf.newXKV(name, item)
	case "Cookie":
		return cf.newCookie(name, item)
	case "XCache":
		return cf.newXCache(name, item)
	default:
		return nil, fmt.Errorf("unsupport session story type %q", tp)
	}
}

func (cf *ConfigFile) getSessionLife(item map[string]any) (time.Duration, error) {
	life, ok := xmap.GetString(item, "Life")
	if !ok || life == "" {
		return 365 * 24 * time.Hour, nil
	}
	return time.ParseDuration(life)
}

func (cf *ConfigFile) newXKV(name string, item map[string]any) (NewStorageFunc, error) {
	life, err := cf.getSessionLife(item)
	if err != nil {
		return nil, err
	}
	ref, ok := xmap.GetString(item, "Ref")
	if !ok || ref == "" {
		return nil, fmt.Errorf("missing 'Ref' in %v", item)
	}
	db, err := xkvx.Load[string](ref)
	if err != nil {
		return nil, err
	}
	store := &KVStore{
		DB:   db,
		Life: life,
	}
	store.DataKeyPrefix, _ = xmap.GetString(item, "DataKeyPrefix")
	store.MetaKeyPrefix, _ = xmap.GetString(item, "MetaKeyPrefix")

	fn := func(writer http.ResponseWriter, request *http.Request) Storage {
		return store
	}
	return fn, nil
}

func (cf *ConfigFile) newCookie(name string, item map[string]any) (NewStorageFunc, error) {
	life, err := cf.getSessionLife(item)
	if err != nil {
		return nil, err
	}

	cipherType, ok := xmap.GetString(item, "CipherType")
	if !ok || cipherType == "" {
		cipherType = "AesOFB"
	}

	cipherKey, _ := xmap.GetString(item, "CipherKey")
	if cipherKey == "" {
		return nil, fmt.Errorf("missing 'CipherKey' in %v", item)
	}
	cipherIV, _ := xmap.GetString(item, "CipherIV")

	var cipher xcodec.Cipher
	switch cipherType {
	case "AesOFB":
		cipher = &xcodec.AesOFB{
			Key: cipherKey,
			IV:  cipherIV,
		}
	case "AesGCM":
		cipher = &xcodec.AesGCM{
			Key: cipherKey,
		}
	case "AesBlock":
		cipher = &xcodec.AesBlock{
			Key: cipherKey,
			IV:  cipherIV,
		}
	default:
		return nil, fmt.Errorf("unsupport CipherType %q", cipherType)
	}

	cookieName, _ := xmap.GetString(item, "CookieName")
	cookiePath, _ := xmap.GetString(item, "CookiePath")

	fn := func(writer http.ResponseWriter, request *http.Request) Storage {
		return &CookieStore{
			Writer:     writer,
			Request:    request,
			Cipher:     cipher,
			CookieName: cookieName,
			CookiePath: cookiePath,
			Life:       life,
		}
	}
	return fn, nil
}

func (cf *ConfigFile) newXCache(name string, item map[string]any) (NewStorageFunc, error) {
	life, err := cf.getSessionLife(item)
	if err != nil {
		return nil, err
	}
	ref, ok := xmap.GetString(item, "Ref")
	if !ok || ref == "" {
		return nil, fmt.Errorf("missing 'Ref' in %v", item)
	}
	ch, err := xcachex.Load[string, string](ref)
	if err != nil {
		return nil, err
	}
	store := &CacheStore{
		Cache: ch,
		Life:  life,
	}
	fn := func(writer http.ResponseWriter, request *http.Request) Storage {
		return store
	}
	return fn, nil
}
