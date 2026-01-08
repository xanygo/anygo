//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-07

package xtype

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
)

const (
	TriNull  TriState = iota // 状态：未设置值
	TriTrue                  // 状态：true
	TriFalse                 // 状态：false
)

// TriState 包含 null、true、false 3 个状态
type TriState int8

// IsTrue 是否真值
func (ts TriState) IsTrue() bool {
	return ts == TriTrue
}

// IsFalse 是否假值
func (ts TriState) IsFalse() bool {
	return ts == TriFalse
}

// IsNull 是否未设置值
func (ts TriState) IsNull() bool {
	return ts == TriNull
}

// NotNull 值是否为 true 或者 false
func (ts TriState) NotNull() bool {
	switch ts {
	case TriTrue, TriFalse:
		return true
	default:
		return false
	}
}

func (ts TriState) Valid() bool {
	switch ts {
	case TriTrue, TriFalse, TriNull:
		return true
	default:
		return false
	}
}

var _ encoding.TextMarshaler = TriNull

func (ts TriState) MarshalText() ([]byte, error) {
	return ts.MarshalJSON()
}

func (ts *TriState) UnmarshalText(bs []byte) error {
	return ts.UnmarshalJSON(bs)
}

var _ json.Marshaler = (*TriState)(nil)
var _ json.Unmarshaler = (*TriState)(nil)

func (ts TriState) MarshalJSON() ([]byte, error) {
	switch ts {
	case TriNull:
		return []byte("null"), nil
	case TriTrue:
		return []byte("true"), nil
	case TriFalse:
		return []byte("false"), nil
	default:
		return nil, fmt.Errorf("invalid TriState(%d)", ts)
	}
}

func (ts *TriState) UnmarshalJSON(data []byte) error {
	if ts == nil {
		return errors.New("unmarshal TriState on nil receiver")
	}

	bs := bytes.Trim(data, `"`)
	switch string(bs) {
	case "true", "yes", "on":
		*ts = TriTrue
	case "false", "no", "off":
		*ts = TriFalse
	case "", "null":
		*ts = TriNull
	default:
		return fmt.Errorf("invalid TriState text %q", data)
	}
	return nil
}

var _ flag.Value = (*TriState)(nil)

func (ts *TriState) Set(s string) error {
	return ts.UnmarshalText([]byte(s))
}

func (ts TriState) String() string {
	b, _ := ts.MarshalText()
	return string(b)
}
