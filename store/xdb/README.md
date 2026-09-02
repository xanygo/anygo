# ORM

## service 配置文件段落
```toml
[Database]
Driver = "sqlite3"  # 可选，mysql 等
Username = "user"
Password = "psw"
DBName ="demo"

# DSN 该 Driver 对应的完整的 DSN,支持变量
DSN="{{.Username}}:{{.Password}}@{{.Network}}({{.HOST_PORT}})/{{.DBName}}?charset=utf8mb4,utf8" 
```

## 驱动
在使用 `xdb` 时，需要自行在自己应用的代码中注册对应的驱动。

| 驱动名称        | 别名     | import path                     | 说明                      |
|-----------|--------|---------------------------------|-------------------------|
| mysql     |        | github.com/go-sql-driver/mysql  | 支持 MySQL 和 MariaDB      |
| postgres  | pgx    | github.com/jackc/pgx/v5         | 支持 postgres             |
| sqlserver | mssql  | github.com/microsoft/go-mssqldb | 支持 Microsoft SQL server |
| sqlite3   | sqlite | github.com/mattn/go-sqlite3     | 支持 sqlite3, 需要 cGo=1    |
| sqlite3   | sqlite | github.com/ncruces/go-sqlite3/driver     | 支持 sqlite3, 不需要 cGo=1    |
| sqlite   |  | modernc.org/sqlite     | 支持 sqlite3, 不需要 cGo=1    |
| sqlite   |  | github.com/glebarez/go-sqlite     | 支持 sqlite3, 不需要 cGo=1    |
