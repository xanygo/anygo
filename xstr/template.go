//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-19

package xstr

import (
	"bytes"
	"text/template"
)

// RenderTemplate 将字符串作为模版渲染
func RenderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("text-template").Option("missingkey=invalid").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
