//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-25

package xsession

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xlog"
)

var _ Storage = (*KVStore)(nil)

// KVStore 用 kv 存储 session 数据
type KVStore struct {
	DB  xkv.StringStorage // 必填，存储对象
	TTL time.Duration     // Session 有效期,可选，默认 30 天，若超过此时间没有读写，则清除掉

	DataKeyPrefix string // 存储 session 实际数据的 key 前缀，可选，默认为 "ss|"
	MetaKeyPrefix string // 存储元信息数据的 key 的前缀，可选，默认为 "session_meta"

	Logger xlog.Logger // 可选，执行清理逻辑时打印日志用

	gcRunning atomic.Bool
	gcTime    xsync.TimeStamp // 上次 GC 时间
}

func (ks *KVStore) getTTL() time.Duration {
	if ks.TTL > 0 {
		return ks.TTL
	}
	return 30 * 24 * time.Hour
}

func (ks *KVStore) getLogger() xlog.Logger {
	if ks.Logger == nil {
		return xlog.NopLogger{}
	}
	return ks.Logger
}

func (ks *KVStore) kvKey(id string) string {
	if ks.DataKeyPrefix == "" {
		return "ss|" + id
	}
	return ks.DataKeyPrefix + id
}

func (ks *KVStore) Get(ctx context.Context, id string) Session {
	ks.saveMeta(ctx, id)
	ss := &kvSession{
		id:    id,
		data:  ks.DB.Hash(ks.kvKey(id)),
		store: ks,
	}
	ks.autoGC()
	return ss
}

func (ks *KVStore) metaKey() string {
	if ks.MetaKeyPrefix == "" {
		return "session_meta"
	}
	return ks.MetaKeyPrefix
}

const (
	kvSuffixCreateTime = "_c" // 创建时间的 key 的 后缀
	kvSuffixVisitTime  = "_v" // 最后访问时间的 key 的 后缀
)

func (ks *KVStore) saveMeta(ctx context.Context, id string) error {
	keyPrefix := ks.metaKey()
	now := time.Now().Unix()

	createMetaDB := xkv.AsZSet[string](ks.DB, xcodec.JSON, keyPrefix+kvSuffixCreateTime)
	_, found, err := createMetaDB.ZScore(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		err = createMetaDB.ZAdd(ctx, float64(now), id)
	}
	if err != nil {
		return err
	}
	return xkv.AsZSet[string](ks.DB, xcodec.JSON, keyPrefix+kvSuffixVisitTime).ZAdd(ctx, float64(now), id)
}

func (ks *KVStore) autoGC() {
	if ks.gcRunning.Load() {
		return
	}
	const cycle = 5 * time.Minute
	next := ks.gcTime.Load().Add(cycle)
	if next.After(time.Now()) {
		return
	}
	if !ks.gcRunning.CompareAndSwap(false, true) {
		return
	}
	go safely.Run(ks.doGC)
}

// 清理过期的 Session 数据
func (ks *KVStore) doGC() {
	start := time.Now()
	ks.gcTime.Store(time.Now())
	defer ks.gcRunning.Store(false)

	logger := ks.getLogger()
	logger.Info(context.Background(), "Session: GC started")

	metaKeyPrefix := ks.metaKey()
	// 按照访问时间清理
	db := xkv.AsZSet[string](ks.DB, xcodec.JSON, metaKeyPrefix+kvSuffixVisitTime)

	expireTime := time.Now().Add(-1 * ks.getTTL())

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	var deleted int64
	db.ZRange(ctx, func(member string, score float64) bool {
		tm := time.Unix(int64(score), 0) // 最后更新时间
		if tm.Before(expireTime) {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
			defer cancel1()
			// 已经过期
			err := ks.delete(ctx1, member)
			logger.Info(ctx1, "clean_expired_session",
				xlog.String("member", member),
				xlog.Time("created", tm),
				xlog.ErrorAttr("resultErr", err),
			)
			deleted++
		}
		return true
	})

	logger.Info(ctx, "Session: GC completed",
		xlog.Int64("deleted", deleted),
		xlog.String("cost", time.Since(start).String()),
	)
}

func (ks *KVStore) createTime(ctx context.Context, id string) (time.Time, error) {
	keyPrefix := ks.metaKey()
	db := xkv.AsZSet[string](ks.DB, xcodec.JSON, keyPrefix+kvSuffixCreateTime)
	value, found, err := db.ZScore(ctx, id)
	if err != nil || !found {
		return time.Time{}, err
	}
	return time.Unix(int64(value), 0), nil
}

func (ks *KVStore) delete(ctx context.Context, id string) error {
	metaKeyPrefix := ks.metaKey()
	keys := []string{
		ks.kvKey(id),
		metaKeyPrefix + kvSuffixCreateTime,
		metaKeyPrefix + kvSuffixVisitTime,
	}
	var errs []error
	for _, key := range keys {
		if err := ks.DB.Delete(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

var _ Session = (*kvSession)(nil)

type kvSession struct {
	id    string
	data  xkv.Hash[string] // 用于存储 Session 的 kv 数据
	store *KVStore
}

func (kv *kvSession) ID() string {
	return kv.id
}

func (kv *kvSession) Set(ctx context.Context, key string, value string) error {
	return kv.data.HSet(ctx, key, value)
}

func (kv *kvSession) MSet(ctx context.Context, keyValues map[string]string) error {
	var errs []error
	for key, value := range keyValues {
		if err := kv.Set(ctx, key, value); err != nil {
		}

	}
	return errors.Join(errs...)
}

func (kv *kvSession) Get(ctx context.Context, key string) (string, error) {
	value, found, err := kv.data.HGet(ctx, key)
	if err != nil || !found {
		return "", err
	}
	return value, nil
}

func (kv *kvSession) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	var errs []error
	for _, key := range keys {
		value, _, err := kv.data.HGet(ctx, key)
		if err != nil {
			errs = append(errs, err)
		} else {
			result[key] = value
		}
	}
	return result, errors.Join(errs...)
}

func (kv *kvSession) Delete(ctx context.Context, keys ...string) error {
	var errs []error
	for _, key := range keys {
		if err := kv.data.HDel(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (kv *kvSession) Created(ctx context.Context) (time.Time, error) {
	return kv.store.createTime(ctx, kv.id)
}

func (kv *kvSession) Save(ctx context.Context) error {
	return nil
}

func (kv *kvSession) Clear(ctx context.Context) error {
	return kv.store.delete(ctx, kv.id)
}
