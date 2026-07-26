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
	dsn := `work:123456@tcp(127.0.0.1)/demo`
	log.Println("Using DSN:", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	client := xdb.NewClient("mysql", "demo", db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	model.WithUser(ctx, client)
}
