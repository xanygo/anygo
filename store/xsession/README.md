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
            "Type":"Cookie",       // 必填，存储类型，session 数据存储在 Cookie 中。由于是在浏览器中，用户可见的，所以需要加密。
            "Life":"8760h",         // 可选，Cookie 有效期。默认为 365 天
            "CipherType":"AesOFB", // 可选，加密算法，支持 AesOFB、AesGCM、AesBlock，默认为 AesOFB
            "CipherKey":"",        // 必填，加密密钥 
            "CipherIV":"",         // 可选，加密向量
            "CookieName":"",       // 可选，存储数据的 cookie 名字，默认为 session
            "CookiePath":"",       // 可选，cookie 的保存路径，默认为 /
        },
         {
            "Name":"session3",
            "Type":"XCache",        // 必填，存储类型。使用 xcache 所配置的缓存实例
            "Ref":"cache1",         // 必填，在 store/xcache.xxx 配置文件中定义好的 cache 的名字
        },
    ]
}
```

正常情况下，一个应用配置一个即可。