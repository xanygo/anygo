//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-07

package xtdr

import "database/sql"

func MustOpen() *sql.DB {
	db, err := sql.Open(Name, "test-db")
	if err != nil {
		panic(err)
	}
	return db
}
