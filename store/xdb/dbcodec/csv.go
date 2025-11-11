//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import "github.com/xanygo/anygo/xcodec"

var _ Codec = (*CSV)(nil)

type CSV struct {
}

func (j CSV) Name() string {
	return "csv"
}

func (j CSV) Encode(obj any) (any, error) {
	if obj == nil {
		return "", nil
	}
	return xcodec.EncodeToString(xcodec.CSV, obj)
}

func (j CSV) Decode(str string, obj any) error {
	if len(str) == 0 {
		return nil
	}
	return xcodec.DecodeFromString(xcodec.CSV, str, obj)
}
