//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xlog"

	"db-example/model"
)

func main() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())

	db, err := sql.Open("sqlite3", "./foo.db")

	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	model.WithUser(ctx, client)
}
