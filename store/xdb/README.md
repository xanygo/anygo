
## service 配置文件段落
```toml
[Database]
Driver = "sqlite3"  # 可选，mysql
Username = "user"
Password = "psw"
DBName ="demo"

# DSN 该 Driver 对应的完整的 DSN,支持变量
DSN="{{.Username}}:{{.Password}}@{{.Network}}({{.HOST_PORT}})/{{.DBName}}?charset=utf8mb4,utf8" 
```

## 数据模型(Model)

```go
package dao

type User struct{
	ID       int64  `db:"id,primaryKey,autoIncr"`
	Name     string `db:"name"`
	Password string `db:"password"`
	Salt     string `db:"salt"`
}


type Admin struct{
	User             // 支持 Embed 类型
	Roles []string `db:"roles,codec:csv"`
}

func (a Admin)TableName()string{
	return "admin"
}

```

### Tag
  默认的 tag 名称为 `db`，可以使用 `SetTagName` 方法修改。
  格式为：
  ```
  db:"{数据库字段名}[,属性1][,属性2]"
  ```
  属性格式为 field:value  或者 field，如 
  ```
  ID int64 `db:"name,primaryKey,autoIncr"`
  
  ArticleIDs []int64  `db:"aids,codec:csv"`
  ```
支持属性如下：

| 名称         | 说明                                 | 示例                      |
|------------|------------------------------------|-------------------------|
| primaryKey | 主键                                 |                         |
| codec      | 对于复杂的类型，在写入数据库时编码，在查询出来后，解码        | codec:csv 或者 codec:json |
| autoIncr   | 标记此字段为数据库主键。Encode 时，若字段为零值，则忽略该字段 |                         |

### codec 参数
数据编解码的方式：

| 名称       | 说明                                             | 输出示例                  |
|----------|------------------------------------------------|-----------------------|
| csv      | csv 格式，支持 string、number、bool 类型的 slice 或 array | `a,b,c`               |
| json     | JSON格式， 可用于 slice、array 、struct、map 类型的字段      | `25`                  |
| text     | 编码为字符串                                         | `alice@example.com`   |
| date     | 可用于 time.Time 类型的字段                            | `2025-11-11 13:00:00` |
| dateTime | 可用于 time.Time 类型的字段                            | `2025-11-11 13:00:00` |
| timespan | 可用于 time.Time 类型的字段,数据库中存储的 int 类型的值           | `1234567890`          |


## 驱动
| 名称      | import path                    | 说明                  |
|---------|--------------------------------|---------------------|
| mysql   | github.com/go-sql-driver/mysql | 支持 MySQL 和 MariaDB  |
| sqlite3 | github.com/mattn/go-sqlite3    | 支持 sqlite3,需要 cGo=1 |