//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-22

package xi18n_test

import (
	"os"
	"testing"

	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xi18n"
	"github.com/xanygo/anygo/xt"
)

func TestLoadFS(t *testing.T) {
	b := &xi18n.Bundle{}
	err := xi18n.LoadFS(b, os.DirFS("testdata/data1"), ".", ".json", xcodec.JSON)
	xt.NoError(t, err)
	xt.SliceSortEqual(t, b.Languages(), []xi18n.Language{"zh", "en"})

	t.Run("zh", func(t *testing.T) {
		zh := b.Localize("zh")
		xt.NotNil(t, zh)
		msg1 := zh.Find("home/k1")
		xt.NotNil(t, msg1)
		xt.Equal(t, msg1.Other, "你好")
	})

	t.Run("en", func(t *testing.T) {
		en := b.Localize("en")
		xt.NotNil(t, en)
		msg1 := en.Find("home/k1")
		xt.NotNil(t, msg1)
		xt.Equal(t, msg1.Other, "hello")

		msg2 := en.Find("ns1/index/k1")
		xt.NotNil(t, msg2)
		xt.Equal(t, msg2.Other, "hello")

		msg3 := en.Find("ns1/z/z/z1")
		xt.NotNil(t, msg3)
		xt.Equal(t, msg3.Other, "hello world")
	})
}
