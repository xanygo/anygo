
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

import "time"

type User struct {
  ID       int64  `db:"id,pk,auto_inc"` // 数据库字段名-id，主键、自增长
  Name     string `db:"name"`
  Password string `db:"password"`
  Salt     string `db:"salt,not-null"` // 数据库字段名-salt, NOT NULL, 默认值空字符串
  Created  time.Time `db:"create_time,auto=Created"` // 该条数据创建时间
  Updated  time.Time `db:"update_time,auto=Updated"` // 该条数据更新时间
}

// "not-null" 可以不写，是默认的，若允许 NULL，可以添加 "null" 属性

type Admin struct {
  User           // 支持 Embed 类型
  Roles []string `db:"roles,codec=csv,default"` // 数据库字段名-roles,数据编解码器：csv, 默认值空字符串
}

func (a Admin) TableName() string {
  return "admin" // 数据库表名，admin
}

```
### 自动赋值字段
1. tag 定义的 `auto`值为 `Created`、`CreatedUnix`（是 `Created` 的别名）、`CreatedNano` 的字段被认为是数据的创建字段，当类型是 `time.Time` 或者 `int64` 类型的时候：
```
  Created time.time `db:"created_at,auto=Created"`   // 赋值 time.Now()
  Created int64 `db:"created_at,auto=Created"`       // 赋值 time.Now().Unix()
```
在 insert 的时候，若值为零值，则自动赋值当前时间。

| auto 值      | 字段类型      | 赋值                    |
|-------------|-----------|-----------------------|
| Created     | time.Time | time.Now()            |
| Created     | int64     | time.Now().Unix()     |
| CreatedNano | time.Time | time.Now()            |
| CreatedNano | int64     | time.Now().UnixNano() |



2. tag 定义的 `auto`值为 `Updated`、`UpdatedUnix`（是 `Updated` 的别名）、`UpdatedNano` 的字段被认为是数据的创建字段，当类型是 `time.Time` 或者 `int64` 类型的时候：
```
  Updated time.time `db:"updated_at,auto=Updated"`   // 赋值 time.Now()
  Updated int64 `db:"updated_at,auto=Updated"`       // 赋值 time.Now().Unix()
```
在 insert 或者 update 的时候，自动赋值当前时间。

| auto 值      | 字段类型      | 赋值                    |
|-------------|-----------|-----------------------|
| Updated     | time.Time | time.Now()            |
| Updated     | int64     | time.Now().Unix()     |
| UpdatedNano | time.Time | time.Now()            |
| UpdatedNano | int64     | time.Now().UnixNano() |

### Tag
  默认的 tag 名称为 `db`，可以使用 `SetTagName` 方法修改。
  格式为：
  ```
  db:"{数据库字段名}[,属性1][,属性2]"
  ```
  属性格式为 field:value  或者 field，如 
  ```
  ID int64 `db:"name,pk,auto_inc"`
  
  ArticleIDs []int64  `db:"aids,codec=csv"`
  ```
支持属性如下：

| 名称           | 说明                                         | 示例                      |
|--------------|--------------------------------------------|-------------------------|
| pk           | 主键，也可以写作 primaryKey。允许在多个字段定义 pk 属性（联合主键）。 |                         |
| codec        | 对于复杂的类型，在写入数据库时编码，在查询出来后，解码                | codec=csv 或者 codec=json |
| auto_inc     | 标记此字段为数据库主键。Encode 时，若字段为零值，则忽略该字段         |                         |
| uniq         | 唯一键，不需要值，也可以是完整的 unique，Migrate 时使用        | uniq                    |
| index        | 索引，Migrate 时使用                             | 详见下文                    |
| unique_index | 唯一索引，Migrate 时使用                           | 格式同 index               |
| size         | 值类型的容量, String 类型的时候有用，Migrate 时使用         | size:255                |
| not-null     | 不允许存储 NULL，Not Null (默认)，Migrate 时使用       |                         |
| null         | 允许存储 NULL，Migrate 时使用                      |                         |
| default      | 默认值，Migrate 时使用                            | 详见下文                    |


#### index/uniqueIndex
index 示例： 
  1. index                 -> 创建独立索引，索引名称为 idx_字段名
  2. index=idx_uid         -> 创建独立索引，索引名称为 idx_uid
  3. index=idx_uid_class,1   -> 创建联合索引，索引名称为 idx_uid_class，此字段在索引中排序为 1

uniqueIndex 示例：
  1. unique_index                         -> 创建独立索引，索引名称为 idx_uniq_字段名
  2. unique_index=idx_uniq_uid            -> 创建独立索引，索引名称为 idx_uniq_uid
  3. unique_index=idx_uniq_uid_class,1    -> 创建联合索引，索引名称为 idx_uniq_uid_class，此字段在索引中排序为 1

#### default
格式为 `default:[[fn|string|number]|]value`。只在 Migrate 时使用，Encode 时不会使用

示例：
  1. 默认值为空字符串：“name,default”
  2. 默认值为数字：“name,default=number|123”
  3. 默认值为字符串：“name,default=string|hello”
  4. 默认值为数据库函数：“name,default=fn|CURRENT_TIMESTAMP”


#### native
设置数据库中字段类型使用数据库原生类型，如 `native:varchar(32)`

### codec 参数
数据编解码的方式：

| 名称         | 说明                                             | 输出示例                  |
|------------|------------------------------------------------|-----------------------|
| csv        | csv 格式，支持 string、number、bool 类型的 slice 或 array | `a,b,c`               |
| json       | JSON 格式， 可用于 slice、array 、struct、map 类型的字段     | `25`                  |
| auto_json  | 需要数据库方言来判断类型，若方言判断不出来，则默认使用 json 编解码           |                       |
| text       | 编码为字符串                                         | `alice@example.com`   |
| date       | 可用于 time.Time 类型的字段                            | `2025-11-11 13:00:00` |
| date_time  | 可用于 time.Time 类型的字段                            | `2025-11-11 13:00:00` |
| timespan   | 可用于 time.Time 类型的字段,数据库中存储的 int 类型的值           | `1234567890`          |

通过 codec 参数指定复杂类型在编码为 SQL 语句时的序列化方式，以及从数据库中读取出来后反序列化的方式。
除了上述内置的 codec，还可以通过 dbcodec.Register 注册自定义的 codec。

`auto_json` 可以这样用：
```
Scores       []int     `db:"scores,codec=auto_json"`
```
对于数据库引擎支持数组的，如 pgx，其方言会依据数据类型做出自动编码。
对于不支持数组的，如 sqlite, 会退化为 json 编码，数据库字段类型时 Text 类型。


## 驱动
| 名称        | 别名     | import path                     | 说明                      |
|-----------|--------|---------------------------------|-------------------------|
| mysql     |        | github.com/go-sql-driver/mysql  | 支持 MySQL 和 MariaDB      |
| sqlite3   | sqlite | github.com/mattn/go-sqlite3     | 支持 sqlite3, 需要 cGo=1    |
| postgres  | pgx    | github.com/jackc/pgx/v5         | 支持 postgres             |
| sqlserver | mssql  | github.com/microsoft/go-mssqldb | 支持 Microsoft SQL server |
