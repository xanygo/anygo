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
func IsHTTPURL(str string) error {
	scheme, _, ok := strings.Cut(str, "://")
	if !ok || (scheme != "http" && scheme != "https") {
		return fmt.Errorf("%q is not HTTP url", str)
	}
	return nil
}

func StringIn(value string, values ...string) error {
	if slices.Contains(values, value) {
		return nil
	}
	return fmt.Errorf("%q is not in %q", value, values)
}

func MapHasKeys[K comparable, V any](m map[K]V, keys ...K) error {
	if len(m) == 0 {
		return errors.New("empty map")
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
