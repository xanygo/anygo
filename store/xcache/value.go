package xcache

import (
	"context"
	"time"
)

// Value 存储一个值的缓存对象
type Value[V any] interface {
	Set(ctx context.Context, value V, ttl time.Duration) error
	Get(ctx context.Context) (V, error)
	Delete(ctx context.Context) error
}

// NewValue 将 cache 转换为 Value Cache 类型
func NewValue[K comparable, V any](c Cache[K, V], key K) Value[V] {
	return &valueCache[K, V]{
		key:   key,
		cache: c,
	}
}

var _ Value[any] = (*valueCache[string, any])(nil)

type valueCache[K comparable, V any] struct {
	key   K
	cache Cache[K, V]
}

func (v *valueCache[K, V]) Set(ctx context.Context, value V, ttl time.Duration) error {
	return v.cache.Set(ctx, v.key, value, ttl)
}

func (v *valueCache[K, V]) Get(ctx context.Context) (V, error) {
	return v.cache.Get(ctx, v.key)
}

func (v *valueCache[K, V]) Delete(ctx context.Context) error {
	return v.cache.Delete(ctx, v.key)
}
