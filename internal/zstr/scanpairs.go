package zstr

import (
	"fmt"
	"strings"
)

// scanPairs 用于解析字符串,支持格式：
//
//	required,min=3,max=20
//	required,regexp(#)=^[a-zA-Z]{3,16}$#,min=3,max=20
func ScanPairs(input string, sep byte, fn func(key, value string) bool) error {
	if input == "" || sep == 0 || fn == nil {
		return nil
	}
	for len(input) > 0 {
		keyRaw, after, found := cutByte(input, '=')
		if !found {
			for _, k := range splitByte(input, sep) {
				if k != "" {
					fn(k, "")
				}
			}

			return nil
		}
		input = after
		for {
			// required,xxx,hello=world
			// keyRaw=required,xxx,hello
			k, a, f := cutByte(keyRaw, sep)
			if f {
				if k != "" {
					fn(k, "")
				}
				keyRaw = a
			} else {
				break
			}
		}
		key, keyAfter, found2 := cutByte(keyRaw, '(')
		if found2 {
			if len(keyAfter) != 2 || keyAfter[1] != ')' {
				return fmt.Errorf("invalid key=%q", keyRaw)
			}

			// 查找自定义截至符
			value, valueAfter, found3 := cutByte(input, keyAfter[0])
			if !found3 {
				return fmt.Errorf("invalid value %q=%q, miss end delimiter %q", keyRaw, input, keyAfter[0])
			}
			fn(key, value)
			input = valueAfter
		} else {
			value, valueAfter, found3 := cutByte(input, sep)
			if found3 {
				fn(key, value)
				input = valueAfter
			} else {
				fn(key, input)
				return nil
			}
		}
	}
	return nil
}

func cutByte(s string, sep byte) (before, after string, found bool) {
	i := strings.IndexByte(s, sep)
	if i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}

func splitByte(s string, sep byte) []string {
	var parts []string

	for {
		i := strings.IndexByte(s, sep)
		if i < 0 {
			return append(parts, s)
		}

		parts = append(parts, s[:i])
		s = s[i+1:]
	}
}
