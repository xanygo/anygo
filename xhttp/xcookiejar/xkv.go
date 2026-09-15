package xcookiejar

import (
	"context"
	"strconv"
	"time"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xsync"
)

var _ Storage = (*KV)(nil)

// KV 使用 xkv 作为存储，存储 cookie 消息的实现
type KV struct {
	// EntryStore 必填，存储 cookie 实体数据的
	EntryStore func(key string) xkv.Hash[Entry]

	// MetaStore 可选，存储最后访问时间
	// 若不为空，则每次读写后，会存储最后访问时间，并且在后台定期自动扫描并清理过期数据
	MetaStore func() xkv.ZSet[string]

	compactTime xsync.Interval // 存储上一次清理的时间
}

// Get implements [Storage].
func (k *KV) Get(ctx context.Context, key string) ([]Entry, error) {
	defer k.autoCompact()
	result, err := k.doGet(ctx, key, false)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 && k.MetaStore != nil {
		k.MetaStore().ZAdd(ctx, float64(time.Now().Unix()), key)
	}
	return result, nil
}

func (k *KV) doGet(ctx context.Context, key string, byCompact bool) ([]Entry, error) {
	hs := k.EntryStore(key)
	values, err := hs.HGetAll(ctx)
	if (len(values) == 0 && !byCompact) || err != nil {
		return nil, err
	}

	result := make([]Entry, 0, len(values))
	var expired []string
	for field, item := range values {
		if item.Expires.Before(time.Now()) {
			expired = append(expired, field)
		} else {
			result = append(result, item)
		}
	}
	if len(expired) > 0 {
		hs.HDel(ctx, expired...)
	}

	if len(result) == 0 && k.MetaStore != nil {
		_ = k.MetaStore().ZRem(ctx, key)
	}
	return result, nil
}

// Set implements [Storage].
func (k *KV) Set(ctx context.Context, key string, items []Entry) error {
	defer k.autoCompact()

	if len(items) == 0 {
		return nil
	}
	if k.MetaStore != nil {
		err := k.MetaStore().ZAdd(ctx, float64(time.Now().Unix()), key)
		if err != nil {
			return err
		}
	}
	seqNum := time.Now().UnixMilli()

	values := make(map[string]Entry, len(items))
	for index, item := range items {
		key := hash(item.ID())
		item.SeqNum = seqNum + int64(index) // 以此递增的
		values[string(key[:])] = item
	}
	return k.EntryStore(key).HMSet(ctx, values)
}

// DeleteEntry implements [Storage].
func (k *KV) DeleteEntry(ctx context.Context, key string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	hs := k.EntryStore(key)
	fields := make([]string, 0, len(ids))
	for _, id := range ids {
		h := hash(id)
		fields = append(fields, string(h[:]))
	}
	err := hs.HDel(ctx, fields...)
	k.doGet(ctx, key, true) // 检查过期数据
	return err
}

func (k *KV) autoCompact() {
	if k.MetaStore == nil {
		return
	}
	if !k.compactTime.Allow(10 * time.Minute) {
		return
	}
	go safely.Run(k.compact)
}

func (k *KV) compact() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ms := k.MetaStore()
	max := time.Now().AddDate(0, 0, -7).Unix()
	// 只扫描 7天前没有访问过的
	ms.ZRangeByScore(ctx, "-inf", strconv.FormatInt(max, 10), func(member string, score float64) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			// pass
		}
		// 读取并检查一遍
		k.doGet(ctx, member, true)
		return true
	})
}
