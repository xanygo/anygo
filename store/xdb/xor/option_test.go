//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-13

package xor

import (
	"database/sql"
	"testing"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/store/xdb/dbschema"
	"github.com/xanygo/anygo/store/xdb/dialect"
	"github.com/xanygo/anygo/xt"
)

func TestOrderByColumn(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		cfg := &config{}
		OrderByColumn().withOption(cfg)
		xt.Empty(t, cfg.orderBy)
	})
	t.Run("case 2", func(t *testing.T) {
		cfg := &config{}
		OrderByColumn().withOption(cfg)
		xt.Empty(t, cfg.orderBy)
	})
	fy := &dialect.SQLite3{}
	schema, err := dbschema.Schema(fy, testUser{})
	xt.NoError(t, err)
	xt.NotEmpty(t, schema)

	t.Run("case 3", func(t *testing.T) {
		cfg := &config{
			schema:  schema,
			dialect: fy,
		}
		OrderByColumn(sql.Named("id", true)).withOption(cfg)
		xt.Equal(t, cfg.orderBy, `"id" ASC`)
	})

	t.Run("case 4", func(t *testing.T) {
		cfg := &config{
			schema:  schema,
			dialect: fy,
		}
		OrderByColumn(sql.Named("id", true), sql.Named("name", "desc")).withOption(cfg)
		xt.Equal(t, cfg.orderBy, `"id" ASC,"name" DESC`)
		xt.NoError(t, cfg.getError())
	})
}

func testNewConfig() *config {
	fy := &dialect.SQLite3{}
	schema, err := dbschema.Schema(fy, testUser{})
	anygo.Must(err)
	return &config{
		schema:  schema,
		dialect: fy,
	}
}

func TestOrderByPkAsc(t *testing.T) {
	cfg := testNewConfig()

	OrderByPkAsc().withOption(cfg)
	xt.NoError(t, cfg.getError())
	xt.Equal(t, cfg.orderBy, `"id" ASC`)
}

func TestOrderByPkDesc(t *testing.T) {
	cfg := testNewConfig()

	OrderByPkDesc().withOption(cfg)
	xt.NoError(t, cfg.getError())
	xt.Equal(t, cfg.orderBy, `"id" DESC`)
}

func TestWhereByPK(t *testing.T) {
	cfg := testNewConfig()

	WhereByPK(testUser{ID: 2}).withOption(cfg)
	xt.NoError(t, cfg.getError())
	xt.Equal(t, cfg.where, `"id"=?`)
	xt.Equal(t, cfg.whereArgs, []any{int64(2)})
}

func TestWhereByMap(t *testing.T) {
	cfg := testNewConfig()

	WhereByMap(map[string]any{"id": 5}).withOption(cfg)
	xt.NoError(t, cfg.getError())
	xt.Equal(t, cfg.where, `"id"=?`)
	xt.Equal(t, cfg.whereArgs, []any{5})
}

type testUser struct {
	ID   int64  `db:"id,auto_inc,pk"`
	Name string `db:"name"`
}
