//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xerror

import (
	"errors"
	"io"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestOnceErr(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var oe *OnceSet
		xt.Equal(t, "<nil>", oe.Error())
		xt.Nil(t, oe.Unwrap())
	})
	t.Run("case 2", func(t *testing.T) {
		var oe OnceSet
		xt.Equal(t, "<nil>", oe.Error())
		xt.Nil(t, oe.Unwrap())
		err1 := errors.New("hello")
		xt.True(t, oe.SetOnce(err1))
		xt.False(t, oe.SetOnce(io.EOF))
		xt.ErrorIs(t, &oe, err1)
	})
}
