//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-13

package xdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/xanygo/anygo/store/xdb/dialect"
)

// Migrate 自动创建、添加字段（非生产环境使用）
func Migrate(db DBCore, obj any) error {
	return MigrateWithTable(db, obj, "")
}

func MigrateWithTable(db DBCore, obj any, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := doMigrate(ctx, db, obj, table); err != nil {
		return fmt.Errorf("%T: %w", obj, err)
	}
	return nil
}

func doMigrate(ctx context.Context, db DBCore, obj any, table string) error {
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

	schema, err := (schemaParser{}).Get(obj)
	if err != nil {
		return err
	}
	if table != "" {
		schema.Table = table
	}

	if table == "" {
		return fmt.Errorf("%T should implement HasTable interface", obj)
	}
	sqls, err := schema.AlterAddSQL(fy)
	if err != nil {
		return err
	}
	for _, sqlStr := range sqls {
		log.Println("sql:", sqlStr)
		// continue
		_, err = db.ExecContext(ctx, sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}
