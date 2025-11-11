
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
	ID int64 `db:"id"`
	Name string `db:"name"`
	Roles []string `db:"roles,codec:csv"`
}
```


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