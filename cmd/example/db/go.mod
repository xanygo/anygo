module db-example

go 1.26.0

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/jackc/pgx/v5 v5.8.0
	github.com/mattn/go-sqlite3 v1.14.34
	github.com/xanygo/anygo v0.0.0-20260226120632-ddf85a12db03
)

require (
	filippo.io/edwards25519 v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace github.com/xanygo/anygo => ../../../
