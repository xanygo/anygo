package buildin

import (
	"github.com/xanygo/anygo/xvld/internal"
)

type (
	FieldContext = internal.FieldContext

	Rule = internal.Rule
)

var Rules = map[string]internal.ValidationFunc{
	"required": required,

	// min/max: 比较值大小，若是 slice / array / map 则比较子元素的大小
	"min": minMax,
	"max": minMax,

	"min-len": validateMinMaxLen,
	"max-len": validateMinMaxLen,
	"len":     validateMinMaxLen,

	"eq":  validateCompare,  // eq=hello
	"ne":  validateNotEqual, // ne=hello   ne=a,b
	"gt":  validateCompare,  // 大于（>） gt=1
	"gte": validateCompare,  // 大于等于(>=) gte=1
	"lt":  validateCompare,  // 小于 lt=1
	"lte": validateCompare,  // 小于等于  lte=1

	"contains": internal.ByMatchString(isContains),
	"excludes": internal.ByMatchString(excludes),
	"prefix":   internal.ByMatchString(isPrefix),
	"suffix":   internal.ByMatchString(isSuffix),

	"one-of": oneOf, // oneof=1|2|3
	"oneof":  oneOf, // oneof=hello|world

	"email":  internal.ByRegexp(mailRegText),
	"url":    internal.ByMatchString(isURL),
	"domain": internal.ByRegexp(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`),
	"port":   validatePort,
	"ip":     validateIP,
	"ipv4":   validateIPV4,
	"ipv6":   validateIPV6,
	"ident":  internal.ByRegexp(`^[a-zA-Z_][a-zA-Z0-9_]*$`),

	"alpha":     internal.ByRegexp(`^[a-zA-Z]+$`),
	"alpha-num": alphaNum,
	"alphanum":  alphaNum, // 别名
	"uuid":      internal.ByMatchString(isUUID),
	"numeric":   internal.ByMatchString(isNumeric), // 可进行运算的数值
	"number":    internal.ByRegexp(`^[0-9]+$`),     // 由数字组成的字符串，如  "0123"
	"uint":      internal.ByRegexp(`^0|[1-9][0-9]*`),
	"nz-uint":   notZeroUint, // 不为 0 的 uint

	"hex-color":  internal.ByRegexp(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`),
	"hex":        internal.ByRegexp(`(?i)^[0-9a-f]+$`),
	"base36":     internal.ByRegexp(`^[0-9a-z]+$`),
	"base58":     internal.ByRegexp(`^[1-9A-HJ-NP-Za-km-z]+$`),
	"base62":     internal.ByRegexp(`^[0-9a-zA-Z]+$`),
	"base64-url": internal.ByRegexp(`^[0-9a-zA-Z\-_]+$`),
	"base64-std": internal.ByRegexp(`^[A-Za-z0-9+/]+={0,2}$`),
	"md4":        internal.ByRegexp(`^[0-9a-f]{32}$`),
	"md5":        internal.ByRegexp(`^[0-9a-f]{32}$`),
	"sha256":     internal.ByRegexp(`^[0-9a-f]{64}$`),
	"sha384":     internal.ByRegexp(`^[0-9a-f]{96}$`),
	"sha512":     internal.ByRegexp(`^[0-9a-f]{128}$`),

	"regexp": internal.ByMatchString(isRegexp), // 校验正则表达式
}

var alphaNum = internal.ByRegexp(`^[a-zA-Z0-9]+$`)
