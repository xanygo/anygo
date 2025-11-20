//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package dbcodec

import (
	"fmt"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

var _ dbtype.Codec = (*DateTime)(nil)

type DateTime struct{}

func (t DateTime) Kind() dbtype.Kind {
	return dbtype.KindDateTime
}

func (t DateTime) Name() string {
	return "dateTime"
}

func (t DateTime) Encode(a any) (any, error) {
	tm, ok := a.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expect time.Time but got %T", a)
	}
	if tm.IsZero() {
		return nil, nil
	}
	return tm.Format(time.DateTime), nil
}

func (t DateTime) Decode(str string, a any) error {
	ptr, ok := a.(*time.Time)
	if !ok {
		return fmt.Errorf("expect *time.Time but got %T", a)
	}
	if len(str) == 0 {
		*ptr = time.Time{}
		return nil
	}
	var tm time.Time
	var err error
	if len(str) == len("2025-11-19T15:13:19Z") {
		tm, err = time.ParseInLocation(time.RFC3339, str, time.Local)
	} else {
		tm, err = time.ParseInLocation(time.DateTime, str, time.Local)
	}
	if err != nil {
		return fmt.Errorf("parse time failed: %w", err)
	}
	*ptr = tm
	return nil
}
