//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-24

package zroute

import (
	"fmt"
	"strings"
)

type Meta struct {
	ID    string
	Other map[string]string
}

func parserMeta(str string) (Meta, error) {
	str = strings.TrimSpace(str)
	arr := strings.Split(str, ",")
	meta := Meta{
		Other: make(map[string]string),
	}
	for i := range arr {
		txt := strings.TrimSpace(arr[i])
		if txt == "" {
			continue
		}
		key, value, ok := strings.Cut(txt, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok || key == "" {
			return Meta{}, fmt.Errorf("invalid meta[%d] = %q", i, arr[i])
		}
		if strings.EqualFold(key, "id") {
			meta.ID = value
		} else {
			meta.Other[key] = value
		}
	}
	return meta, nil
}
