//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-12

package dbcodec

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/xt"
)

func TestDateTime_Encode(t *testing.T) {
	de := DateTime{}
	t.Run("time", func(t *testing.T) {
		tm, err := time.Parse(time.DateTime, time.DateTime)
		xt.NoError(t, err)
		got, err := de.Encode(tm)
		xt.NoError(t, err)
		xt.Equal(t, time.DateTime, got)
	})
	t.Run("not-time", func(t *testing.T) {
		got, err := de.Encode("string")
		xt.Error(t, err)
		xt.Empty(t, got)
	})
}

func TestDateTime_Decode(t *testing.T) {
	de := DateTime{}
	t.Run("time", func(t *testing.T) {
		var tm time.Time
		err := de.Decode(time.DateTime, &tm)
		xt.NoError(t, err)
		xt.Equal(t, 2006, tm.Year())

		err = de.Decode("hello", &tm)
		xt.Error(t, err)
	})

	t.Run("not-time", func(t *testing.T) {
		var tm string
		err := de.Decode(time.DateTime, &tm)
		xt.Error(t, err)
		xt.Empty(t, tm)

		err = de.Decode("hello", &tm)
		xt.Error(t, err)
	})
}
