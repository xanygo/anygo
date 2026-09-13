package dbcodec

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xanygo/anygo/xdb/dbtype"
)

var _ dbtype.Codec = (*Milliseconds)(nil)
var _ dbtype.HasKind = (*Milliseconds)(nil)

// Milliseconds 用于 time.Time 类型的数据，将时间编码为 time.Time.UnixMilli()
type Milliseconds struct{}

func (t Milliseconds) Kind() dbtype.Kind {
	return dbtype.KindInt64
}

func (t Milliseconds) Name() string {
	return "milliseconds"
}

func (t Milliseconds) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	// 将 time.Time{} 特殊处理，避免默认值编码后得到的是负数
	if tm.IsZero() {
		return 0, nil
	}
	return tm.UnixMilli(), nil
}

func (t Milliseconds) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	// 这里特殊处理 0: 和 Encode 对应，是 0 的话就当作 time.Time{}
	// 影响：1970-01-01 00:00:00 UTC 的表示
	if len(str) == 0 || str == "0" {
		*ptr = time.Time{}
		return nil
	}

	ms, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("parser Milliseconds %q: %w", str, err)
	}

	*ptr = time.UnixMilli(ms)
	return nil
}
