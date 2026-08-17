package dbcodec

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var _ dbtype.Codec = (*Microseconds)(nil)
var _ dbtype.HasKind = (*Microseconds)(nil)

// Microseconds 用于 time.Time 类型的数据，将时间编码为 time.Time.UnixMicro()
type Microseconds struct{}

func (t Microseconds) Kind() dbtype.Kind {
	return dbtype.KindInt64
}

func (t Microseconds) Name() string {
	return "microseconds"
}

func (t Microseconds) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	return tm.UnixMicro(), nil
}

func (t Microseconds) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 {
		*ptr = time.Time{}
		return nil
	}

	us, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}

	*ptr = time.UnixMicro(us)
	return nil
}
