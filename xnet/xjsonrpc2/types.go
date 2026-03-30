//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package xjsonrpc2

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xanygo/anygo/xcodec"
)

type ID interface {
	Bytes() []byte
}

type Int64ID int64

func (id Int64ID) Bytes() []byte {
	return []byte(strconv.FormatInt(int64(id), 10))
}

type StringID string

func (id StringID) Bytes() []byte {
	return []byte(strconv.Quote(string(id)))
}

func idBytes(id ID) []byte {
	if id == nil {
		return nil
	}
	return id.Bytes()
}

func parserID(bf []byte) (ID, error) {
	if len(bf) == 0 || bytes.Equal(bf, []byte("null")) {
		return nil, nil
	}
	var id any
	err := xcodec.JSON.Decode(bf, &id)
	if err != nil {
		return nil, err
	}
	switch val := id.(type) {
	case string:
		return StringID(val), nil
	case float64: // int64 类型数字，解析为 any 时，会优先解析为 float64 类型
		return Int64ID(int64(val)), nil
	default:
		return nil, fmt.Errorf("invalid id type: %#v", id)
	}
}
