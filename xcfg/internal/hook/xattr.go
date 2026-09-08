//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xanygo/anygo/xattr"
)

// 模板变量格式：{xattr.变量名} 或者 {xattr.变量名|连接的值}
var attrVarReg = regexp.MustCompile(`\{xattr\.([A-Za-z0-9_]+)(\|[^}]+)?\}`)

func XAttrVars(_ context.Context, _ string, content []byte) ([]byte, error) {
	var err error
	contentNew := attrVarReg.ReplaceAllFunc(content, func(subStr []byte) []byte {
		// 将 {xattr.xxx} {xattr.xxx|value} 中的 xxx 或 xxx|value 取出
		keyWithVal := subStr[len("{xattr.") : len(subStr)-1] // eg: yyy 或者 yyy|val
		key, value, found := strings.Cut(string(keyWithVal), "|")

		rawKey := key

		isDir := strings.HasSuffix(key, "Dir")
		var isRel bool
		if isDir {
			// {xattr.RelRootDir} 取回的是相对目录
			key, isRel = strings.CutPrefix(key, "Rel")
		}
		var val string
		val, err = getAttrValue(key, rawKey)
		if err != nil {
			return nil
		}
		value = strings.TrimSpace(value)

		if found && isDir {
			val = filepath.Join(val, value)
		}
		if isRel {
			var wd string
			wd, err = os.Getwd()
			if err != nil {
				return nil
			}
			val, err = filepath.Rel(wd, val)
		}
		return []byte(val)
	})
	if err != nil {
		return nil, err
	}
	return contentNew, err
}

func getAttrValue(key string, rawKey string) (string, error) {
	switch key {
	case "RootDir":
		return xattr.RootDir(), nil
	case "IDC":
		return xattr.IDC(), nil
	case "DataDir":
		return xattr.DataDir(), nil
	case "ConfDir":
		return xattr.ConfDir(), nil
	case "TempDir":
		return xattr.TempDir(), nil
	case "LogDir":
		return xattr.LogDir(), nil
	case "RunMode":
		return xattr.RunMode().String(), nil
	default:
		return "", fmt.Errorf("key=%q not support", rawKey)
	}
}
