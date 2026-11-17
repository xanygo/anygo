# xvld

通用的 Go 字段规则校验器。

## 内置规则

| 名称         | 功能                    | 支持的数据类型     | 举例                           |
| ------------ | ----------------------- | ------------------ | ------------------------------ |
| `required`   | 内容不为空              | 所有               | `required`                     |
| `min`        | 最小值                  | string/number      | `min=1`                        |
| `max`        | 最大值                  | string/number      | `max=10`                       |
| `len`        | 长度等于                | string/bytes       | `len=3`                        |
| `min-len`    | 最小长度                | string/bytes       | `min-len=3`                    |
| `max-len`    | 最小长度                | string/bytes       | `max-len=10`                   |
| `eq`         | 等于                    | string/bool/number | `eq=10`, `eq=hello`            |
| `ne`         | 不等于                  | string/bool/number | `ne=10`(不等于10), `ne=a\|b`(不等于 a 和b ) |
| `gt`         | 大于                    | number             | `gt=10`                        |
| `gte`        | 大于等于                | number             | `gte=10`                       |
| `lt`         | 小于                    | number             | `lt=10`                        |
| `lte`        | 小于等于                | number             | `lte=10`                       |
| `contains`   | 包含                    | string/bytes       | `contains=10`,`contains=hello` |
| `excludes`   | 不包含                  | string/bytes       | `excludes=10`,`excludes=hello` |
| `prefix`     | 包含前缀                | string/bytes       | `prefix=hello`                 |
| `suffix`     | 包含后缀                | string/bytes       | `suffix=world`                 |
| `one-of`     | 是枚举值中的一个        | string/number/bool | `one-of=1\|2\|3`               |
| `email`      | 邮箱地址                | string             | `email`                        |
| `url`        | URL 地址                | string             | `url`                          |
| `domain`     | 域名                    | string             | `domain`                       |
| `port`       | 端口号                  | string/int/uint    | `port`                         |
| `ip`         | IP 地址                 | string/bytes             | `ip`                           |
| `ipv4`       | IPV4 地址               | string/bytes    | `ipv4`                         |
| `ipv6`       | IPV6 地址               | string/bytes             | `ipv6`                         |
| `ident`      | 变量标识符              | string             | `ident`                        |
| `alpha`      | 英文字母                | string             | `alpha`                        |
| `alpha-num`  | 英文字母和数字          | string             | `alpha-num`                    |
| `uuid`       | 是 UUID                 | string             | `uuid`                         |
| `numeric`    | 可运算的数值            | string             | `numeric`                      |
| `number`     | 数字、数字字符串        | string             | `number`                       |
| `uint`       | uint 数字字符串         | string             | `uint`                         |
| `nz-uint`    | 非0 uint 字符串         | string/int/uint    | `nz-uint`                      |
| `hex-color`  | 16进制颜色              | string             | `hex-color` ，如 "#1e1f22"     |
| `hex`        | 16进制                  | string             | `hex` ，如 “1e1f22"            |
| `base36`     | Base36 字符串           | string             | `base36`                       |
| `base58`     | Base58 字符串           | string             | `base58`                       |
| `base62`     | Base62 字符串           | string             | `base62`                       |
| `base64-url` | Base64 URL 字符串       | string             | `base64-url`                   |
| `base64-std` | Base64 STD 字符串       | string             | `base64-std`                   |
| `md4`        | 长度为 32 的 md4 的值   | string             | `md4`                          |
| `md5`        | 长度为 32 的 md5 的值   | string             | `md5`                          |
| `sha256`     | sha256 的值，长度为 64  | string             | `sha256`                       |
| `sha384`     | sha384 的值，长度为 96  | string             | `sha384`                       |
| `sha512`     | sha256 的值，长度为 128 | string             | `sha512`                       |
| `regexp`     | 自定义正则表达式        | string             | `regexp=\d+`                   |
| `dive`       | 后面的规则校验子条目    | array/slice/map    | `dive,min=1`                   |

注：
1. `number` : 包含 int/uint/float。
2. 值中包含英文逗号时:可以使用在 Key 后面附加 `({Byte})` 来定义特殊的截至符,如:
   - `regexp(#)=\d{3,5}#`
   - `required,regexp(#)=\d{3,5}#,min=120`
3. 递归：上述所有规则，当类型是 array/slice/map 时，均支持递归建议其子元素:
    - `array/slice` : 如 `Numbers []int "min=10"` 当 `len > 0` 时，会校验每一个条目是否满足 `min=10` 这个条件
    -  `map` : 当 `len > 0` 时，会校验每一个 Value 是否满足条件（不校验 key）。
4. ip、ipv4、ipv6: 只校验 IP 格式合法，不校验 IP 是否可用。如 "0.0.0.0"、"127.0.0.0" 等均是合法的。

## 例子

```go
package main

import 	(
    "fmt"

    "github.com/xanygo/anygo/xvld"
)

type User struct {
	Name string `xv:"required,min-len=2,max-len=32"`
    Age  uint16 `xv:"required,min=1,max=200"`
}


func main(){
    u:=&User{Name:"a"}
    fmt.Println(xvld.Struct(u))
    
    // Field 'Name': length must be 2 or greater, got 1
}
```