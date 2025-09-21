//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-20

package xhtml

import "regexp"

// 匹配所有 <xxx> 形式的标签
var stripReg = regexp.MustCompile(`<[^>]*>`)

func StripTags(input string) string {
	return stripReg.ReplaceAllString(input, "")
}
