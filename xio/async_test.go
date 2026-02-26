// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package xio

import (
	"bytes"
	"sync"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mw := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     mw,
			ChanSize:   100,
			NeedStatus: true,
		}

		for range 1000 {
			_, err := aw.Write([]byte("H"))
			xt.NoError(t, err)
		}
		xt.NoError(t, aw.Close())
		want := WriteStatus{
			Wrote: 1,
		}
		got := <-aw.WriteStatus()
		xt.Equal(t, want, got)
		xt.Equal(t, 1000, mw.Len())
	})

	t.Run("no write", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     b,
			ChanSize:   100,
			NeedStatus: true,
		}
		xt.NoError(t, aw.Close())
		_, ok := <-aw.WriteStatus()
		xt.False(t, ok)
	})

	t.Run("with gor", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:   b,
			ChanSize: 100,
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 1000 {
				_, _ = aw.Write([]byte("abc"))
			}
		}()
		go func() {
			defer wg.Done()
			for range 1000 {
				_ = aw.Close()
			}
		}()
		wg.Wait()
	})
}
