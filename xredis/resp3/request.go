//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-29

package resp3

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Request interface {
	Name() string
	Args() []any
	Bytes(bf *bytes.Buffer) []byte
	ResponseType() DataType
}

func NewRequest(dt DataType, args ...any) Request {
	vs := make(Array, 0, len(args))
	for i, arg := range args {
		value := toString(arg)
		if i == 0 {
			value = strings.ToUpper(value)
		}
		vs = append(vs, BulkString(value))
	}
	return cmd{
		dt:      dt,
		args:    args,
		payload: vs,
	}
}

var _ Request = cmd{}

type cmd struct {
	args    []any
	payload Array
	dt      DataType
}

func (c cmd) ResponseType() DataType {
	return c.dt
}

func (c cmd) Name() string {
	if len(c.args) == 0 {
		return ""
	}
	return toString(c.args[0])
}

func (c cmd) Args() []any {
	if len(c.args) < 1 {
		return nil
	}
	return c.args[1:]
}

func (c cmd) Bytes(bf *bytes.Buffer) []byte {
	return c.payload.Bytes(bf)
}

func toString(obj any) string {
	switch v := obj.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint8:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case bool:
		return strconv.FormatBool(v)
	default:
		if ss, ok := obj.(fmt.Stringer); ok {
			return ss.String()
		}
		return fmt.Sprintf("%v", v)
	}
}
