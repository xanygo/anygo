//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package resp3

import (
	"fmt"
	"io"
)

type Reader interface {
	io.Reader
	ReadByte() (byte, error)
	ReadSlice(delim byte) (line []byte, err error)
}

// ReadByType 读取指定的类型数据，Reader 的首位必须是传入的 DataType
func ReadByType(rd Reader, dt DataType) (Element, error) {
	tp, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}
	dt1 := DataType(tp)
	if dt1 == DataTypeNull {
		return Null{}, ErrNil
	}
	if !dt1.Equal(dt) {
		el, ok, err1 := asErrorType(dt1, rd)
		if ok {
			return el, err1
		}
		return nil, fmt.Errorf("invalid data type %s, expect %s", DataType(tp), dt)
	}
	return dt1.Load(rd)
}

func asErrorType(dt DataType, rd Reader) (Element, bool, error) {
	switch dt {
	case DataTypeSimpleError:
		item, err := dt.loadSimpleError(rd)
		if err != nil {
			return nil, true, err
		}
		return item, true, item
	case DataTypeBulkError:
		item, err := dt.loadBulkError(rd)
		if err != nil {
			return nil, true, err
		}
		return item, true, item
	default:
		return nil, false, nil
	}
}

func ReadOneElement(rd Reader) (Element, error) {
	return readOne(rd)
}
