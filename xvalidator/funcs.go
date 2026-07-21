//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-02

package xvalidator

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// IsHTTPURL 是否有效的 HTTP URL 地址
func IsHTTPURL(str string) bool {
	scheme, _, ok := strings.Cut(str, "://")
	return ok && (scheme == "http" || scheme == "https")
}

// CheckHTTPURL 是否有效的 HTTP URL 地址,若不是则返回 error
func CheckHTTPURL(str string) error {
	if IsHTTPURL(str) {
		return nil
	}
	return fmt.Errorf("%q is not HTTP url", str)
}

func CheckStringIn(value string, values ...string) error {
	if slices.Contains(values, value) {
		return nil
	}
	return fmt.Errorf("%q is not in %q", value, values)
}

// CheckMapHasKeys 检查 map 是否包含传入的全部 keys。
// 若传入 map 或 keys 为空，总是返回 error
func CheckMapHasKeys[K comparable, V any](m map[K]V, keys ...K) error {
	if len(m) == 0 {
		return errors.New("empty map")
	}
	if len(keys) == 0 {
		return errors.New("empty keys")
	}
	var missKeys []K
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			missKeys = append(missKeys, k)
		}
	}
	if len(missKeys) == 0 {
		return nil
	}
	return fmt.Errorf("missing keys: %v", missKeys)
}
