//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import "context"

// Func helper 的函数
type Func func(ctx context.Context, cfPath string, confContent []byte) ([]byte, error)
