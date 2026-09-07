package xtime

import (
	"encoding"
	"encoding/json"
	"strconv"
	"time"
)

// TimestampSecond 表示 Unix 时间戳，单位为秒，自 UTC 1970-01-01 00:00:00 起计算。
//
//	Timespan 底层类型为 int64，零值表示 Unix epoch（1970-01-01 00:00:00 UTC）。
//	JSON 编码时，Timestamp 会被编码为秒级 Unix 时间戳：1719110400
//	JSON 解码时，同时支持数字和字符串形式：1719110400, "1719110400"
//	可通过 Time 方法转换为 time.Time：
//
//	ts := TimestampSecond(1719110400)
//	t := ts.Time()
type TimestampSecond int64

func (t TimestampSecond) Time() time.Time {
	return time.Unix(int64(t), 0)
}

func (t TimestampSecond) Unix() int64 {
	return int64(t)
}

func (t *TimestampSecond) SetTime(v time.Time) {
	*t = TimestampSecond(v.Unix())
}

var _ json.Marshaler = TimestampSecond(0)

func (t TimestampSecond) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, int64(t), 10), nil
}

var _ json.Unmarshaler = (*TimestampSecond)(nil)

func (t *TimestampSecond) UnmarshalJSON(data []byte) error {
	// null
	if string(data) == "null" {
		*t = 0
		return nil
	}

	// 数字
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*t = TimestampSecond(n)
		return nil
	}

	// 字符串数字
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = TimestampSecond(n)
		return nil
	}

	return strconv.ErrSyntax
}

var _ encoding.TextMarshaler = TimestampSecond(0)

func (t TimestampSecond) MarshalText() ([]byte, error) {
	return t.MarshalJSON()
}

var _ encoding.TextUnmarshaler = (*TimestampSecond)(nil)

func (t *TimestampSecond) UnmarshalText(b []byte) error {
	return t.UnmarshalJSON(b)
}
