# xcache

## Loader (缓存对象加载器)

全局的 `Load` 方法，从配置文件 `{ConfDir}/store/xcache.[yml|json]` 可以加载一个支持泛型类型的缓存对象。

```json5
{
    "Items":[
        {
            "Name":"cache1",     // 必填，名称，Load 方法里使用的 name 参数值
            "Type":"File",       // 必填，缓存类型，使用本地文件存储数据
            "Dir":"{xattr.DataDir}/filecache/cache1",  // 必填参数, 缓存数据目录
            "Codec":{    // 可选。数据编码方式，默认为 json
                "Type":"JSON"
            },     
            "Capacity":12345,    // 可选，最大缓存个数(非严格限定)，在清理时，会按照创建时间删除多余的
            "Life":{             // 可选，缓存有效期，所有类型的缓存都可以设置，详见下文
                "Min":"1h",
                "Max":"2d",
                "Force":"1h"
            },
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
            "KeyPrefix":"user_", // 可选，缓存前缀
            "Codec":{              // 可选。数据编码方式，默认为 JSONV2
                "Type":"JSON",     // 可选值，JSON，JSONV2
            }       
        },
        {
            "Name":"cache6",
            "Type":"DB",          // 必填，缓存类型，数据存储在数据库(如 sqlite、mysql、pgx 等)中
            "Service":"mysql1",   // 必填，数据库的服务名称，对应服务配置一般在 {app}/conf/service/mysql1.yml 
            "KeyPrefix":"user_", // 可选，缓存前缀
            "Table":"xcache",    // 可选，缓存数据的表名，默认为 xcache
            "Capacity":12345,     // 可选, 缓存容量大小(非严格限定)
            "GC":"30s",           // 可选，自动清理任务的运行周期，默认 60s
            "BGTimeout":"30s",    // 可选，后台任务的超时时间
            "Codec":{             // 可选。数据编码方式，默认为 json
                "Type":"JSON", 
            }      
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
                    "Life":{             // 必填，缓存有效期
                        "Default":"1h",
                        "Min":"1h",
                        "Max":"2d",
                        "Force":"1h"
                    },
                    "WriteTimeout":"3s"  // 可选，异步写超时时间
                },
                {
                    "Ref":"cache6",      // 必填，引用的数据库名称，在此配置中已经定义好的
                    "Life":{             // 必填，缓存有效期
                        "Default":"1h",
                        "Min":"1h",
                        "Max":"1d",
                        "Force":"2d"
                    },
                    "WriteTimeout":"3s"  // 可选
                }
            ]
        },
        {
            "Name":"cache9",
            "Type":"Wrap",            // 必填，缓存类型。链式多级缓存
            "Ref":"cache1",           // 必填，引用的数据库名称，在此配置中已经定义好的
            "Life":{                  // 可选，缓存有效期，所有类型的缓存都可以设置，详见下文
                "Min":"1h",
                "Max":"2d",
                "Force":"1h"
            },
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

### Type

目前的 `Type` 已支持 `File`,`MemoryLRU`,`MemoryFIFO`,`MemoryLIFO`,`Redis`,`DB`,`Nop`,`Chains`,`Wrap` 这些。

### Life

所有的 类型(`Type`)都支持配置 `Life` 来强制干预缓存的有效期：
```json5
 {
    "Type":"XXX",
    "Life":{
        "Default":"1h",      // 当调用 API 时，传入的 life=0 时生效，第一个判断使用
        "Min":"1h",          // 值 >0 时，若调用 API 时传入的值小于此值，则使用此值
        "Max":"2d",          // 值 >0 时，若调用 API 时传入的值大于此值，则使用此值
        "Force":"1h"         // 值 >0 时，覆盖掉调用 API 时传入的的值
    }
}
```
有效期采用 StringDuration 类型来表示，如：

    1. 30d    -> 30 天 ( 30*24 小时)
    2. 1h     -> 1 小时
    3. 1d1h   -> 25 小时
    4. 2h5m   -> 2 小时 5 分钟
    5. 1800s  -> 1800 秒

当配置的时间长度值 >0 时生效，优先级: `Default`  > `Force` > `Min` 和 `Max` 。

### Codec

由于底层存储 value 数据的类型是 `[]byte`, 上层对应使用的时候使用的是泛型,即缓存对象是任意类型，
所以在存储的时候，需要使用 `Codec` 将数据编码为 `[]byte`,在读取的时候，将 `[]byte` 解码为对象。

`Codec` 内部还可以配置 `Cipher` 来实现对数据的加密、压缩、编码。 
`Cipher` 可以配置一个对象或者数组。具体如下：

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
`Cipher` 里可以有 N>=0 个配置项。可以是: `AesGCM`、`AesGCM` + `GZip`、`AesGCM` + `Base64` 等组合方式。


Cipher `Type`: 数据处理算法名称，可选值 No，AesOFB，AesGCM，AesBlock, GZip, Base64, Base62, Base58，Base36