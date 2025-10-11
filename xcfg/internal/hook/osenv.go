//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"bytes"
	"context"
	"os"
	"regexp"
)

// 模板变量格式：{env.变量名} 或者 {osenv.变量名|默认值}
var osEnvVarReg = regexp.MustCompile(`\{env\.([A-Za-z0-9_]+)(\|[^}]+)?\}`)

// OsEnvVars 将配置文件中的 {env.xxx} 的内容，从环境变量中读取并替换
func OsEnvVars(_ context.Context, _ string, content []byte) ([]byte, error) {
	contentNew := osEnvVarReg.ReplaceAllFunc(content, func(subStr []byte) []byte {
		// 将 {env.xxx} 中的 xxx 部分取出
		// 或者 将 {env.yyy|val} 中的 yyy|val 部分取出

		keyWithDefaultVal := subStr[len("{env.") : len(subStr)-1] // eg: xxx 或者 yyy|val
		idx := bytes.Index(keyWithDefaultVal, []byte("|"))
		if idx > 0 {
			// {env.变量名|默认值} 有默认值的格式
			key := string(keyWithDefaultVal[:idx])  // eg: yyy
			defaultVal := keyWithDefaultVal[idx+1:] // eg: val
			envVal := os.Getenv(key)
			if len(envVal) == 0 {
				return defaultVal
			}
			return []byte(envVal)
		}

		// {osenv.变量名} 无默认值的部分
		return []byte(os.Getenv(string(keyWithDefaultVal)))
	})
	return contentNew, nil
}
