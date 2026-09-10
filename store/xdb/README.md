## service 配置文件

### toml 格式
```toml
# conf/service/kvdb.toml
Name = "kvdb"

# ConnectTime = 1000        # 创建网络连接的超时时间，可选，单位 ms，默认值为 5 秒
# ReadTimeout = 1000        # socket Read 超时时间，可选，单位 ms，默认值为 10 秒
# WriteTimeout = 1000       # socket Write 超时时间，可选，单位 ms，默认值为 5 秒

[Database]
Driver = "mysql"  # 必填，驱动类型,可选值，mysql, pgx, mssql, sqlite3
Username = "user" # 可选，用户名, 也可以直接写在 DSN
Password = "psw"  # 可选，密码，也可以直接写在 DSN
DBName ="demo"    # 可选，数据库名，也可以直接写在 DSN

# 下面几个是配置 sql.DB 的连接池参数的，可选参数，含义详见 sql.DB 的文档
# MaxOpenConns=1
# MaxIdleConns=1
# ConnMaxLifeTime="60s"
# ConnMaxIdleTime="60s"

# DSN，必填，该 Driver 对应的完整的 DSN,支持变量
# 变量 Username、Password、DBName 均为上述配置内容
# 变量 Network、HOST_PORT 由框架补充得到
DSN="{{.Username}}:{{.Password}}@{{.Network}}({{.HOST_PORT}})/{{.DBName}}?charset=utf8mb4,utf8"

[DownStream]
  Address = ["dummy:80"]
```

### yaml 格式
```yaml
Name : kvdb

Database:
  Driver: sqlite3 
  DSN: {xattr.RelDataDir|store/aimux.db}

DownStream:
  Address:
    - 'dummy:80'
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
