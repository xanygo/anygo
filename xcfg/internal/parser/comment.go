//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package parser

import (
	"bytes"
)

// StripComment 去除单行的'#'注释
// 只支持单行，不支持行尾
func StripComment(input []byte) (out []byte) {
	var buf bytes.Buffer
	lines := bytes.Split(input, []byte("\n"))
	for _, line := range lines {
		lineN := bytes.TrimSpace(line)
		if !bytes.HasPrefix(lineN, []byte("#")) {
			buf.Write(line)
		}
		buf.WriteString("\n")
	}
	return bytes.TrimSpace(buf.Bytes())
}

// HeadComments 获取头部的所有注释内容
func HeadComments(input []byte) []string {
	var cmts []string
	lines := bytes.Split(input, []byte("\n"))
	for _, line := range lines {
		lineN := bytes.TrimSpace(line)
		if len(lineN) == 0 {
			continue
		}
		if bytes.HasPrefix(lineN, []byte("#")) {
			cm := bytes.TrimSpace(bytes.TrimLeft(lineN, "#"))
			if len(cm) > 0 {
				cmts = append(cmts, string(cm))
			}
		} else {
			break
		}
	}
	return cmts
}
