# xkvx

## Loader (缓存对象加载器)

全局的 `Load` 方法，从配置文件 `{ConfDir}/xkx.[yml|json]` 可以加载一个支持泛型类型的缓存对象。

```json
{
    "Items":[
        {
            "Name":"kv1",       // 必填，名称，Load 方法里使用的 name 参数值
            "Type":"File",      // 必填，存储类型，使用本地文件存储数据
            "Dir":"{xattr.DataDir}/xkv/kv1",  // 必填, 缓存数据目录
        },
        {
            "Name":"kv2",
            "Type":"Memory"  // 必填，缓存类型。数据存储在进程的内存中
        },
        {
            "Name":"kv3",
            "Type":"Redis",     // 必填，存储类型，数据存储在 Redis 数据库中
            "Service":"rds",    // 必填，Redis 数据库的服务名称，对应服务配置一般在 {app}/conf/service/rds.yml 
        },
        {
            "Name":"kv4",
            "Type":"DB",          // 必填，存储类型，数据存储在数据库(如 sqlite、mysql、pgx 等)中
            "Service":"mysql1",   // 必填，数据库的服务名称，对应服务配置一般在 {app}/conf/service/mysql1.yml 
        },
        {
            "Name":"kv5",
            "Type":"Nop"   // 必填，缓存类型。黑洞。写入总是成功，读取总是不存在
        },
    ]
}
```