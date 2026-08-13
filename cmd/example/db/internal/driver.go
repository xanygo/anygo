package internal

import (
	_ "github.com/go-sql-driver/mysql"  // mysql driver
	_ "github.com/jackc/pgx/v5/stdlib"  // pgx driver
	_ "github.com/mattn/go-sqlite3"     // sqlite driver
	_ "github.com/microsoft/go-mssqldb" // mssql driver
)

func Init() {}
