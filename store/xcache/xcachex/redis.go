//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xcachex

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/xerror"
)

var _ xcache.StringCache = (*Redis)(nil)
var _ xcache.MCache[string, string] = (*Redis)(nil)

type Redis struct {
	KeyPrefix string // 缓存 key 的前缀，可选
	Client    *xredis.Client

	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
	hitCnt    atomic.Uint64

	mux     sync.Mutex
	mSetSha string
}

func (r *Redis) Init(param map[string]any) error {
	if r.KeyPrefix == "" {
		r.KeyPrefix, _ = xmap.GetString(param, "KeyPrefix")
	}
	if r.Client == nil {
		service, ok := xmap.GetString(param, "Service")
		if !ok || service == "" {
			return fmt.Errorf("invalid Service in %v", param)
		}
		r.Client = xredis.NewClient(service)
	}
	return nil
}

func (r *Redis) Has(ctx context.Context, key string) (bool, error) {
	num, err := r.Client.EXISTS(ctx, r.KeyPrefix+key)
	return num == 1, err
}

const ttlNoExpire = 31 * 24 * time.Hour

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.Client.TTL(ctx, r.KeyPrefix+key)
	if err != nil {
		if errors.Is(err, xredis.ErrNil) {
			return 0, nil
		}
		return 0, err
	}
	if ttl == -1 { // 理论应该不存在
		return ttlNoExpire, nil
	}
	return ttl, nil
}

func (r *Redis) Expire(ctx context.Context, key string, life time.Duration) error {
	ok, err := r.Client.ExpireOpt(ctx, r.KeyPrefix+key, life, "XX")
	if ok {
		return nil
	}
	return err
}

func (r *Redis) fullKey(key string) string {
	return r.KeyPrefix + key
}

func (r *Redis) Get(ctx context.Context, key string) (value string, err error) {
	r.readCnt.Add(1)
	value, err = r.Client.Get(ctx, r.fullKey(key))
	if err == nil {
		r.hitCnt.Add(1)
		return value, nil
	}
	if errors.Is(err, xredis.ErrNil) {
		return value, xerror.NotFound
	}
	return value, err
}

func (r *Redis) MGet(ctx context.Context, keys ...string) (result map[string]string, err error) {
	r.readCnt.Add(uint64(len(keys)))
	for idx, key := range keys {
		keys[idx] = r.fullKey(key)
	}
	result, err = r.Client.MGet(ctx, keys...)
	r.hitCnt.Add(uint64(len(result)))
	return result, err
}

func (r *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	r.writeCnt.Add(1)
	return r.Client.SetWithTTL(ctx, r.fullKey(key), value, ttl)
}

const mSetScript = `
for i = 1, #KEYS do
    redis.call('SET', KEYS[i], ARGV[i], 'PXAT', ARGV[#ARGV])
end
return 'OK'
`

func (r *Redis) loadScript(ctx context.Context) (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.mSetSha != "" {
		return r.mSetSha, nil
	}
	ret, err := r.Client.ScriptLoad(ctx, mSetScript)
	r.mSetSha = ret
	return ret, err
}

// MSet 使用 lua 脚本实现的批量 Set 功能
func (r *Redis) MSet(ctx context.Context, data map[string]string, ttl time.Duration) error {
	r.writeCnt.Add(uint64(len(data)))
	tm := time.Now().Add(ttl)
	keys := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	for key, value := range data {
		keys = append(keys, r.fullKey(key))
		values = append(values, value)
	}
	values = append(values, strconv.FormatInt(tm.UnixMilli(), 10))

	var result error
	for range 2 {
		sha, err := r.loadScript(ctx)
		if err != nil {
			return err
		}
		ret := r.Client.EvalSha(ctx, sha, keys, values...)
		result = ret.OKStatus()
		// 若遇到 NOSCRIPT 错误则重新执行一次
		if result != nil && strings.Contains(result.Error(), "NOSCRIPT") {
			r.mux.Lock()
			r.mSetSha = ""
			r.mux.Unlock()
			continue
		}
		return result
	}
	return result
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	r.deleteCnt.Add(uint64(len(keys)))
	keysNew := make([]string, len(keys))
	for i, key := range keys {
		keysNew[i] = r.fullKey(key)
	}
	_, err := r.Client.Del(ctx, keysNew...)
	return err
}

func (r *Redis) Stats() xcache.Stats {
	return xcache.Stats{
		Read:   r.readCnt.Load(),
		Write:  r.writeCnt.Load(),
		Delete: r.deleteCnt.Load(),
		Hit:    r.hitCnt.Load(),
	}
}
