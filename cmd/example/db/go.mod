module db-example

go 1.26.1

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/jackc/pgx/v5 v5.9.1
	github.com/mattn/go-sqlite3 v1.14.38
	github.com/xanygo/anygo v0.0.0-20260401064044-5d1de4c0a27b
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/xanygo/anygo => ../../../
