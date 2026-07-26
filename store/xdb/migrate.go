//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-13

package xdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dbtype"
	"github.com/xanygo/anygo/store/xdb/dialect"
)

// Migrate 自动创建、添加字段（非生产环境使用）
func Migrate(db DBCore, obj any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return MigrateWithTable(ctx, db, obj, "")
}

// MigrateWithTable 自动创建、添加字段（非生产环境使用）
//
// obj:模型对象
// table: 表名，可选，若传入为空，则自动尝试从 obj.TableName() 读取
func MigrateWithTable(ctx context.Context, db DBCore, obj any, table string) error {
	if err := doMigrate(ctx, db, obj, table); err != nil {
		return fmt.Errorf("%T: %w", obj, err)
	}
	return nil
}

func doMigrate(ctx context.Context, db dbtype.DBCore, obj any, table string) error {
	if table == "" {
		if ht, ok := obj.(HasTable); ok {
			table = ht.TableName()
		} else {
			return errors.New("should implement HasTable interface")
		}
	}
	hd, ok := db.(HasDriver)
	if !ok {
		return errors.New("db does not implement HasDriver")
	}
	fy, err := dialect.Find(hd.Driver())
	if err != nil {
		return err
	}
	md, ok := fy.(dbtype.MigrateDialect)
	if !ok {
		return errors.New("db does not implement MigrateDialect")
	}

	schema, err := dbschema.Schema(fy, obj)
	if err != nil {
		return err
	}
	if table != "" {
		schema.Table = table
	}

	if schema.Table == "" {
		return fmt.Errorf("%T should implement HasTable interface", obj)
	}
	err = md.Migrate(ctx, db, *schema)
	return err
}
