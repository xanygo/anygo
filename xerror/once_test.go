//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xerror

import (
	"errors"
	"io"
	"testing"

	"github.com/fsgo/fst"
)

func TestOnceErr(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var oe *OnceSet
		fst.Equal(t, "<nil>", oe.Error())
		fst.Nil(t, oe.Unwrap())
	})
	t.Run("case 2", func(t *testing.T) {
		var oe OnceSet
		fst.Equal(t, "<nil>", oe.Error())
		fst.Nil(t, oe.Unwrap())
		err1 := errors.New("hello")
		fst.True(t, oe.SetOnce(err1))
		fst.False(t, oe.SetOnce(io.EOF))
		fst.ErrorIs(t, &oe, err1)
	})
}
