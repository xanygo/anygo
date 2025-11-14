//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"fmt"
	"strconv"
	"time"
)

var _ Codec = (*TimeSpan)(nil)

type TimeSpan struct {
}

func (t TimeSpan) Kind() Kind {
	return KindInt64
}

func (t TimeSpan) Name() string {
	return "timespan"
}

func (t TimeSpan) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	if tm.IsZero() {
		return nil, nil
	}
	return tm.Unix(), nil
}

func (t TimeSpan) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 {
		*ptr = time.Time{}
		return nil
	}
	sec, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	tm := time.Unix(sec, 0)
	*ptr = tm
	return nil
}
