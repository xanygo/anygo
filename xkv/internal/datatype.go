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

	DataTypeAny // 特殊的，但用于判断读取的值不为空时，使用此判断
)

var ErrInvalidType = errors.New("key exists, but type not match")

func (dt DataType) String() string {
	switch dt {
	case DataTypeUnset:
		return "unset"
	case DataTypeString:
		return "string"
	case DataTypeList:
		return "list"
	case DataTypeHash:
		return "hash"
	case DataTypeSet:
		return "set"
	case DataTypeZSet:
		return "zset"
	case DataTypeAny:
		return "any"
	default:
		return "invalid type"
	}
}

func (dt DataType) Equal(dt2 DataType) bool {
	return dt == dt2 || dt == DataTypeAny || dt2 == DataTypeAny
}
