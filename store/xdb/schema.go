//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-11

package xdb

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xstruct"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/store/xdb/dbcodec"
	"github.com/xanygo/anygo/store/xdb/dialect"
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

// 一下常量带有 Mig 的，表示目前只在 Migrate 逻辑中使用到
const (
	tagPrimaryKey = "primaryKey"
	tagCodec      = "codec"

	tagAutoIncr      = "autoInc"
	tagAutoIncrement = "autoIncrement" // tagAutoIncr 的缩写

	tagMigUnique = "unique" // 唯一键，不需要值

	// 标记此字段需要添加索引
	// 示例：
	// index                   -> 创建独立索引，索引名称为 idx_字段名
	// index:idx_uid           -> 创建独立索引，索引名称为 idx_uid
	// index:idx_uid_class,1   -> 创建联合索引，索引名称为 idx_uid_class，此字段在索引中排序为 1
	tagMigIndex       = "index"
	tagMigUniqueIndex = "uniqueIndex" // 值格式同 tagIndex

	tagMigSize    = "size"
	tagMigNotNull = "notNull"

	tagDefault = "default" // 默认值
)

type tableSchema struct {
	Table  string
	Fields []schemaField
}

type schemaField struct {
	Name          string // 字段名
	IsPrimaryKey  bool
	AutoIncrement bool // 自增长
	Kind          dbcodec.Kind
	Unique        bool              // 是否唯一键
	Index         *schemaIndexValue // 索引的名称
	UniqueIndex   *schemaIndexValue // 唯一索引
	Size          int               // 定义列数据类型的大小或长度
	NotNull       bool
}

type schemaIndexValue struct {
	FieldName  string
	IndexName  string // 索引
	FieldOrder int    // 字段在索引中的顺序
}

func (sf schemaField) String(d dialect.Dialect, sd dialect.SchemaDialect) string {
	var sb strings.Builder
	sb.WriteString(d.QuoteIdentifier(sf.Name))
	sb.WriteString(" ")
	baseType := sd.ColumnType(sf.Kind, sf.Size)
	if sf.AutoIncrement {
		baseType = sd.AutoIncrementColumnType(baseType, sf.IsPrimaryKey)
	} else if sf.IsPrimaryKey {
		baseType += " PRIMARY KEY"
	}
	sb.WriteString(baseType)
	if sf.Unique {
		sb.WriteString(" UNIQUE")
	}
	if sf.NotNull {
		sb.WriteString(" NOT NULL")
	}
	return sb.String()
}

func isTagAutoIncr(tag xstruct.Tag) bool {
	return tag.Has(tagAutoIncr) || tag.Has(tagAutoIncrement)
}

func getCodecName(tag xstruct.Tag) string {
	name := tag.Value(tagCodec)
	if name != "" {
		return name
	}
	return dbcodec.TextName
}

func (ts *tableSchema) AlterAddSQL(d dialect.Dialect) ([]string, error) {
	sd, ok := d.(dialect.SchemaDialect)
	if !ok {
		return nil, fmt.Errorf(" %q dialect does not implement SchemaDialect", d.Name())
	}
	var results []string
	str := ts.addTableSQL(d, sd)
	results = append(results, str)
	return results, nil
}

func (ts *tableSchema) addTableSQL(d dialect.Dialect, sd dialect.SchemaDialect) string {
	str := sd.CreateTableIfNotExists(ts.Table) + " (\n"

	var lines []string
	indexMap := map[string][]*schemaIndexValue{}
	uniqIndexMap := map[string][]*schemaIndexValue{}
	for _, field := range ts.Fields {
		tmp := field.String(d, sd)
		lines = append(lines, tmp)
		if field.Index != nil {
			indexName := field.Index.IndexName
			indexMap[indexName] = append(indexMap[indexName], field.Index)
		}
		if field.UniqueIndex != nil {
			indexName := field.UniqueIndex.IndexName
			uniqIndexMap[indexName] = append(uniqIndexMap[indexName], field.UniqueIndex)
		}
	}

	for indexName, indexes := range indexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var names []string
		for _, index := range indexes {
			names = append(names, d.QuoteIdentifier(index.FieldName))
		}
		tmp := "Index " + indexName + "(" + strings.Join(names, ",") + ")"
		lines = append(lines, tmp)
	}

	for indexName, indexes := range uniqIndexMap {
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].FieldOrder < indexes[j].FieldOrder
		})
		var names []string
		for _, index := range indexes {
			names = append(names, d.QuoteIdentifier(index.FieldName))
		}
		tmp := "UNIQUE KEY " + indexName + "(" + strings.Join(names, ",") + ")"
		lines = append(lines, tmp)
	}
	str += strings.Join(lines, ",\n") + ")"
	return str
}

type schemaParser struct{}

func (sp schemaParser) Get(obj any) (*tableSchema, error) {
	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Struct {
		return nil, errors.New("obj is not a struct")
	}

	sc := &tableSchema{}
	if ht, ok := obj.(HasTable); ok {
		sc.Table = ht.TableName()
	}

	tn := TagName()
	err := zreflect.RangeStructFields(rt, func(field reflect.StructField) error {
		if !field.IsExported() {
			return nil
		}

		tag := xstruct.ParserTagCached(field.Tag, tn)
		name := tag.Name()
		if name == "-" || name == "" {
			return nil
		}
		scf, err := sp.parserField(field, tag)
		if err != nil {
			return fmt.Errorf("field=%q: %w", field.Name, err)
		}
		sc.Fields = append(sc.Fields, *scf)
		return nil
	})
	return sc, err
}

func (sp schemaParser) parserIndex(fieldName string, tag xstruct.Tag, isUniq bool) (*schemaIndexValue, error) {
	indexTagName := tagMigIndex
	indexNamePrefix := "idx_"
	if isUniq {
		indexTagName = tagMigUniqueIndex
		indexNamePrefix += "uniq_"
	}
	index, has := tag.Get(indexTagName)
	if !has {
		return nil, nil
	}

	if index == "" {
		return &schemaIndexValue{
			IndexName:  indexNamePrefix + fieldName,
			FieldOrder: -1,
		}, nil
	}

	idxName, order, found := strings.Cut(index, ",")
	if found {
		num, err := strconv.Atoi(order)
		if err != nil || num < 0 {
			return nil, fmt.Errorf("invalid field order in: %q", order)
		}
		return &schemaIndexValue{
			IndexName:  idxName,
			FieldOrder: num,
		}, nil
	}

	return &schemaIndexValue{
		IndexName:  index,
		FieldOrder: 0,
	}, nil
}

func (sp schemaParser) parserField(f reflect.StructField, tag xstruct.Tag) (*schemaField, error) {
	field := &schemaField{
		Name:          tag.Name(),
		AutoIncrement: isTagAutoIncr(tag),
		IsPrimaryKey:  tag.Has(tagPrimaryKey),
		NotNull:       tag.Has(tagMigNotNull),
		Unique:        tag.Has(tagMigUnique),
	}

	var err error
	field.Index, err = sp.parserIndex(field.Name, tag, false)
	if err == nil {
		field.UniqueIndex, err = sp.parserIndex(field.Name, tag, true)
	}
	if err != nil {
		return nil, err
	}

	if size, has := tag.Get(tagMigSize); has {
		num, err0 := strconv.Atoi(size)
		if err0 != nil || num <= 0 {
			return nil, fmt.Errorf("invalid size: %s", size)
		}
		field.Size = num
	}

	codec := tag.Value(tagCodec)
	if codec != "" {
		cc, err0 := dbcodec.Find(codec)
		if err0 != nil {
			return nil, err0
		}
		field.Kind = cc.Kind()
	}
	if !field.Kind.IsValid() {
		field.Kind, err = sp.typeToKind(f.Type)
	}
	return field, err
}

var typeToKindMap = map[reflect.Kind]dbcodec.Kind{
	reflect.Int:   dbcodec.KindInt,
	reflect.Int8:  dbcodec.KindInt8,
	reflect.Int16: dbcodec.KindInt16,
	reflect.Int32: dbcodec.KindInt32,
	reflect.Int64: dbcodec.KindInt64,

	reflect.Uint:   dbcodec.KindUint,
	reflect.Uint8:  dbcodec.KindUint8,
	reflect.Uint16: dbcodec.KindUint16,
	reflect.Uint32: dbcodec.KindUint32,
	reflect.Uint64: dbcodec.KindUint64,

	reflect.Float32: dbcodec.KindFloat32,
	reflect.Float64: dbcodec.KindFloat64,

	reflect.String: dbcodec.KindString,

	reflect.Bool: dbcodec.KindBoolean,

	reflect.Struct:    dbcodec.KindJSON,
	reflect.Map:       dbcodec.KindJSON,
	reflect.Slice:     dbcodec.KindJSON,
	reflect.Array:     dbcodec.KindJSON,
	reflect.Interface: dbcodec.KindJSON,
}

var specTypeToKindMap = map[reflect.Type]dbcodec.Kind{
	reflect.TypeOf(time.Time{}): dbcodec.KindDateTime,
	reflect.TypeOf([]byte(nil)): dbcodec.KindBinary,
}

func (sp schemaParser) typeToKind(rt reflect.Type) (dbcodec.Kind, error) {
	if k, ok := specTypeToKindMap[rt]; ok {
		return k, nil
	}
	kind := rt.Kind()
	if k, ok := typeToKindMap[kind]; ok {
		return k, nil
	}
	if kind == reflect.Pointer {
		return sp.typeToKind(rt.Elem())
	}

	return dbcodec.KindInvalid, fmt.Errorf("invalid data type: %s", rt.String())
}
