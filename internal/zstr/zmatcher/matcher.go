//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-16

package zmatcher

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type MatchFunc func(string) bool

var cache sync.Map // map[string]MatchFunc

// Compile 编译规则为 MatchFunc（带缓存）
func Compile(pattern string) (MatchFunc, error) {
	if v, ok := cache.Load(pattern); ok {
		return v.(MatchFunc), nil
	}

	fn, err := compile(pattern)
	if err != nil {
		// 即使编译失败，也要返回  fn
		return AllFalse, err
	}

	cache.Store(pattern, fn)
	return fn, nil
}

func compile(pattern string) (MatchFunc, error) {
	// Regex
	if strings.HasPrefix(pattern, "re:") {
		re, err := regexp.Compile(pattern[3:])
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		return re.MatchString, nil
	}

	// Wildcard 包含 * 和 ?
	if strings.HasPrefix(pattern, "wc:") {
		pattern = pattern[3:]
		//  Fast path（优先于 其他 wildcard）
		if fn := tryFastMatch(pattern); fn != nil {
			return fn, nil
		}

		// Wildcard (包含 * 和 ?)
		if strings.ContainsAny(pattern, "*?") {
			re, err := regexp.Compile(wildcardToRegex(pattern))
			if err != nil {
				return nil, fmt.Errorf("invalid wildcard: %w", err)
			}
			return re.MatchString, nil
		}
		// 若无，继续，当做普通字符串匹配
	}

	// Literal（默认）
	return func(s string) bool {
		return s == pattern
	}, nil
}

// tryFastMatch 针对简单 wildcard 做优化
func tryFastMatch(pattern string) MatchFunc {
	if pattern == "*" {
		return func(string) bool { return true }
	}

	// *xxx*
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		mid := pattern[1 : len(pattern)-1]
		if !strings.ContainsAny(mid, "*?") {
			return func(s string) bool {
				return strings.Contains(s, mid)
			}
		}
	}

	// xxx*
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		if !strings.ContainsAny(prefix, "*?") {
			return func(s string) bool {
				return strings.HasPrefix(s, prefix)
			}
		}
	}

	// *xxx
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		if !strings.ContainsAny(suffix, "*?") {
			return func(s string) bool {
				return strings.HasSuffix(s, suffix)
			}
		}
	}
	return nil
}

// wildcardToRegex 将 * ? 转换为正则
func wildcardToRegex(p string) string {
	var b strings.Builder
	b.Grow(len(p) * 2)

	b.WriteByte('^')

	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteByte('.')
		default:
			b.WriteString(regexp.QuoteMeta(string(p[i])))
		}
	}

	b.WriteByte('$')
	return b.String()
}

func AllFalse(string) bool {
	return false
}
