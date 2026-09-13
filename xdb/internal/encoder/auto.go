package encoder

import (
	"time"
	"uuid"

	"github.com/xanygo/anygo/xdb/dbtype"
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
	insertAutoFns["CreatedNano"] = autoCreatedTimeNano
	insertAutoFns["CreatedMS"] = autoCreatedTimeMS
	insertAutoFns["UUID4"] = autoCreatedUUID4
	insertAutoFns["UUID7"] = autoCreatedUUID7

	// -------- Update ------------------------------
	updateAutoFns["Updated"] = autoUpdatedTimeUnix
	updateAutoFns["UpdatedNano"] = autoUpdatedTimeNano
	updateAutoFns["UpdatedMS"] = autoUpdatedTimeMS

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

func autoCreatedTimeMS(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true
		}
	case int64:
		if tv == 0 {
			return time.Now().UnixMilli(), true
		}
	}
	return nil, false
}

var uuidZero = uuid.Nil()

func autoCreatedUUID4(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case uuid.UUID:
		if tv == uuidZero {
			return uuid.NewV4(), true
		}
	case [16]byte:
		if tv == uuidZero {
			return uuid.NewV4(), true
		}
	case string:
		if tv == "" {
			return uuid.NewV4().String(), true
		}
	}
	return nil, false
}

func autoCreatedUUID7(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch tv := val.(type) {
	case uuid.UUID:
		if tv == uuidZero {
			return uuid.NewV7(), true
		}
	case [16]byte:
		if tv == uuidZero {
			return uuid.NewV7(), true
		}
	case string:
		if tv == "" {
			return uuid.NewV7().String(), true
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

func autoUpdatedTimeMS(schema dbtype.ColumnSchema, val any) (any, bool) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true
	case int64:
		return time.Now().UnixNano(), true
	}
	return nil, false
}
