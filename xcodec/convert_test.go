//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-13

package xcodec_test

import (
	"testing"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xt"
)

func TestConvert(t *testing.T) {
	data := map[string]any{
		"Name": "hello",
		"Age":  18,
	}
	var u1 user
	xt.NoError(t, xcodec.Convert(data, &u1))
	xt.Equal(t, u1, user{Name: "hello", Age: 18})
}

func TestConvertAs(t *testing.T) {
	data := map[string]any{
		"Name": "hello",
		"Age":  18,
	}

	u2, err := xcodec.ConvertAs[user](data)
	xt.NoError(t, err)
	xt.Equal(t, u2, user{Name: "hello", Age: 18})

	u3, err := xcodec.ConvertAs[*user](data)
	xt.NoError(t, err)
	xt.Equal(t, u3, &user{Name: "hello", Age: 18})
}

type user struct {
	Name string
	Age  int
}
