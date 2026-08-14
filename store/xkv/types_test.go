//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-24

package xkv_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/ut/xkvut"
	"github.com/xanygo/anygo/store/xkv"
	"github.com/xanygo/anygo/xt"
)

type testT struct {
	*testing.T
}

func (t testT) Run(name string, fn func(tb xt.TB)) {
	t.T.Run(name, func(tt *testing.T) {
		fn(testT{T: tt})
	})
}

func testStringStorage(t *testing.T, ff xkv.StringStorage) {
	tb := testT{T: t}
	tb.Run("t1", func(tb xt.TB) {
		xkvut.TestStringStorage1(tb, ff)
	})

	tb.Run("t2", func(tb xt.TB) {
		xkvut.TestStringStorage2(tb, ff)
	})
}

func benchStorage(b *testing.B, st xkv.StringStorage) {
	b.Run("string", func(b *testing.B) {
		s1 := st.String("str1")
		b.Run("set", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = s1.Set(b.Context(), "v1")
			}
		})
		b.Run("get", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s1.Get(b.Context())
			}
		})
	})

	b.Run("list", func(b *testing.B) {
		l1 := st.List("list1")
		for i := 0; i < b.N; i++ {
			_, err1 := l1.LPush(b.Context(), "v1")
			if err1 != nil {
				b.Fatal(err1.Error())
			}
			got, found, err2 := l1.LPop(b.Context())
			if err2 != nil {
				b.Fatal(err2.Error())
			}
			if !found || got != "v1" {
				b.Fatalf("not found or value is wrong %v %v", found, got)
			}
		}
	})
}
