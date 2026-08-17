package dbcodec

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbtype"
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
	return tm.UnixMilli(), nil
}

func (t Milliseconds) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 {
		*ptr = time.Time{}
		return nil
	}

	ms, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}

	*ptr = time.UnixMilli(ms)
	return nil
}
