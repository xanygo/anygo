//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"encoding/xml"

	"github.com/xanygo/anygo/xcfg/internal/parser"
	"github.com/xanygo/anygo/xcodec"
)

type parserNameFn struct {
	Fn   xcodec.Decoder
	Name string
}

// defaultDecoders 所有默认的解析器，
// 当传入配置文件名不包含后置的时候，会使用此顺序依次查找
var defaultDecoders = []parserNameFn{
	{Name: ".json", Fn: xcodec.DecodeFunc(parser.JSON)},
	{Name: ".xml", Fn: xcodec.DecodeFunc(xml.Unmarshal)},
}
