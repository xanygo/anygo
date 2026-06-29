package xpp

import (
	"sync"
	"time"

	"github.com/xanygo/anygo/safely"
)

// After  和 time.AfterFunc 功能类似，但是支持多次设置延迟认任务，若旧任务还未运行则会自动将其取消
type After struct {
	next *time.Timer
	mu   sync.Mutex
}

// Schedule 延迟 delay 时长后，执行回调函数 fn 。
// 若多次调用，只会保留最后一个
func (a *After) Schedule(delay time.Duration, fn func()) {
	if delay < 0 {
		delay = 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.next != nil {
		a.next.Stop()
	}
	a.next = time.AfterFunc(delay, func() {
		safely.RunVoid(fn)

		a.mu.Lock()
		a.next = nil
		a.mu.Unlock()
	})
}

func (a *After) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.next != nil {
		a.next.Stop()
		a.next = nil
	}
}
