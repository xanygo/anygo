//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"github.com/xanygo/anygo/xcodec"
)

const TextName = "text"

var _ Codec = (*Text)(nil)

type Text struct{}

func (t Text) Kind() Kind {
	return KindString
}

func (t Text) Name() string {
	return TextName
}

func (t Text) Encode(obj any) (any, error) {
	return xcodec.EncodeToString(xcodec.Text, obj)
}

func (t Text) Decode(str string, obj any) error {
	return xcodec.DecodeFromString(xcodec.Text, str, obj)
}
