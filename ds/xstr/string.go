//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-04

package xstr

import (
	"unicode"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/internal/zstr/zmatcher"
)

// ToSnakeCase 驼峰转换为下划线格式
//
//	UserID           -> user_id
//	UserName         -> user_name
//	XMLHTTPRequest   -> xmlhttp_request
func ToSnakeCase(s string) string {
	runes := []rune(s)
	var out []rune

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				var next rune
				if i+1 < len(runes) {
					next = runes[i+1]
				}
				// 如果不是第一个字符，并且（前一个是小写）或（后一个存在且是小写），则插入下划线
				if unicode.IsLower(prev) || (next != 0 && unicode.IsLower(next)) {
					out = append(out, '_')
				}
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
	}

	return string(out)
}

// Match 判断字符串 str 是否和 pattern 匹配
//
// pattern 的几种模式：
//  1. 若有 "wc:" 前缀，支持字符串中包含 *?。* 匹配任意长度字符串（包含空字符串），? 匹配任意单个字符
//  2. 若有 "re:" 前缀，则将后面的内容整个当做正则表达式，采用正则匹配
//  3. 其他情况，当做普通字符串比较是否相等: return pattern == str
//  4. 正则表达式会被缓存，缓存最近使用的 256 个
func Match(pattern, str string) bool {
	ok, _ := MatchE(pattern, str)
	return ok
}

// MatchE 判断字符串 str 是否和 pattern 匹配
//
// 若 pattern 存在错误，会返回 error，pattern 规则详见 CompileMatch 的文档
func MatchE(pattern, str string) (bool, error) {
	fn, err := CompileMatch(pattern)
	return fn(str), err
}

var matchCompilePool = xmap.NewLRU[string, *patternCompile](2048)

type patternCompile struct {
	Match func(string) bool
	Err   error
}

// CompileMatch 将字符串规则编译为一个可执行的匹配函数。
//
// 支持以下规则类型（通过前缀区分）：
//
//  1. 正则表达式（re: 前缀）。例如：re:^h.*o$
//
//  2. 通配符（wc: 前缀）。支持：* 匹配任意长度字符串（包含空字符串），? 匹配任意单个字符
//     例如：wc:he*o ，wc:*ell*
//
//  3. 普通字符串（无前缀），表示完全匹配，例如：hello
//
// 返回值：
//   - func(string) bool：匹配函数（总是不为 nil，当 err!=nil 时），输入字符串，返回是否匹配。
//   - error：当规则非法（如正则语法错误）时返回
//
// 行为说明：
//   - 编译过程只执行一次，建议复用返回的函数以提升性能
//   - 已对结果添加 LRU 缓存
//   - 若 pattern 非法，返回的函数也不会为 nil （实际为 func(string)bool{ return false}），同时返回 error
//
// 示例：
//
//	fn, err := CompileMatch("wc:he*o")
//	if err != nil {
//	    panic(err)
//	}
//	ok := fn("hello") // true
func CompileMatch(pattern string) (func(string) bool, error) {
	val, ok := matchCompilePool.Get(pattern)
	if ok {
		return val.Match, val.Err
	}
	fn, err := zmatcher.Compile(pattern)
	val = &patternCompile{Match: fn, Err: err}
	matchCompilePool.Set(pattern, val)
	return fn, err
}
