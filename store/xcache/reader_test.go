package xcache_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/xt"
)

func TestReader_Get(t *testing.T) {
	rd1 := &xcache.Reader[int, string]{
		New: func(ctx context.Context, key int) (string, error) {
			return strconv.Itoa(key), nil
		},
		Cache: xcache.NewLRU[int, xcache.ValueError[string]](1000),
		Life:  time.Hour,
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Go(func() {
			v, err := rd1.Get(t.Context(), i)
			xt.NoError(t, err)
			xt.Equal(t, v, strconv.Itoa(i))
		})
	}
	wg.Wait()

	v, err := rd1.Flush(t.Context(), 1)
	xt.NoError(t, err)
	xt.Equal(t, v, "1")
}
