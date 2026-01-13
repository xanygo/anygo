//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-13

package xtype

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ByteCount 字节大小,最大可表示 EiB
type ByteCount uint64

type byteCountUnit struct {
	size uint64
	name string
}

var byteCountUnits = []byteCountUnit{
	{size: 1 << 60, name: "EiB"},
	{size: 1 << 50, name: "PiB"},
	{size: 1 << 40, name: "TiB"},
	{size: 1 << 30, name: "GiB"},
	{size: 1 << 20, name: "MiB"},
	{size: 1 << 10, name: "KiB"},
	{size: 1, name: "B"},
}

// String 将字节大小格式化为字符串，
// 如:
//   - 1B，1KiB，1MiB，1GiB，1TiB，1PiB，1EiB
//   - 1PiB1GiB1B
func (d ByteCount) String() string {
	if d == 0 {
		return "0B"
	}
	var buf []byte
	remain := uint64(d)

	for _, u := range byteCountUnits {
		if remain >= u.size {
			n := remain / u.size
			remain %= u.size
			buf = append(buf, strconv.FormatUint(n, 10)...)
			buf = append(buf, u.name...)
			buf = append(buf, ' ')
		}
	}
	buf = bytes.TrimSpace(buf)
	return string(buf)
}

func ParserByteCount(s string) (ByteCount, error) {
	var d ByteCount
	err := d.Parser(s)
	return d, err
}

func (d *ByteCount) Parser(s string) error {
	if len(s) == 0 {
		return errors.New("empty string")
	}

	// 全部是数字的情况
	last := s[len(s)-1]
	if last >= '0' && last <= '9' {
		num, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			*d = ByteCount(num)
		}
		return err
	}

	if !strings.HasSuffix(s, "B") {
		return fmt.Errorf("%q missing 'B' character suffix", s)
	}

	var num uint64
	var start int
	for i := 0; i < len(s); i++ {
		if s[i] == 'B' {
			a, err := d.parserOne(s[start : i+1])
			if err != nil {
				return err
			}
			if num > math.MaxUint64-a {
				return errors.New("too many bytes, overflow uint64")
			}
			start = i + 1
			num += a
		}
	}

	*d = ByteCount(num)
	return nil
}

// parserOne 只解析一部分，如 1 GiB 或者 1B
func (d *ByteCount) parserOne(s string) (uint64, error) {
	s = strings.TrimSpace(s)

	last := s[len(s)-1]
	if last != 'B' || len(s) < 2 {
		return 0, fmt.Errorf("invalid byte count %q", s)
	}
	if c := s[len(s)-2]; c >= '0' && c <= '9' { // 如 100B 或者 1B
		return strconv.ParseUint(s[:len(s)-1], 10, 64)
	} else if c != 'i' { // 如 1KiB
		return 0, fmt.Errorf("invalid byte count %q", s)
	}
	// 最短为 4，如  1KiB
	if len(s) < 4 {
		return 0, fmt.Errorf("invalid byte count %q，to short", s)
	}
	power := 0
	switch s[len(s)-3] {
	case 'K': // KiB
		power = 1
	case 'M': // MiB
		power = 2
	case 'G': // GiB
		power = 3
	case 'T': // TiB
		power = 4
	case 'P': // PiB
		power = 5
	case 'E': // EiB
		power = 6
	case 'Z': // ZiB
		power = 7
	default:
		// Invalid suffix.
		return 0, fmt.Errorf("invalid byte count suffix %q", s)
	}
	m := uint64(1)
	for i := 0; i < power; i++ {
		m *= 1024
	}
	n, err := strconv.ParseUint(s[:len(s)-3], 10, 64)
	if err != nil {
		return 0, err
	}
	un := n
	if un > math.MaxUint64/m {
		// Overflow.
		return 0, fmt.Errorf("byte count overflow %q", s)
	}
	un *= m
	if un > uint64(math.MaxInt64) {
		// Overflow.
		return 0, fmt.Errorf("byte count overflow %q", s)
	}
	return un, nil
}

var _ encoding.TextUnmarshaler = (*ByteCount)(nil)

func (d *ByteCount) UnmarshalText(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	return d.Parser(string(b))
}

var _ json.Unmarshaler = (*ByteCount)(nil)

func (d *ByteCount) UnmarshalJSON(b []byte) error {
	return d.UnmarshalText(b)
}
