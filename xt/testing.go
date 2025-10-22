// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/8

package xt

type Testing interface {
	Fatalf(format string, args ...any)
}

type Helper interface {
	Helper()
}
