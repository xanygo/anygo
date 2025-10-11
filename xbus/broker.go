//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbus

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/xanygo/anygo/safely"
)

// NewBroker 创建 Broker
func NewBroker() *Broker {
	return &Broker{
		consumers: make(map[Topic][]Consumer),
	}
}

// Broker 负责管理 Producer/Consumer，并进行派发
type Broker struct {
	mu        sync.RWMutex
	consumers map[Topic][]Consumer
	preProds  []Producer // Start 前注册的 producers
	started   bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup // 等待所有内部 goroutine
	errChan   EventBus[error]
}

// RegisterConsumer 动态注册消费者（Start 前后均可调用）
func (b *Broker) RegisterConsumer(topic Topic, c Consumer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consumers[topic] = append(b.consumers[topic], c)
}

// MustRegisterConsumer 将 c 转换为 Consumer 并注册，若不是 Consumer 则 panic
func (b *Broker) MustRegisterConsumer(topic Topic, c any) {
	c1, ok := c.(Consumer)
	if !ok {
		panic(fmt.Errorf("%T is not a Consumer", c))
	}
	b.RegisterConsumer(topic, c1)
}

// RegisterProducer 动态注册生产者：
// - 若未 Start，先暂存；
// - 若已 Start，立刻接入并开始消费其消息流。
func (b *Broker) RegisterProducer(p Producer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		b.preProds = append(b.preProds, p)
		return
	}
	b.addProducerLocked(p)
}

// MustRegisterProducer 将 p 转换为 Producer 并注册，若不是 Producer 则 panic
func (b *Broker) MustRegisterProducer(p any) {
	p1, ok := p.(Producer)
	if !ok {
		panic(fmt.Errorf("%T is not a Producer", p))
	}
	b.RegisterProducer(p1)
}

// Start 启动 Broker（幂等，多次调用仅第一次生效）
func (b *Broker) Start() {
	b.mu.Lock()
	if b.started {
		b.mu.Unlock()
		return
	}
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.started = true

	// 启动已注册的 producers
	for _, p := range b.preProds {
		b.addProducerLocked(p)
	}
	// 释放 preProds，后续不再使用
	b.preProds = nil
	b.mu.Unlock()
}

// Stop 停止 Broker，等待内部 goroutine 退出
func (b *Broker) Stop() {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return
	}
	b.cancel()
	b.started = false
	b.mu.Unlock()

	b.wg.Wait()
}

// 必须在持有 b.mu 的情况下调用
func (b *Broker) addProducerLocked(p Producer) {
	ch := p.Messages()
	b.wg.Go(func() {
		for {
			select {
			case <-b.ctx.Done():
				return
			case m, ok := <-ch:
				if !ok {
					return
				}
				b.dispatch(m)
			}
		}
	})
}

func (b *Broker) dispatch(m Message) {
	b.mu.RLock()
	cs := slices.Clone(b.consumers[m.Topic])
	if m.Topic.id != 0 {
		cs = append(cs, b.consumers[AnyTopic]...)
	}
	b.mu.RUnlock()

	for _, c := range cs {
		select {
		case <-b.ctx.Done():
			return
		default:
		}
		c := c
		go safely.RunVoid(func() {
			if err := c.Consume(b.ctx, m); err != nil {
				b.errChan.Publish(fmt.Errorf("%s(%s): %w", Name(c), m.Topic.String(), err))
			}
		})
	}
}

func (b *Broker) WatchError() <-chan error {
	return b.errChan.Subscribe()
}
