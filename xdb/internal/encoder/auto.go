package encoder

import (
	"fmt"
	"time"
	"uuid"

	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xtime"
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
//
// 返回值 (any, bool, error) -> (新值, 已替换标记, 错误)
type autoFunc func(schema dbtype.ColumnSchema, val any) (any, bool, error)

type autoFuncMap map[string]autoFunc

func (ma autoFuncMap) do(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	fn, ok := ma[schema.Auto]
	if !ok {
		return nil, false, nil
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

func autoNotSupportErr(val any, auto string) error {
	return fmt.Errorf("%T not support auto=%s", val, auto)
}

func autoNowTime(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true, nil
		}
		return val, false, nil
	case xtime.DateInt:
		if tv.Time().IsZero() {
			return xtime.DateIntOf(time.Now()), true, nil
		}
		return val, false, nil
	case xtime.TimestampSecond:
		if tv == 0 {
			return time.Now().Unix(), true, nil
		}
		return val, false, nil
	}
	return val, false, autoNotSupportErr(val, "Now")
}

// 自增长，uint8 等类型，存储的值较小，容易溢出，故不支持
func autoIncr(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case int:
		return tv + 1, true, nil
	case int64:
		return tv + 1, true, nil
	case uint64:
		return tv + 1, true, nil
	case float64:
		return tv + 1, true, nil
	case float32:
		return tv + 1, true, nil
	default:
		return nil, false, autoNotSupportErr(val, "Incr")
	}
}

func autoCreatedTimeUnix(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true, nil
		}
		return val, false, nil
	case int64:
		if tv == 0 {
			return time.Now().Unix(), true, nil
		}
		return val, false, nil
	case xtime.DateInt:
		if tv.Time().IsZero() {
			return xtime.DateIntOf(time.Now()), true, nil
		}
		return val, false, nil
	case xtime.TimestampSecond:
		if tv == 0 {
			return time.Now().Unix(), true, nil
		}
		return val, false, nil
	default:
		return val, false, autoNotSupportErr(val, "Created")
	}
}

func autoCreatedTimeNano(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true, nil
		}
		return val, false, nil
	case int64:
		if tv == 0 {
			return time.Now().UnixNano(), true, nil
		}
		return tv, false, nil
	default:
		return nil, false, autoNotSupportErr(val, "CreatedNano")
	}
}

func autoCreatedTimeMS(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case time.Time:
		if tv.IsZero() {
			return time.Now(), true, nil
		}
		return val, false, nil
	case int64:
		if tv == 0 {
			return time.Now().UnixMilli(), true, nil
		}
		return val, false, nil
	default:
		return nil, false, autoNotSupportErr(val, "CreatedMS")
	}
}

var uuidZero = uuid.Nil()

func autoCreatedUUID4(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case uuid.UUID:
		if tv == uuidZero {
			return uuid.NewV4(), true, nil
		}
		return val, false, nil
	case [16]byte:
		if tv == uuidZero {
			return uuid.NewV4(), true, nil
		}
		return val, false, nil
	case string:
		if tv == "" {
			return uuid.NewV4().String(), true, nil
		}
		return val, false, nil
	default:
		return nil, false, autoNotSupportErr(val, "UUID4")
	}
}

func autoCreatedUUID7(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch tv := val.(type) {
	case uuid.UUID:
		if tv == uuidZero {
			return uuid.NewV7(), true, nil
		}
		return val, false, nil
	case [16]byte:
		if tv == uuidZero {
			return uuid.NewV7(), true, nil
		}
		return val, false, nil
	case string:
		if tv == "" {
			return uuid.NewV7().String(), true, nil
		}
		return val, false, nil
	default:
		return nil, false, autoNotSupportErr(val, "UUID7")
	}
}

func autoUpdatedTimeUnix(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true, nil
	case int64:
		return time.Now().Unix(), true, nil
	case xtime.DateInt:
		return xtime.DateIntOf(time.Now()), true, nil
	case xtime.TimestampSecond:
		return time.Now().Unix(), true, nil
	default:
		return nil, false, autoNotSupportErr(val, "Updated")
	}
}

func autoUpdatedTimeNano(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true, nil
	case int64:
		return time.Now().UnixNano(), true, nil
	default:
		return nil, false, autoNotSupportErr(val, "UpdatedNano")
	}
}

func autoUpdatedTimeMS(schema dbtype.ColumnSchema, val any) (any, bool, error) {
	switch val.(type) {
	case time.Time:
		return time.Now(), true, nil
	case int64:
		return time.Now().UnixNano(), true, nil
	default:
		return nil, false, autoNotSupportErr(val, "UpdatedMS")
	}
}
