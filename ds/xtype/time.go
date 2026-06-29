//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-01-12

package xtype

import (
	"bytes"
	"encoding"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Duration 时间长度类型，可用于配置的字段定义，相比原始的 time.Duration，
// 支持了 MarshalText 和 UnmarshalText，可以被 json、yaml、toml 等直接解析和编码
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) Nanoseconds() int64 { return int64(d) }

func (d Duration) Microseconds() int64 { return int64(d) / 1e3 }

func (d Duration) Milliseconds() int64 { return int64(d) / 1e6 }

func (d Duration) Seconds() float64 {
	return d.Duration().Seconds()
}

func (d Duration) Minutes() float64 {
	return d.Duration().Minutes()
}

func (d Duration) Hours() float64 {
	return d.Duration().Hours()
}

func (d Duration) Truncate(m Duration) Duration {
	if m <= 0 {
		return d
	}
	return d - d%m
}

// String 格式化输出,保留 3 位小数有效位
func (d Duration) String() string {
	if d == 0 {
		return "0s"
	}
	dv := d.Duration()
	str := dv.String()
	if !strings.HasSuffix(str, "s") {
		return str
	}
	dot := strings.IndexByte(str, '.')
	if dot < 0 || dot >= len(str)-4 {
		return str
	}
	var ui int
	for ui = dot + 1; ui < len(str); ui++ {
		if str[ui] >= '0' && str[ui] <= '9' {
			continue
		} else {
			break
		}
	}
	pre := strings.TrimRight(str[:dot+4], "0")
	pre = strings.TrimRight(pre, ".")
	return pre + str[ui:]
}

var _ encoding.TextMarshaler = Duration(0)

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration().String()), nil
}

var _ encoding.TextUnmarshaler = (*Duration)(nil)

func (d *Duration) UnmarshalText(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	// 若是数字，没有单位，默认当做 ms 解析
	last := b[len(b)-1]
	if last >= '0' && last <= '9' {
		num, err := strconv.ParseUint(string(b), 10, 64)
		if err == nil {
			*d = Duration(time.Duration(num) * time.Millisecond)
		}
		return err
	}

	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	v, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

var _ json.Unmarshaler = (*Duration)(nil)

func (d *Duration) UnmarshalJSON(b []byte) error {
	return d.UnmarshalText(b)
}

// UnixTimestamp 表示 Unix 时间戳，单位为秒，自 UTC 1970-01-01 00:00:00 起计算。
//
// Timespan 底层类型为 int64，零值表示 Unix epoch（1970-01-01 00:00:00 UTC）。
//
// # JSON 编码时，Timestamp 会被编码为秒级 Unix 时间戳：1719110400
//
// JSON 解码时，同时支持数字和字符串形式：1719110400, "1719110400"
//
// 可通过 Time 方法转换为 time.Time：
//
//	ts := Timestamp(1719110400)
//	t := ts.Time()
type UnixTimestamp int64

func (t UnixTimestamp) Time() time.Time {
	return time.Unix(int64(t), 0)
}

func (t UnixTimestamp) Unix() int64 {
	return int64(t)
}

func (t *UnixTimestamp) SetTime(v time.Time) {
	*t = UnixTimestamp(v.Unix())
}

var _ json.Marshaler = UnixTimestamp(0)

func (t UnixTimestamp) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, int64(t), 10), nil
}

var _ json.Unmarshaler = (*UnixTimestamp)(nil)

func (t *UnixTimestamp) UnmarshalJSON(data []byte) error {
	// null
	if string(data) == "null" {
		*t = 0
		return nil
	}

	// 数字
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*t = UnixTimestamp(n)
		return nil
	}

	// 字符串数字
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = UnixTimestamp(n)
		return nil
	}

	return strconv.ErrSyntax
}

var _ encoding.TextMarshaler = UnixTimestamp(0)

func (t UnixTimestamp) MarshalText() ([]byte, error) {
	return t.MarshalJSON()
}

var _ encoding.TextUnmarshaler = (*UnixTimestamp)(nil)

func (t *UnixTimestamp) UnmarshalText(b []byte) error {
	return t.UnmarshalJSON(b)
}
