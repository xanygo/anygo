package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xkv/xkvx"
	"github.com/xanygo/anygo/xlog"
)

func main() {
	xdb.RegisterIT((&xdb.Logger{Logger: xlog.NewSimple(os.Stderr)}).ToInterceptor())

	db, err := sql.Open("sqlite3", "./foo.db")

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()
	client := xdb.NewClient("sqlite3", "demo", db)

	kv := &xkvx.DatabaseStorage{
		DB: client,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err = kv.Migrate(ctx); err != nil {
		log.Fatalln("Migrate:", err)
	}

	ks := kv.String("hello")
	err = ks.Set(ctx, "world")
	log.Println("Set:", err)

	val, found, err := ks.Get(ctx)
	log.Println("Get:", val, found, err)
}
