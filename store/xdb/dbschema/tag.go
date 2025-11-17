//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-15

package dbschema

import (
	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
)

var tagName = xsync.OnceInit[string]{
	New: func() string {
		return "db"
	},
}

func TagName() string {
	return tagName.Load()
}

func SetTagName(name string) {
	if name == "" {
		panic("empty tag name")
	}
	tagName.Store(name)
}

const (
	TagPrimaryKey = "primaryKey"
	TagPK         = "pk" // TagPrimaryKey 的缩写

	TagCodec = "codec"

	TagAutoIncrement = "autoIncrement"
	TagAutoIncr      = "autoInc" // TagAutoIncrement 的缩写

	TagUnique = "unique" // 唯一键，不需要值

	// TagIndex 标记此字段需要添加索引
	// 示例：
	// index                   -> 创建独立索引，索引名称为 idx_字段名
	// index:idx_uid           -> 创建独立索引，索引名称为 idx_uid
	// index:idx_uid_class,1   -> 创建联合索引，索引名称为 idx_uid_class，此字段在索引中排序为 1
	TagIndex       = "index"
	TagUniqueIndex = "uniqueIndex" // 值格式同 TagIndex

	TagSize = "size"

	TagNotNull = "notNull"

	TagType = "type" // 数据类型，必须是有效的 dbschema.Kind 的值

	// TagDefault 默认值，格式为 default:[[fn|string|number]|]value
	// 示例：
	// 1.默认值为空字符串：“name,default”
	// 2.默认值为数字：“name,default:number|123”
	// 3.默认值为字符串：“name,default:string|hello”
	// 4.默认值为数据库函数：“name,default:fn|CURRENT_TIMESTAMP”
	TagDefault = "default" // 默认值
)

func TagHasAutoIncr(tag xstruct.Tag) bool {
	return tag.Has(TagAutoIncr) || tag.Has(TagAutoIncrement)
}

func TagHasPrimaryKey(tag xstruct.Tag) bool {
	return tag.Has(TagPK) || tag.Has(TagPrimaryKey)
}

func TagCodecName(tag xstruct.Tag) string {
	name := tag.Value(TagCodec)
	if name != "" {
		return name
	}
	return dbcodec.TextName
}
