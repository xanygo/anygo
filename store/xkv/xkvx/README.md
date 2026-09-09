# xkvx

## Loader (缓存对象加载器)

全局的 `Load` 方法，从配置文件 `{ConfDir}/store/xkx.[yml|json|toml]` 可以加载一个支持泛型类型的缓存对象。

```json5
{
    "Items":[
        {
            "Name":"kv1",                    // 必填，名称，Load 方法里使用的 name 参数值
            "Type":"File",                   // 必填，存储类型，使用本地文件存储数据
            "Dir":"{xattr.DataDir}/xkv/kv1"  // 必填, 缓存数据目录
        },
        {
            "Name":"kv2",
            "Type":"Memory"  // 必填，缓存类型。数据存储在进程的内存中
        },
        {
            "Name":"kv3",
            "Type":"Redis",       // 必填，存储类型，数据存储在 Redis 数据库中
            "Service":"rds",      // 必填，Redis 数据库的服务名称，对应服务配置一般在 {app}/conf/service/rds.yml 
            "KeyPrefix":"prefix"  // 可选，key 的前缀
        },
        {
            "Name":"kv4",
            "Type":"DB",           // 必填，存储类型，数据存储在数据库(如 sqlite、mysql、pgx 等)中
            "Service":"mysql1",    // 必填，数据库的服务名称，对应服务配置一般在 {app}/conf/service/mysql1.yml
            "KeyPrefix":"prefix",  // 可选，key 的前缀
            "AutoMigrate":true,    // 可选，是否自动创建表结构。生产环境配置为 false 或者不配置
            "Codec":{
                "Type":"JSON",     // 必填，可选值 JSON、JSONV2 
                "Cipher":{               // 可选，用于对编码后的数据数据加密
                    "Type":"AesGCM",     // 必填，加密算法，可选值：No，AesOFB 等
                    "Key":"hello-world", // 必填，加密密钥
                }
            }
        },
        {
            "Name":"kv5",
            "Type":"Nop"           // 必填，缓存类型。黑洞。写入总是成功，读取总是不存在
        },
    ]
}
```

### Cipher

`Type`: 数据处理算法名称，可选值 No，AesOFB，AesGCM，AesBlock, GZip, Base64, Base62, Base58，Base36

### Codec 的补充

`Codec` 内部可以配置 `Cipher` 来实现对数据的加密、压缩、编码。 `Cipher` 可以配置一个对象或者数组。具体如下：

```json5
{
    "Codec":{
        "Type":"JSON",     // 必填，可选值 JSON、JSONV2 
        "Cipher":{               // 可选，用于对编码后的数据数据加密
            "Type":"AesGCM",     // 必填，加密算法
            "Key":"hello-world", // 必填，加密密钥
        }
    }
}
```

多个对象（先加密，然后压缩、编码）：
```json5
{
    "Codec":{
        "Type":"JSON",     // 必填，可选值 JSON、JSONV2 
        "Cipher":[         // 可选，用于对编码后的数据数据加密
            {               
                "Type":"AesGCM",     // 必填，加密算法。使用 AesGCM 加密 json encode 后的数据
                "Key":"hello-world", // 必填，加密密钥
            },
            {               
                "Type":"GZip",       // 必填，压缩算法。对加密后的数据压缩
            },
            {               
                "Type":"Base64",     // 必填，编码算法。对压缩后的数据编码
            },
        ]
    }
}
```
`Cipher` 里可以有 N>=0 个配置项。可以是: `AesGCM`、`AesGCM` + `GZip`、`AesGCM` + `Base64` 等组合方式