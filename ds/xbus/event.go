//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-06

package xbus

import (
	"sync"

	"github.com/xanygo/anygo/safely"
)

type EventBus[T any] struct {
	subs []chan T
	mu   sync.RWMutex
}

func (e *EventBus[T]) Publish(err T) {
	e.mu.RLock()
	total := len(e.subs)
	subs := e.subs
	e.mu.RUnlock()

	var wg sync.WaitGroup
	for i := range total {
		sub := subs[i]
		wg.Go(safely.WrapVoid(func() {
			select {
			case sub <- err:
			default:
			}
		}))
	}
	wg.Wait()
}

func (e *EventBus[T]) Subscribe() <-chan T {
	ch := make(chan T, 1)
	e.mu.Lock()
	e.subs = append(e.subs, ch)
	e.mu.Unlock()
	return ch
}

func (e *EventBus[T]) Subscribed() bool {
	e.mu.RLock()
	total := len(e.subs)
	e.mu.RUnlock()
	return total > 0
}

func (e *EventBus[T]) Unsubscribe(ch <-chan T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, sub := range e.subs {
		if sub == ch {
			close(sub)
			e.subs = append(e.subs[:i], e.subs[i+1:]...)
			break
		}
	}
}

func (e *EventBus[T]) Stop() {
	e.Close()
}

func (e *EventBus[T]) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, sub := range e.subs {
		close(sub)
	}
	e.subs = nil
}
