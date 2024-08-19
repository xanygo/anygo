//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-19

package xerrors

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestOnceErr(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var oe *OnceSet
		fst.Equal(t, "<nil>", oe.Error())
		fst.Nil(t, oe.Unwrap())
	})
}
