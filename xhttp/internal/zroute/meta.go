//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-24

package zroute

import (
	"fmt"
	"strings"
)

type Meta struct {
	ID         string
	Other      map[string]string
	PathValues map[string]string
}

// 解析meta 字符串，如
// id=1,k1=v1,,k2=
// id=1,k1=v1,,k2=/PathValues{}
func parserMeta(str string) (Meta, error) {
	lines := strings.Split(str, "|")
	meta := Meta{
		Other: map[string]string{},
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		const prefix = "PathValues{"
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, "}") {
			pv, err := parserMetaKVStr(line[len(prefix) : len(line)-1])
			if err != nil {
				return meta, err
			}
			meta.PathValues = pv
			continue
		}
		kvs, err := parserMetaKVStr(line)
		if err != nil {
			return meta, err
		}
		if id, ok := kvs["id"]; ok {
			meta.ID = id
			delete(kvs, "id")
		}
		meta.Other = kvs
	}
	return meta, nil
}

func parserMetaKVStr(str string) (map[string]string, error) {
	result := map[string]string{}
	arr := strings.Split(str, ",")
	for i := range arr {
		txt := strings.TrimSpace(arr[i])
		if txt == "" {
			continue
		}
		key, value, ok := strings.Cut(txt, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid meta[%d] %q", i, str)
		}
		result[key] = value
	}
	return result, nil
}
