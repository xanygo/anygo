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

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xlog"

	"db-example/model"
)

func main() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())
	db, err := sql.Open("mysql", "mss_test:mss_test_pass@tcp(127.0.0.1)/mss_test_db0")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	client := xdb.NewClient("mysql", "demo", db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	model.WithUser(ctx, client)
}
