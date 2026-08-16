package encoder

import (
	"time"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

type autoFunc func(schema dbtype.ColumnSchema, val any) (any, bool)

type autoFuncMap map[string]autoFunc

func (ma autoFuncMap) do(schema dbtype.ColumnSchema, val any) (any, bool) {
	fn, ok := ma[schema.Auto]
	if !ok {
		return nil, false
	}
	return fn(schema, val)
}

var insertAutoFns = autoFuncMap{}
var updateAutoFns = autoFuncMap{}

func init() {
	insertAutoFns["Created"] = autoCreatedTimeUnix
	insertAutoFns["CreatedUnix"] = autoCreatedTimeUnix
	insertAutoFns["CreatedNano"] = autoCreatedTimeNano

	insertAutoFns["Updated"] = autoUpdatedTimeUnix
	insertAutoFns["UpdatedUnix"] = autoUpdatedTimeUnix
	insertAutoFns["UpdatedNano"] = autoUpdatedTimeNano
}

func autoCreatedTimeUnix(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true
		}
	case int64:
		if tv == 0 {
			return time.Now().Unix(), true
		}
	}
	return nil, false
}

func autoCreatedTimeNano(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true
		}
	case int64:
		if tv == 0 {
			return time.Now().UnixNano(), true
		}
	}
	return nil, false
}

func autoUpdatedTimeUnix(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true
	case int64:
		return time.Now().Unix(), true
	}
	return nil, false
}

func autoUpdatedTimeNano(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true
	case int64:
		return time.Now().UnixNano(), true
	}
	return nil, false
}
