package xor

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/xanygo/anygo/ds/xcmp"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/internal/encoder"
)

type Option interface {
	withOption(o *config)
}

type optionFunc func(o *config)

func (f optionFunc) withOption(o *config) {
	f(o)
}

var optionNop = optionFunc(func(o *config) {})

func Where(where string, args ...any) Option {
	return optionFunc(func(o *config) {
		o.where = strings.TrimSpace(where)
		o.whereArgs = args
	})
}

func WhereByPK(v any) Option {
	rt := reflect.TypeOf(v)
	return optionFunc(func(o *config) {
		if rt != o.schema.ValueType {
			o.addError(fmt.Errorf("invalid value type WhereByPK(%#v), expect %s", v, o.schema.ValueType.String()))
			return
		}
		pkData, err := o.getEncoder(encoder.ActionSelect).PKNameAndValues(v)
		if err != nil {
			o.addError(err)
			return
		}
		if len(pkData) == 0 {
			o.addError(fmt.Errorf("no pk in %#v", v))
			return
		}
		WhereByMap(pkData).withOption(o)
	})
}

func WhereByMap(kv map[string]any) Option {
	return optionFunc(func(o *config) {
		if len(kv) == 0 {
			o.addError(fmt.Errorf("is zero WhereMap %#v", kv))
			return
		}
		cond := &xdb.Condition{}
		for key, value := range kv {
			cond.And(o.dialect.QuoteIdentifier(key)+"=?", value)
		}
		var err error
		o.where, o.whereArgs, err = cond.Build()
		if err != nil {
			o.addError(err)
		}
	})
}

func WhereByCond(cond *xdb.Condition) Option {
	return optionFunc(func(o *config) {
		var err error
		o.where, o.whereArgs, err = cond.Build()
		if err != nil {
			o.addError(err)
		}
	})
}

func WhereWithCond(fn func(nc *xdb.Condition)) Option {
	cond := &xdb.Condition{}
	fn(cond)
	return WhereByCond(cond)
}

func WhereTrue() Option {
	return optionFunc(func(o *config) {
		o.noWhere = true
	})
}

func Limit[N xcmp.IntegerTypes](n N) Option {
	return optionFunc(func(o *config) {
		o.limit = int(n)
	})
}

func Offset[N xcmp.IntegerTypes](n N) Option {
	return optionFunc(func(o *config) {
		o.offset = int(n)
	})
}

func LimitOffset[A xcmp.IntegerTypes, B xcmp.IntegerTypes](limit A, offset B) Option {
	return optionFunc(func(o *config) {
		o.limit = int(limit)
		o.offset = int(offset)
	})
}

// OrderBy 排序规则，如 "id asc"
func OrderBy(orderBy string) Option {
	return optionFunc(func(o *config) {
		o.orderBy = orderBy
	})
}

// OrderByRand 查询结果随机排序
func OrderByRand() Option {
	return optionFunc(func(o *config) {
		o.orderBy = o.dialect.RandomOrder()
	})
}

// OrderByColumn 多个（N >= 0）字段混合排序，由 NamedArg.Value 指定排序规则：
//
//	Value 值为 true  或者  "asc" 为升序
//	Value 值为 false 或者 "desc" 为降序
func OrderByColumn(columns ...sql.NamedArg) Option {
	if len(columns) == 0 {
		return optionNop
	}
	return optionFunc(func(o *config) {
		lines := make([]string, 0, len(columns))
		for _, item := range columns {
			var flag string
			switch val := item.Value.(type) {
			case bool:
				flag = orderByFlags[val]
			case string:
				flag = strings.ToUpper(val)
				switch flag {
				case "ASC", "DESC", "": // 可以允许空字符串，即相当于 ASC
					// pass
				default:
					o.addError(fmt.Errorf("invalid order by %#v", item))
					continue
				}
			default:
				o.addError(fmt.Errorf("invalid order by %#v", item))
				continue
			}
			lines = append(lines, o.dialect.QuoteIdentifier(item.Name)+" "+flag)
		}
		o.orderBy = strings.Join(lines, ",")
	})
}

var orderByFlags = map[bool]string{
	true:  "ASC",
	false: "DESC",
}

func OrderByPkAsc() Option {
	return orderByPk(true)
}

func OrderByPkDesc() Option {
	return orderByPk(false)
}

func orderByPk(asc bool) Option {
	flag := orderByFlags[asc]
	return optionFunc(func(o *config) {
		names := o.schema.PKColumns().Names()
		if len(names) == 0 {
			o.addError(fmt.Errorf("%w in %s", dbtype.ErrNoPK, o.schema.ValueType.String()))
			return
		}
		if len(names) == 1 {
			o.orderBy = o.dialect.QuoteIdentifier(names[0]) + " " + flag
			return
		}
		lines := make([]string, 0, len(names))
		for _, name := range names {
			lines = append(lines, o.dialect.QuoteIdentifier(name)+" "+flag)
		}
		o.orderBy = strings.Join(lines, ",")
	})
}

// Columns 设置 查询、写入、更新的字段列表
func Columns(cols ...string) Option {
	return optionFunc(func(o *config) {
		o.columns = cols
	})
}

// Ignores  设置 查询、写入、更新时忽略的字段列表
func Ignores(cols ...string) Option {
	return optionFunc(func(o *config) {
		o.ignores = cols
	})
}

func Reset() Option {
	return optionFunc(func(o *config) {
		o.reset()
	})
}
