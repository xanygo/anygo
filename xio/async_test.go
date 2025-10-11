// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package xio

import (
	"bytes"
	"sync"
	"testing"

	"github.com/fsgo/fst"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mw := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     mw,
			ChanSize:   100,
			NeedStatus: true,
		}

		for i := 0; i < 1000; i++ {
			_, err := aw.Write([]byte("H"))
			fst.NoError(t, err)
		}
		fst.NoError(t, aw.Close())
		want := WriteStatus{
			Wrote: 1,
		}
		got := <-aw.WriteStatus()
		fst.Equal(t, want, got)
		fst.Equal(t, 1000, mw.Len())
	})

	t.Run("no write", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     b,
			ChanSize:   100,
			NeedStatus: true,
		}
		fst.NoError(t, aw.Close())
		_, ok := <-aw.WriteStatus()
		fst.False(t, ok)
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
			for i := 0; i < 1000; i++ {
				_, _ = aw.Write([]byte("abc"))
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_ = aw.Close()
			}
		}()
		wg.Wait()
	})
}
