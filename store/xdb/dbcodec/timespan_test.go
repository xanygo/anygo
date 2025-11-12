//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package dbcodec

import (
	"strconv"
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestTimeSpan_Encode(t *testing.T) {
	de := TimeSpan{}
	t.Run("time", func(t *testing.T) {
		tm, err := time.Parse(time.DateTime, time.DateTime)
		xt.NoError(t, err)
		sec := tm.Unix()

		got, err := de.Encode(tm)
		xt.NoError(t, err)
		xt.Equal(t, any(sec), got)
	})
	t.Run("not-time", func(t *testing.T) {
		got, err := de.Encode("string")
		xt.Error(t, err)
		xt.Empty(t, got)
	})
}

func TestTimeSpan_Decode(t *testing.T) {
	de := TimeSpan{}
	t.Run("time", func(t *testing.T) {
		bt, err := time.Parse(time.DateTime, time.DateTime)
		xt.NoError(t, err)
		sec := bt.Unix()

		var tm time.Time
		err = de.Decode(strconv.FormatInt(sec, 10), &tm)
		xt.NoError(t, err)
		xt.Equal(t, 2006, tm.Year())

		err = de.Decode("hello", &tm)
		xt.Error(t, err)
	})

	t.Run("not-time", func(t *testing.T) {
		var tm string
		err := de.Decode("1234567890", &tm)
		xt.Error(t, err)
		xt.Empty(t, tm)

		err = de.Decode("hello", &tm)
		xt.Error(t, err)
	})
}
