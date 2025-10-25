//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package xcolor

type Code uint8

// 普通色
const (
	CodeFgBlack Code = iota + 30
	CodeFgRed
	CodeFgGreen
	CodeFgYellow
	CodeFgBlue
	CodeFgMagenta
	CodeFgCyan
	CodeFgWhite
)

// 高亮色
const (
	CodeFgHiBlack Code = iota + 90
	CodeFgHiRed
	CodeFgHiGreen
	CodeFgHiYellow
	CodeFgHiBlue
	CodeFgHiMagenta
	CodeFgHiCyan
	CodeFgHiWhite
)

const reset Code = 0

// 背景色：普通色
const (
	CodeBgBlack Code = iota + 40
	CodeBgRed
	CodeBgGreen
	CodeBgYellow
	CodeBgBlue
	CodeBgMagenta
	CodeBgCyan
	CodeBgWhite
)

// 背景色：高亮色
const (
	CodeBgHiBlack Code = iota + 100
	CodeBgHiRed
	CodeBgHiGreen
	CodeBgHiYellow
	CodeBgHiBlue
	CodeBgHiMagenta
	CodeBgHiCyan
	CodeBgHiWhite
)
