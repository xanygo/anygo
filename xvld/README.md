# xvld

通用的 Go 字段规则校验器。

## 内置规则

| 名称         | 功能                 | 支持的数据类型                       | 举例                                               |
| ------------ | -------------------- | ------------------------------------ | -------------------------------------------------- |
| `required`   | 内容不为空           | 所有                                 | `required`                                         |
| `min`        | 最小值               | string/number, 递归：array/map       | `min=1`                                            |
| `max`        | 最大值               | string/number, 递归：array/map       | `max=10`                                           |
| `len`        | 长度等于             | string/bytes, 递归：array/map        | `len=3`                                            |
| `min-len`    | 最小长度             | string/bytes, 递归：array/map        | `min-len=3`                                        |
| `max-len`    | 最小长度             | string/bytes, 递归：array/map        | `max-len=10`                                       |
| `eq`         | 等于                 | string/bool/number,  递归：array/map | `eq=10`, `eq=hello`                                |
| `ne`         | 不等于               | string/bool/number, 递归：array/map  | `ne=10`, `ne=hello`                                |
| `gt`         | 大于                 | number, 递归：array/map              | `gt=10`                                            |
| `gte`        | 大于等于             | number, 递归：array/map              | `gte=10`                                           |
| `lt`         | 小于                 | number, 递归：array/map              | `lt=10`                                            |
| `lte`        | 小于等于             | number, 递归：array/map              | `lte=10`                                           |
| `contains`   | 包含                 | string/bytes,  递归：array/map       | `contains=10`,`contains=hello`                     |
| `excludes`   | 不包含               | string/bytes,  递归：array/map       | `excludes=10`,`excludes=hello`                     |
| `prefix`     | 包含前缀             | string/bytes,  递归：array/map       | `prefix=hello`                                     |
| `suffix`     | 包含后缀             | string/bytes,  递归：array/map       | `suffix=world`                                     |
| `one-of`     | 是枚举值中的一个     | string/number/bool, 递归：array/map  | `one-of=1\|2\|3`                                   |
| `email`      | 邮箱地址             | string, 递归：array/map              | `email`                                            |
| `url`        | URL 地址             | string, 递归：array/map              | `url`                                              |
| `alpha`      | 英文字母             | string, 递归：array/map              | `alpha`                                            |
| `alpha-num`  | 英文字母和数字       | string, 递归：array/map              | `alpha-num`                                        |
| `uuid`       | 是 UUID              | string, 递归：array/map              | `uuid`                                             |
| `numeric`    | 可运算的数值         | string, 递归：array/map              | `numeric`                                          |
| `number`     | 数字、数字字符串     | string, 递归：array/map              | `number`                                           |
| `uint`       | uint 数字字符串      | string, 递归：array/map              | `uint`                                             |
| `nz-uint`    | 非0 uint 字符串      | string/int/uint, 递归：array/map     | `nz-uint`                                          |
| `hex-color`  | 16进制颜色           | string, 递归：array/map              | `hex-color` ，如 "#1e1f22"                         |
| `base36`     | Base36 字符串        | string, 递归：array/map              | `base36`                                           |
| `base58`     | Base58 字符串        | string, 递归：array/map              | `base58`                                           |
| `base62`     | Base62 字符串        | string, 递归：array/map              | `base62`                                           |
| `base64-url` | Base64 URL 字符串    | string, 递归：array/map              | `base64-url`                                       |
| `base64-std` | Base64 STD 字符串    | string, 递归：array/map              | `base64-std`                                       |
| `regexp`     | 自定义正则表达式     | string, 递归：array/map              | `regexp=\d+`, `required,regexp(#)=\d{3,5}#,number` |
| `dive`       | 后面的规则校验子条目 | array/map                            | `dive,min=1`                                       |

注：
1. 上表中: `array` 包含 `Go`的 array 和 slice,`number` 包含 int/uint/float。
2. `regexp(#)=\d{3,5}#`：值中包含英文逗号时，可以使用在 Key 后面附加 `({Byte})` 来定义特殊的截至符。
3. `array` 的递归：如 `Numbers []int "min=10"` 当 `len > 0` 时，会校验每一个条目是否满足 `min=10` 这个条件
4. `map` 的递归：当 `len > 0` 时，会校验每一个 Value 是否满足条件（不校验 key）。
