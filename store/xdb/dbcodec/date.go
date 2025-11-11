//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"fmt"
	"time"
)

var _ Codec = (*Date)(nil)

type Date struct{}

func (t Date) Name() string {
	return "date"
}

func (t Date) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	if tm.IsZero() {
		return nil, nil
	}
	return tm.Format("2006"), nil
}

func (t Date) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 {
		*ptr = time.Time{}
		return nil
	}
	tm, err := time.ParseInLocation("2006", str, time.Local)
	if err != nil {
		return fmt.Errorf("parse time failed: %w", err)
	}
	*ptr = tm
	return nil
}
