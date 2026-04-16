//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-16

package zmatcher_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zstr/zmatcher"
	"github.com/xanygo/anygo/xt"
)

func TestCompile(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		matchFn, err := zmatcher.Compile("hello")
		xt.NoError(t, err)
		xt.True(t, matchFn("hello"))
		xt.False(t, matchFn("Hello"))
	})
	t.Run("case 2", func(t *testing.T) {
		matchFn, err := zmatcher.Compile("wc:*hello")
		xt.NoError(t, err)
		xt.True(t, matchFn("hello"))
		xt.True(t, matchFn("-hello"))
		xt.True(t, matchFn("ABChello"))

		xt.False(t, matchFn("helloW"))
	})
	t.Run("case 3", func(t *testing.T) {
		matchFn, err := zmatcher.Compile("wc:*hello*")
		xt.NoError(t, err)
		xt.True(t, matchFn("hello"))
		xt.True(t, matchFn("-hello"))
		xt.True(t, matchFn("ABChello"))

		xt.True(t, matchFn("helloW"))
		xt.True(t, matchFn("AhelloW"))

		xt.False(t, matchFn("abc"))
	})
	t.Run("case 4", func(t *testing.T) {
		matchFn, err := zmatcher.Compile("wc:hello*")
		xt.NoError(t, err)
		xt.True(t, matchFn("hello"))
		xt.False(t, matchFn("-hello"))
		xt.False(t, matchFn("ABChello"))

		xt.True(t, matchFn("helloW"))
		xt.False(t, matchFn("AhelloW"))

		xt.False(t, matchFn("abc"))
	})

	t.Run("case 5", func(t *testing.T) {
		matchFn, err := zmatcher.Compile("re:^hello$")
		xt.NoError(t, err)
		xt.True(t, matchFn("hello"))
		xt.False(t, matchFn("-hello"))
		xt.False(t, matchFn("ABChello"))

		xt.False(t, matchFn("helloW"))
		xt.False(t, matchFn("AhelloW"))

		xt.False(t, matchFn("abc"))
	})
}
