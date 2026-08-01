//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-08-11

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/microsoft/go-mssqldb" // mssql driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xlog"

	"cmd/example/db/model"
)

// Microsoft SQL server

func main() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())
	dsn := `sqlserver://sa:123456-Abc@localhost:1433?database=demo`
	log.Println("Using DSN:", dsn)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	client := xdb.NewClient("sqlserver", "demo", db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	model.WithUser(ctx, client)
}
