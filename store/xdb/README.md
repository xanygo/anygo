
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
	ID       int64  `db:"id,pk,autoInc"`  // 数据库字段名-id，主键、自增长
	Name     string `db:"name"`
	Password string `db:"password"`
	Salt     string `db:"salt,notNull,default"` // 数据库字段名-salt, NOT NULL, 默认值空字符串
}


type Admin struct{
	User             // 支持 Embed 类型
	Roles []string `db:"roles,codec:csv,default"`  // 数据库字段名-roles,数据编解码器：csv, 默认值空字符串
}

func (a Admin)TableName()string{
	return "admin"                    // 数据库表名，admin
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
  ID int64 `db:"name,primaryKey,autoInc"`
  
  ArticleIDs []int64  `db:"aids,codec:csv"`
  ```
支持属性如下：

| 名称          | 说明                                  | 示例                      |
|-------------|-------------------------------------|-------------------------|
| pk          | 主键，也可以写作 primaryKey                 |                         |
| codec       | 对于复杂的类型，在写入数据库时编码，在查询出来后，解码         | codec:csv 或者 codec:json |
| autoInc     | 标记此字段为数据库主键。Encode 时，若字段为零值，则忽略该字段  |                         |
| uniq        | 唯一键，不需要值，也可以是完整的 unique，Migrate 时使用 | uniq                    |
| index       | 索引，Migrate 时使用                      | 详见下文                    |
| uniqueIndex | 唯一索引，Migrate 时使用                    | 格式同 index               |
| size        | 值类型的容量, String 类型的时候有用，Migrate 时使用  | size:255                |
| notNull     | Not Null，Migrate 时使用                |                         |
| default     | 默认值，Migrate 时使用                     | 详见下文                    |


#### index/uniqueIndex
index 示例： 
  1. index                 -> 创建独立索引，索引名称为 idx_字段名
  2. index:idx_uid         -> 创建独立索引，索引名称为 idx_uid
  3. iex:idx_uid_class,1   -> 创建联合索引，索引名称为 idx_uid_class，此字段在索引中排序为 1

uniqueIndex 示例：
  1. uniqueIndex                         -> 创建独立索引，索引名称为 idx_uniq_字段名
  2. uniqueIndex:idx_uniq_uid            -> 创建独立索引，索引名称为 idx_uniq_uid
  3. uniqueIndex:idx_uniq_uid_class,1    -> 创建联合索引，索引名称为 idx_uniq_uid_class，此字段在索引中排序为 1

#### default
格式为 `default:[[fn|string|number]|]value`。只在 Migrate 时使用，Encode 时不会使用

示例：
  1. 默认值为空字符串：“name,default”
  2. 默认值为数字：“name,default:number|123”
  3. 默认值为字符串：“name,default:string|hello”
  4. 默认值为数据库函数：“name,default:fn|CURRENT_TIMESTAMP”


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