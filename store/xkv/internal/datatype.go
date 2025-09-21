//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-21

package internal

import "errors"

type DataType uint8

const (
	DataTypeUnset = iota
	DataTypeString
	DataTypeList
	DataTypeHash
	DataTypeSet
	DataTypeZSet
)

var ErrInvalidType = errors.New("key exists, but type not match")
