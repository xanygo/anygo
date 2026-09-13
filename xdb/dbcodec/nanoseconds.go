package dbcodec

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xanygo/anygo/xdb/dbtype"
)

var _ dbtype.Codec = (*Nanoseconds)(nil)
var _ dbtype.HasKind = (*Nanoseconds)(nil)

// Nanoseconds 用于 time.Time 类型的数据，将时间编码为 time.Time.UnixNano()
type Nanoseconds struct{}

func (t Nanoseconds) Kind() dbtype.Kind {
	return dbtype.KindInt64
}

func (t Nanoseconds) Name() string {
	return "nanoseconds"
}

func (t Nanoseconds) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	if tm.IsZero() {
		return 0, nil
	}
	return tm.UnixNano(), nil
}

func (t Nanoseconds) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 || str == "0" {
		*ptr = time.Time{}
		return nil
	}

	ns, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("parser Nanoseconds %q: %w", str, err)
	}

	*ptr = time.Unix(0, ns).UTC()
	return nil
}
