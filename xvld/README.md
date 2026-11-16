# xvld

通用的 Go 字段规则校验器。

## 内置规则

| 名称         | 功能                 | 支持的数据类型                              | 举例                                         |
| ------------ | -------------------- | ------------------------------------------- | -------------------------------------------- |
| `required`   | 内容不为空           | 所有                                        | `required`                                   |
| `min`        | 最小值               | string/int/uint/float, 递归：array/map       | `min=1`                                      |
| `max`        | 最大值               | string/int/uint/float, 递归：array/map       | `max=10`                                     |
| `len`        | 长度等于             | string/float/array/map                      | `len=3`                                      |
| `min-len`    | 最小长度             | string/float/array/map                      | `min-len=3`                                  |
| `max-len`    | 最小长度             | string/float/array/map                      | `max-len=10`                                 |
| `eq`         | 等于                 | string/bool/int/uint/float,  递归：array/map | `eq=10`, `eq=hello`                          |
| `ne`         | 不等于               | string/bool/int/uint/float, 递归：array/map  | `ne=10`, `ne=hello`                          |
| `gt`         | 大于                 | int/uint/float, 递归：array/map              | `gt=10`                                      |
| `gte`        | 大于等于             | int/uint/float, 递归：array/map              | `gte=10`                                     |
| `lt`         | 小于                 | int/uint/float, 递归：array/map              | `lt=10`                                      |
| `lte`        | 小于等于             | int/uint/float, 递归：array/map              | `lte=10`                                     |
| `contains`   | 包含                 | string/bytes,  递归：array/map               | `contains=10`,`contains=hello`               |
| `excludes`   | 不包含               | string/bytes,  递归：array/map               | `excludes=10`,`excludes=hello`               |
| `one-of`     | 是枚举值中的一个     | string/int/uint/float/bool, 递归：array/map  | `one-of=1\|2\|3`                                    | 
| `email`      | 邮箱地址             | string, 递归：array/map                      | `email`                                      |
| `url`        | URL 地址             | string, 递归：array/map                      | `url`                                        |
| `alpha`      | 英文字母             | string, 递归：array/map                      | `alpha`                                      |
| `alpha-num`  | 英文字母和数字       | string, 递归：array/map                      | `alpha-num`                                  |
| `uuid`       | 是 UUID              | string, 递归：array/map                      | `uuid`                                       |
| `numeric`    | 可运算的数值         | string, 递归：array/map                      | `numeric`                                    |
| `number`     | 数字、数字字符串     | string, 递归：array/map                      | `number`                                     |
| `uint`       | uint 数字字符串      | string, 递归：array/map                      | `uint`                                       |
| `nz-uint`    | 非0 uint 字符串      | string/int/uint, 递归：array/map             | `nz-uint`                                    |
| `hex-color`  | 16进制颜色           | string, 递归：array/map                      | `hex-color`                                  |
| `base36`     | Base 36 字符串       | string, 递归：array/map                      | `base36`                                     |
| `base58`     | Base 58 字符串       | string, 递归：array/map                      | `base58`                                     |
| `base62`     | Base 62 字符串       | string, 递归：array/map                      | `base62`                                     |
| `base64-url` | Base 64 URL 字符串   | string, 递归：array/map                      | `base64-url`                                 |
| `base64-std` | Base 64 STD 字符串   | string, 递归：array/map                      | `base64-std`                                 |
| `regexp`     | 使用自定义正则表达式 | string, 递归：array/map                      | `regexp=\d+`, `required,regexp(#)=\d{3,5}#,number` |

注：
1. 上表中 `array` 包含 `Go`的 array 和 slice。
2. `regexp(#)=\d{3,5}#`：值中包含英文逗号时，可以使用在 Key 后面附加 `({Byte})` 来定义特殊的截至符。
