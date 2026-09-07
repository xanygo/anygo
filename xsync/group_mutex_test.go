package xsync

import (
	"sync"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestGroupMutex_Do(t *testing.T) {
	g := &GroupMutex[any]{}
	var num int
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			g.Do("num", func() {
				num++
			})
		})
	}
	wg.Wait()
	xt.Equal(t, num, 100)
	xt.Empty(t, g.items)
}
