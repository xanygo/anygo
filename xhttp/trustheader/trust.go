package trustheader

import (
	"net/http"
	"net/textproto"
)

var trustHeaders = map[string]struct{}{}

// Add 添加可信 Header 字段
func Add(keys ...string) {
	for _, key := range keys {
		trustHeaders[textproto.CanonicalMIMEHeaderKey(key)] = struct{}{}
	}
}

// IsTrusted 判断是否可信字段
func IsTrusted(key string) bool {
	key = textproto.CanonicalMIMEHeaderKey(key)
	_, trusted := trustHeaders[key]
	return trusted
}

// Get 读取 Header value,若 key 是可信的，并且不为空，则正常返回
func Get(h http.Header, key string) (value string, ok bool) {
	if !IsTrusted(key) {
		return "", false
	}
	value = h.Get(key)
	return value, value != ""
}
