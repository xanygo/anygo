# CookieJar


## Loader (缓存对象加载器)

全局的 `Load` 方法，从配置文件 `{ConfDir}/store/xcookiejar.[yml|json|toml]` 可以加载一个 `CookieJar` 对象。

```json5
{
    "Items":[
        {
            "Name":"jar1",        // 必填，名称，Load 方法里使用的 name 参数值
            "Type":"Memory",      // 必填，存储类型，使用内存
            "PSList":"default",   // 可选，PublicSuffixList 策略名称，需要提前使用 RegisterPSList 注册 
        },
        {
            "Name":"jar2",
            "Type":"No"           // 必填，存储类型。Load() 会返回 nil
        },
        {
            "Name":"jar3",
            "Type":"Nop"          // 必填，存储类型。数据写入黑洞，写入后，读取总是为空
        },
        {
            "Name":"jar4",
            "Type":"DB",           // 必填，存储类型，数据存储在数据库(如 sqlite、mysql、pgx 等)中
            "Service":"mysql1",    // 必填，数据库的服务名称，对应服务配置一般在 {app}/conf/service/mysql1.yml
            "Table":"xcookiejar",  // 可选，数据库表明，默认为 xcookiejar
            "AutoMigrate":true,    // 可选，是否自动创建表结构。生产环境配置为 false 或者不配置
            "PSList":"default",    // 可选，PublicSuffixList 策略名称，需要提前使用 RegisterPSList 注册 
        },
        {
            "Name":"jar5",
            "Type":"KV",               // 必填，存储类型，使用 xkv
            "Ref":"kv1",               // 必填，在 `{ConfDir}/store/xkv.[yml|json|toml]` 中定义的 kv 名称
            "KeyPrefix":"prefix",      // 可选，存储数据的 key的前缀，默认为空

            // MetaKey 可选，存储 cookiejar 的 key 列表的字段名，默认为 xcookiejar-keys
            // 存储 key 列表后，Jar 会在后台定期遍历所有key，删除过期的 Cookie
            // 若值为 no，则不存储 key 列表
            "MetaKey":"xcookiejar-keys", 
            "PSList":"default",        // 可选，PublicSuffixList 策略名称，需要提前使用 RegisterPSList 注册 
        },
    ]
}
```

## PSList
用于判断一个域名的 Public Suffix（公共后缀）以及注册域名，在 CookieJar 中用于判断 Cookie 的域名边界是否正确。 

```
www.example.com
      │       │
      │       └── public suffix = com
      └────────── registrable domain = example.com
```

可以 `import "golang.org/x/net/publicsuffix"` ,数据来源是 Mozilla 的 Public Suffix List(PSL) 。

在使用 `xcookiejar` 之前，应当使用 `RegisterPSList` 注册：
```go
package main

import "golang.org/x/net/publicsuffix"

func init(){
    // 注册后，可以在配置中使用  "PSList":"default", 引用该策略
    xcookiejar.RegisterPSList("default",publicsuffix.List)
}
```