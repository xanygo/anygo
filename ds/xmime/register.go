// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123+git@gmail.com>
// Date: 2023/7/28

package xmime

import (
	"mime"
	"sync"
)

func register() {
	for ext, tp := range apache {
		if mime.TypeByExtension(ext) == "" {
			_ = mime.AddExtensionType(ext, tp)
		}
	}
}

var once sync.Once

// Register 调用 mime.AddExtensionType 注册所有的 mime 类型
func Register() {
	once.Do(register)
}
