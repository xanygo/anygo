//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
)

var (
	ErrClosed = errors.New("pool is closed")

	// ErrBadEntry 异常的节点，只有此错误，Entry 才会被从 pool 中移除
	ErrBadEntry = errors.New("bad pool entry")
)

type (
	Pool[V io.Closer] interface {
		// Get 获取一个，若有 idle 先使用 idle 的，若没有并且 Open 总个数在运行范围内，则创建一个新的，否则会一直等待
		//
		Get(ctx context.Context) (Entry[V], error)

		// GetIdle 可用于调试场景，查看 IDLE 状态的元素，当没有的时候会返回  nil,nil
		//
		// 特别注意：通过 Get 或 GetIdle 读取到的 Entry，都需要通过 Put 放回 Pool
		GetIdle(ctx context.Context) (Entry[V], error)

		// Put 将用过的对象放回 Pool，若 error 被判断为 Entry 对象不可用了，则将对象关闭，否则放回 idle 队列
		Put(e Entry[V], err error)

		// Close 关闭 Pool
		Close() error

		// Stats 返回 Pool 的统计状态
		Stats() Stats
	}

	Putter[V io.Closer] interface {
		Put(e Entry[V], err error)
	}
)

type Entry[V io.Closer] interface {
	ID() uint64
	Object() V
	CreatedAt() time.Time  // 创建时间
	LastUsedAt() time.Time // 上次使用时间
	UsageCount() uint64    // 使用次数

	// Release 放回连接池，
	// 若 err == ErrBadEntry,则将此元素从连接池里移除
	// 其他 err!=nil 认为是业务异常，照常放回连接池
	Release(err error)

	// ReleaseErr 上一次 Release 时候的 error
	ReleaseErr() error

	// Closer 关闭底层对象
	io.Closer
}

type Factory[V io.Closer] interface {
	New(ctx context.Context) (V, error)
}

// Validator 校验对象池里的对象是否有效
type Validator[V io.Closer] interface {
	// Validate 第二个参数 error 是，最后一次使用后，业务层返回的错误，可能是 nil
	Validate(V, error) error
}

// Option simple 配置选项，当前所有的选项都是可选的
type Option struct {
	// MaxOpen 最大打开数量
	// <= 0 为不限制
	MaxOpen int

	// MaxIdle 最大空闲数，应 <= MaxOpen
	// <=0 为不允许存在 Idle 元素
	MaxIdle int

	// MaxLifeTime 最大使用时长，超过后将被销毁
	// <=0 时使用默认值 30 分钟
	MaxLifeTime time.Duration

	// MaxIdleTime 最大空闲等待时间，超过后将被销毁
	// <=0 时使用默认值 5 分钟
	MaxIdleTime time.Duration

	// MaxPoolIdleTime GroupPool 使用，当超过此时长未被使用后，关闭并清理对应的 Pool
	// <=0 时使用默认值 5 minute
	MaxPoolIdleTime time.Duration
}

// Normalization 返回整理后的，有效值
func (opt *Option) Normalization() *Option {
	if opt == nil {
		return &Option{
			MaxLifeTime:     30 * time.Minute,
			MaxIdleTime:     5 * time.Minute,
			MaxPoolIdleTime: 5 * time.Minute,
		}
	}
	nv := &Option{
		MaxOpen:         opt.MaxOpen,
		MaxIdle:         opt.MaxIdle,
		MaxLifeTime:     opt.MaxLifeTime,
		MaxIdleTime:     opt.MaxIdleTime,
		MaxPoolIdleTime: opt.MaxPoolIdleTime,
	}
	if nv.MaxIdle > 0 && nv.MaxIdle > nv.MaxOpen {
		nv.MaxIdle = nv.MaxOpen
	}
	if nv.MaxLifeTime <= 0 {
		nv.MaxLifeTime = 30 * time.Minute
	}
	if nv.MaxIdleTime <= 0 {
		nv.MaxIdleTime = 5 * time.Minute
	}
	if nv.MaxPoolIdleTime <= 0 {
		nv.MaxPoolIdleTime = 5 * time.Minute
	}
	return nv
}

// Stats Pool's Stats
type Stats struct {
	Open bool // 连接池的状态，true-正常，false-已关闭

	MaxOpen     int           // 配置项：最大打开数
	MaxLifeTime time.Duration // 配置项：最大存活时长
	MaxIdleTime time.Duration // 配置项：最大空闲时长

	NumOpen int // 当前，已打开的总数
	InUse   int // 当前，正被使用的总数
	Idle    int // 当前，连接池里空闲的总数
	Wait    int // 当前，当前等待的总数

	// Counters
	WaitCount int64 // 累计等待的请求数

	WaitDuration      time.Duration // 累计等待的总时间
	MaxIdleClosed     int64         // 由于超过 MaxIdle, 被关闭的总数
	MaxIdleTimeClosed int64         // 由于超过 MaxIdleTime，被关闭的总数
	MaxLifeTimeClosed int64         // 由于超过 MaxLifetime，被关闭的总数
}

func (s Stats) Add(b Stats) Stats {
	return Stats{
		MaxOpen:     b.MaxOpen,
		MaxIdleTime: b.MaxIdleTime,
		MaxLifeTime: b.MaxLifeTime,

		Open:    s.Open || b.Open,
		NumOpen: s.NumOpen + b.NumOpen,
		InUse:   s.InUse + b.InUse,
		Idle:    s.Idle + b.Idle,
		Wait:    s.Wait + b.Wait,

		WaitCount: s.WaitCount + b.WaitCount,

		WaitDuration:      s.WaitDuration + b.WaitDuration,
		MaxIdleClosed:     s.MaxIdleClosed + b.MaxIdleClosed,
		MaxIdleTimeClosed: s.MaxIdleTimeClosed + b.MaxIdleTimeClosed,
		MaxLifeTimeClosed: s.MaxLifeTimeClosed + b.MaxLifeTimeClosed,
	}
}

func NewOpenEntry[V io.Closer](obj V, p Putter[V]) *OpenEntry[V] {
	entry := &OpenEntry[V]{
		obj:       obj,
		id:        globalID.Add(1),
		createdAt: time.Now(),
		pool:      p,
	}
	return entry
}

var _ Entry[io.Closer] = (*OpenEntry[io.Closer])(nil)

type OpenEntry[V io.Closer] struct {
	id         uint64
	obj        V
	createdAt  time.Time
	usedAt     xsync.TimeStamp
	pool       Putter[V]
	usageCount atomic.Uint64
	using      atomic.Bool
	releaseErr xsync.Value[error]
}

func (oe *OpenEntry[V]) ReleaseErr() error {
	return oe.releaseErr.Load()
}

func (oe *OpenEntry[V]) UpdateUsing() {
	oe.using.Store(true)
	oe.usedAt.Store(time.Now())
	oe.usageCount.Add(1)
}

func (oe *OpenEntry[V]) ID() uint64 {
	return oe.id
}

func (oe *OpenEntry[V]) Object() V {
	return oe.obj
}

func (oe *OpenEntry[V]) CreatedAt() time.Time {
	return oe.createdAt
}

func (oe *OpenEntry[V]) LastUsedAt() time.Time {
	return oe.usedAt.Load()
}

func (oe *OpenEntry[V]) UsageCount() uint64 {
	return oe.usageCount.Load()
}

func (oe *OpenEntry[V]) Release(err error) {
	if oe.using.CompareAndSwap(true, false) {
		oe.pool.Put(oe, err)
		oe.releaseErr.Store(err)
	}
}

func (oe *OpenEntry[V]) Close() error {
	err := oe.obj.Close()
	var emp V
	oe.obj = emp
	return err
}
