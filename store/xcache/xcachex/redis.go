//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-12

package xcachex

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xredis"
	"github.com/xanygo/anygo/store/xredis/resp3"
	"github.com/xanygo/anygo/xerror"
)

var _ xcache.StringCache = (*Redis)(nil)
var _ xcache.MCache[string, string] = (*Redis)(nil)

type Redis struct {
	KeyPrefix string
	Client    *xredis.Client

	readCnt   atomic.Uint64
	writeCnt  atomic.Uint64
	deleteCnt atomic.Uint64
	hitCnt    atomic.Uint64

	mux     sync.Mutex
	mSetSha string
}

func (r *Redis) Get(ctx context.Context, key string) (value string, err error) {
	r.readCnt.Add(1)
	value, err = r.Client.Get(ctx, r.KeyPrefix+key)
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
		keys[idx] = r.KeyPrefix + key
	}
	result, err = r.Client.MGet(ctx, keys...)
	r.hitCnt.Add(uint64(len(result)))
	return result, err
}

func (r *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	r.writeCnt.Add(1)
	return r.Client.Set(ctx, r.KeyPrefix+key, value, ttl)
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

func (r *Redis) MSet(ctx context.Context, data map[string]string, ttl time.Duration) error {
	r.writeCnt.Add(uint64(len(data)))
	tm := time.Now().Add(ttl)
	keys := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	for key, value := range data {
		keys = append(keys, r.KeyPrefix+key)
		values = append(values, value)
	}
	values = append(values, strconv.FormatInt(tm.UnixMilli(), 10))

	var result error
	for i := 0; i < 2; i++ {
		sha, err := r.loadScript(ctx)
		if err != nil {
			return err
		}
		ret, err := r.Client.EvalSha(ctx, sha, keys, values...)
		result = resp3.ToOkStatus(ret, err)
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
		keysNew[i] = r.KeyPrefix + key
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
