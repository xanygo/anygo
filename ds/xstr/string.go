//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-04

package xstr

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/xanygo/anygo/ds/xmap"
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

var matchPool = xmap.NewLRU[string, *regexp.Regexp](256)

// Match 判断字符串 str 适合和 pattern 匹配
//
// pattern 的几种模式：
//  1. 若有 "star:" 前缀，支持字符串中包含 * 通配符，* 可表示 >=0 个任意字符
//  2. 若有 "regexp:" 前缀，则将后面的内容整个当做正则表达式，采用正则匹配
//  3. 其他情况，当做普通字符串比较是否相等: return pattern == str
//  4. 正则表达式会被缓存，缓存最近使用的 256 个
func Match(pattern, str string) bool {
	reCached, ok := matchPool.Get(pattern)
	if ok {
		return reCached.MatchString(str)
	}
	const regexpPrefix = `regexp:`
	var found bool
	pattern, found = strings.CutPrefix(pattern, regexpPrefix)
	if found {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		matchPool.Set(pattern, re)
		return re.MatchString(str)
	}
	const starPrefix = `star:`
	pattern, found = strings.CutPrefix(pattern, starPrefix)
	if !found {
		return pattern == str
	}
	rePattern := regexp.QuoteMeta(pattern)
	rePattern = strings.ReplaceAll(rePattern, `\*`, ".*")
	rePattern = "^" + rePattern + "$"
	re, err := regexp.Compile(rePattern)
	if err != nil {
		return false
	}
	matchPool.Set(pattern, re)
	return re.MatchString(str)
}
