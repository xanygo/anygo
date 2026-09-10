# HTTP Session 存储能力

## LoadStorageFunc

全局的 `LoadStorageFunc` 方法，从配置文件 `{ConfDir}/store/xsession.[yml|json|toml]` 可以加载一个支持泛型类型的缓存对象。

```json
{
    "Items":[
        {
            "Name":"session1",   // 必填，LoadStorageFunc 所需要的名称
            "Type":"XKV",        // 必填，存储类型，使用 xkv 所配置的存储实例
            "Ref":"xkv1",        // 必填，在 store/xkv.xxx 配置文件中定义好的 kv 存储的名字
        },
        {
            "Name":"session2",
            "Type":"Cookie",        // 必填，存储类型，session 数据存储在 Cookie 中。由于是在浏览器中，用户可见的，所以需要加密。
            "Life":"8760h",         // 可选，Cookie 有效期。默认为 365 天。
            "Cipher":{
                "Type":"AesOFB",     // 可选，加密算法，支持 AesOFB、AesGCM、AesBlock，默认为 AesOFB
                "Key":"hello-world", // 必填，加密密钥 
            }
            "CookieName":"",        // 可选，存储数据的 cookie 名字，默认为 session
            "CookiePath":"",        // 可选，cookie 的保存路径，默认为 /
        },
         {
            "Name":"session3",
            "Type":"XCache",        // 必填，存储类型。使用 xcache 所配置的缓存实例
            "Ref":"cache1",         // 必填，在 store/xcache.xxx 配置文件中定义好的 cache 的名字
        },
    ]
}
```

正常情况下，一个应用只需要在 Items 中配置一个。

`Cipher` 可以配置一个对象或者数组。具体如下：

```json5
{
    "Cipher":{              
        "Type":"AesGCM",     // 必填，加密算法
        "Key":"hello-world", // 必填，加密密钥
    }
}
```
或者：
```json5
{
    "Cipher":[        
        {               
            "Type":"AesGCM",     // 必填，加密算法。使用 AesGCM 加密 json encode 后的数据
            "Key":"hello-world", // 必填，加密密钥
        },
        {               
            "Type":"GZip",       // 必填，压缩算法。对加密后的数据压缩
        },
    ]
}
```

Cipher `Type`: 数据处理算法名称，可选值 No，AesOFB，AesGCM，AesBlock, GZip。
采用 `Cookie` 存储时，session 数据自动会添加 Base64 编码处理，故在 `Cipher` 中不需要配置 `Base64` 等编码器。