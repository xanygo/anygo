//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package encoder

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xtime"
)

func Test_autoNowTime(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var a time.Time
		got, ok, err := autoNowTime(dbtype.ColumnSchema{}, a)
		xt.NoError(t, err)
		xt.True(t, ok)
		xt.NotEmpty(t, got)
	})
	t.Run("case 2", func(t *testing.T) {
		var a xtime.DateInt
		got, ok, err := autoNowTime(dbtype.ColumnSchema{}, a)
		xt.NoError(t, err)
		xt.True(t, ok)
		xt.NotEmpty(t, got)
	})
}
