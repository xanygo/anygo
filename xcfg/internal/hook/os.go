//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-11

package hook

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// 模板变量格式：{os.变量名}
var osVarReg = regexp.MustCompile(`\{os\.([A-Za-z0-9_]+)\}`)

func OSVars(_ context.Context, _ string, content []byte) ([]byte, error) {
	var err error
	contentNew := osVarReg.ReplaceAllFunc(content, func(subStr []byte) []byte {
		// 将 {os.xxx} 中的 xxx 部分取出
		key := subStr[len("{os.") : len(subStr)-1] // eg: xxx
		var val string
		val, err = getOsValue(string(key))
		if err != nil {
			return nil
		}
		return []byte(val)
	})
	if err != nil {
		return nil, err
	}
	return contentNew, err
}

func getOsValue(key string) (string, error) {
	switch key {
	case "PID":
		return strconv.Itoa(os.Getpid()), nil
	case "PPID":
		return strconv.Itoa(os.Getppid()), nil
	case "TempDir":
		return os.TempDir(), nil
	case "Hostname":
		return os.Hostname()
	case "UserHomeDir":
		return os.UserHomeDir()
	case "UserCacheDir":
		return os.UserCacheDir()
	case "UserConfigDir":
		return os.UserConfigDir()
	default:
		return "", fmt.Errorf("key=%q not support", key)
	}
}
