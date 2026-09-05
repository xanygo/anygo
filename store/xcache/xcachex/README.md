# xcache

## Loader (缓存对象加载器)

全局的 `Load` 方法，从配置文件 `{ConfDir}/xcache.[yml|json]` 可以加载一个支持泛型类型的缓存对象。

```json
{
    "Items":[
        {
            "Name":"cache1",   // 必填，名称，Load 方法里使用的 name 参数值
            "Type":"File",     // 必填，缓存类型，使用本地文件存储数据
            "Dir":"{xattr.DataDir}/filecache/cache1",  // 必填参数, 缓存数据目录
            "Codec":"json",    // 可选。数据编码方式，默认为 json
            "Capacity":12345   // 可选，最大缓存个数(非严格限定)，在清理时，会按照创建时间删除多余的
        },
        {
            "Name":"cache2",    // 必填，名称
            "Type":"MemoryLRU", // 必填，缓存类型，数据存储在内存，按照使用频率，淘汰最早使用的
            "Capacity":12345    // 必填，缓存容量大小
        },
        {
            "Name":"cache3",      // 必填，名称
            "Type":"MemoryFIFO",  // 必填，缓存类型，数据存储在内存，按照缓存创建时间，先入先出
            "Capacity":12345      // 必填，缓存容量大小
        },
        {
            "Name":"cache4",      // 必填，名称
            "Type":"MemoryLIFO",  // 必填，缓存类型，数据存储在内存，按照缓存创建时间，后入先出
            "Capacity":12345      // 必填，缓存容量大小
        },
        { 
            "Name":"cache5",      // 必填，名称
            "Type":"Redis",       // 必填，缓存类型，数据存储在 Redis 数据库中
            "Service":"rds",      // 必填，Redis 数据库的服务名称，对应服务配置一般在 {app}/conf/service/rds.yml 
            "KeyPrefix":"user_"， // 可选，缓存前缀
            "Codec":"json"        // 可选。数据编码方式，默认为 json
        },
        {
            "Name":"cache6",
            "Type":"DB",          // 必填，缓存类型，数据存储在数据库(如 sqlite、mysql、pgx 等)中
            "Service":"mysql1",   // 必填，数据库的服务名称，对应服务配置一般在 {app}/conf/service/mysql1.yml 
            "KeyPrefix":"user_"， // 可选，缓存前缀
            "Table":"xcache"，    // 可选，缓存数据的表名，默认为 xcache
            "Capacity":12345,     // 可选, 缓存容量大小(非严格限定)
            "GC":"30s",           // 可选，自动清理任务的运行周期，默认 60s
            "BGTimeout":"30s",    // 可选，后台任务的超时时间
            "Codec":"json"        // 可选。数据编码方式，默认为 json
        },
        {
            "Name":"cache7", 
            "Type":"Nop"     // 必填，缓存类型。黑洞。写入总是成功，读取总是不存在
        },
        {
            "Name":"cache8",
            "Type":"Chains",          // 必填，缓存类型。链式多级缓存
            "Chains":[                // 必填。应包含 >=1 个有效值
                {
                    "Ref":"cache2",      // 必填，引用的数据库名称，在此配置中已经定义好的
                    "Life":"1800s",      // 必填，缓存有效期
                    "WriteTimeout":"3s"  // 可选，异步写超时时间
                },
                {
                    "Ref":"cache6",      // 必填，引用的数据库名称，在此配置中已经定义好的
                    "Life":"3600s",      // 必填，但是最后一个对象，此值不用
                    "WriteTimeout":"3s"  // 可选
                }
            ]
        },
        {
            "Name":"cache9",
            "Type":"Wrap",            // 必填，缓存类型。链式多级缓存
            "Ref":"cache1",           // 必填，引用的数据库名称，在此配置中已经定义好的
            "Life":"1800s",           // 可选，强制设置的缓存有效期
            "KeyTransform":{          // 可选，对缓存的 key 做变换处理
                "string":{            // 可选，对于 key 的类型是 string 的调用，可以添加前缀和后缀
                    "Prefix":"prefix_",  // 可选，给 key 添加前缀
                    "Suffix":"_suffix"   // 可选，给 key 添加后缀 
                },
                "Default":{           // 可选，对于没有找到的情况。
                    "Refuse":true,     // 可选，拒绝。让 Cache 调用报错
                    "Panic":true      // 可选，拒绝。让 Cache 调用 panic，在 Refuse 前判断
                }
            }
        }
    ]
}
```

目前的 `Type` 以支持 `File`,`MemoryLRU`,`MemoryFIFO`,`MemoryLIFO`,`Redis`,`DB`,`Nop`,`Chains`,`Wrap` 这些。