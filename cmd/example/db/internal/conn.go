package internal

import (
	"database/sql"

	"github.com/xanygo/anygo/store/xredis"
)

func NewPostgres() (*sql.DB, error) {
	const dsn = `user=work password=123456 host=127.0.0.1 port=5432 database=demo sslmode=disable`
	return sql.Open("pgx", dsn)
}

func NewMySQL() (*sql.DB, error) {
	const dsn = `work:123456@tcp(127.0.0.1)/demo`
	return sql.Open("mysql", dsn)
}

func NewMSSQL() (*sql.DB, error) {
	const dsn = `sqlserver://sa:123456-Abc@localhost:1433?database=demo`
	return sql.Open("sqlserver", dsn)
}

func NewRedis() *xredis.Client {
	const uri = `redis://127.0.0.1:6379/10`
	_, client, err := xredis.NewClientByURI("demo", uri)
	if err != nil {
		panic(err)
	}
	return client
}
