//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-12-05

package xcodec

import "encoding/json"

// Convert 类型转换
func Convert(from any, to any) error {
	bf, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bf, to)
}
