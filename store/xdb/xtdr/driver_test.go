//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xtdr_test

import (
	"database/sql"
	"testing"

	"github.com/xanygo/anygo/store/xdb/xtdr"
	"github.com/xanygo/anygo/xt"
)

func TestRegister(t *testing.T) {
	xtdr.Register()
	db, err := sql.Open(xtdr.Name, "")
	xt.NoError(t, err)
	xt.NotEmpty(t, db)

	t.Run("query 1", func(t *testing.T) {
		xtdr.ExpectQuery("select 1", []string{"id", "name"}, [][]any{{1, "hello"}})
		rows, err := db.Query("select 1")
		xt.NoError(t, err)
		xt.NotEmpty(t, rows)
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			xt.NoError(t, err)
			xt.Equal(t, 1, id)
			xt.Equal(t, "hello", name)
		}
		xt.NoError(t, rows.Err())
		xt.NoError(t, rows.Close())

		rows, err = db.Query("select 1")
		xt.Error(t, err)
		xt.Empty(t, rows)
	})

	t.Run("exec 1", func(t *testing.T) {
		xtdr.ExpectExec("delete 1", xtdr.ResultOf(3, 2), nil)
		ret, err := db.Exec("delete 1")
		xt.NoError(t, err)
		num1, err1 := ret.RowsAffected()
		xt.NoError(t, err1)
		xt.Equal(t, 2, num1)

		num2, err2 := ret.LastInsertId()
		xt.NoError(t, err2)
		xt.Equal(t, 3, num2)

		ret, err = db.Exec("delete 1")
		xt.Error(t, err)
		xt.Empty(t, ret)
	})
}
