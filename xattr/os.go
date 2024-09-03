//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	eKeyIDC  = "ANYGO_IDC"
	eKeyConf = "ANYGO_CONF"
	eKeyRoot = "ANYGO_ROOT"
	eKeyData = "ANYGO_DATA"
	eKeyTemp = "ANYGO_TEMP"
	eKeyLog  = "ANYGO_LOG"
	eKeyMode = "ANYGO_MODE"
)

func osEnvDefault(key string, def string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return def
	}
	return val
}

// parserDirName 解析目录名称，若以 |abs 结尾则表示是绝对路径
func parserDirName(path string) (string, bool) {
	before, found := strings.CutSuffix(path, "|abs")
	if found {
		return before, true
	}
	return path, filepath.IsAbs(path)
}
