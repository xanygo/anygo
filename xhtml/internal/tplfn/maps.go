//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-06

package tplfn

import (
	"html/template"
	"math/rand/v2"

	"encoding/json"
	"fmt"
	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xstr"
)

var Funcs = template.FuncMap{
	"xInputChecked":   Checked,
	"xOptionSelected": OptionSelected,

	"xRandStr": func() string {
		return xstr.RandNChar(8)
	},
	"xRandStrN": xstr.RandNChar,

	"xRandUint":   rand.Uint,
	"xRandUint32": rand.Uint32,
	"xRandUint64": rand.Uint64,

	"xRandUintN":   rand.UintN,
	"xRandUint32N": rand.Uint32N,
	"xRandUint64N": rand.Uint64N,

	"xRandInt":   rand.Int,
	"xRandInt32": rand.Int32,
	"xRandInt64": rand.Int64,

	"xRandIntN":   rand.IntN,
	"xRandInt32N": rand.Int32N,
	"xRandInt64N": rand.Int64N,

	"xRandFloat64": rand.Float64,
	"xRandFloat32": rand.Float32,

	"xNewMap": xmap.Creat,

	"xDateTime":   DateTime,
	"xEachOfIter": EachOfIter,
	"xRandOfIter": RandOfIter,

	"xJSON": func(val any) string {
		bf, err := json.MarshalIndent(val, " ", "  ")
		if err != nil {
			return err.Error()
		}
		return string(bf)
	},

	"xDump": func(value any) string {
		return fmt.Sprintf("%#v", value)
	},
}
