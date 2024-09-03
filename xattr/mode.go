//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import "fmt"

// Mode 运行模式
type Mode int32

const (
	// ModeProduct 运行模式-线上发布
	ModeProduct Mode = iota

	// ModeDebug 运行模式-调试
	ModeDebug
)

func (m Mode) String() string {
	switch m {
	case ModeProduct:
		return "product"
	case ModeDebug:
		return "debug"
	default:
		return fmt.Sprintf("unknown(%d)", m)
	}
}

func modeFromEnv(def Mode) Mode {
	switch osEnvDefault(eKeyMode, "") {
	case "product":
		return ModeProduct
	case "debug":
		return ModeDebug
	default:
		return def
	}
}
