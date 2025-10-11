//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"context"
	"fmt"
	"regexp"

	"github.com/xanygo/anygo/xattr"
)

// 模板变量格式：{xattr.变量名}
var attrVarReg = regexp.MustCompile(`\{xattr\.([A-Za-z0-9_]+)\}`)

func XAttrVars(_ context.Context, _ string, content []byte) ([]byte, error) {
	var err error
	contentNew := attrVarReg.ReplaceAllFunc(content, func(subStr []byte) []byte {
		// 将 {xattr.xxx} 中的 xxx 部分取出
		key := subStr[len("{xattr.") : len(subStr)-1] // eg: xxx
		var val string
		val, err = getAttrValue(string(key))
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

func getAttrValue(key string) (string, error) {
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
		return "", fmt.Errorf("key=%q not support", key)
	}
}
