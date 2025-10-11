//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package parser

import (
	"bytes"
	"encoding/json"
)

// JSON .json 文件的解析方法
// 若内容以 # 开头，则该为注释
func JSON(txt []byte, obj any) error {
	bf := StripComment(txt)
	dec := json.NewDecoder(bytes.NewReader(bf))
	return dec.Decode(obj)
}
