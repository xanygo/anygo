package encoder

import (
	"time"

	"github.com/xanygo/anygo/store/xdb/dbtype"
)

//	type User struct{
//	 Name string `db:"name,pk"`
//
//	 CreatedAt time.Time `db:"c,auto=Created"`   // auto=Created -> insert 时，若为空值，自动赋值当前时间
//	 CreatedAt2 int64 `db:"c2,auto=Created"`      // auto=Created -> insert 时，若为空值，自动赋值当前时间戳
//
//	 Updated time.Time `db:"u,auto=Updated"`     // auto=Updated -> insert/update 时，自动赋值当前时间
//	 Updated2 int64 `db:"u2,auto=Updated"`       // auto=Updated -> insert/update 时，自动赋值当前时间戳
//
//	 OtherTime int64 `db:"other,auto=Now"`       // auto=Now -> insert/update 时，自动赋值当前时间
//	}
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
	// 字段名称 => 自动化赋值函数

	// -------- Insert ------------------------------
	insertAutoFns["Created"] = autoCreatedTimeUnix
	insertAutoFns["CreatedUnix"] = autoCreatedTimeUnix
	insertAutoFns["CreatedNano"] = autoCreatedTimeNano

	// -------- Update ------------------------------
	updateAutoFns["Updated"] = autoUpdatedTimeUnix
	updateAutoFns["UpdatedUnix"] = autoUpdatedTimeUnix
	updateAutoFns["UpdatedNano"] = autoUpdatedTimeNano

	// 只需要注册 update，updateAutoFns 在 insert 时总是会被执行
	updateAutoFns["Now"] = autoNowTime
	updateAutoFns["Incr"] = autoIncr
}

func autoNowTime(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true
		}
	}
	return nil, false
}

// 自增长，uint8 等类型，存储的值较小，容易溢出，故不支持
func autoIncr(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case int:
		return tv + 1, true
	case int64:
		return tv + 1, true
	case uint64:
		return tv + 1, true
	case float64:
		return tv + 1, true
	case float32:
		return tv + 1, true
	default:
		return nil, false
	}
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
