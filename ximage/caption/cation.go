//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package caption

import (
	"image"
	"net/http"

	"github.com/xanygo/anygo/ximage"
)

type Caption interface {
	http.Handler
	Image() image.Image
	Code() string
}

type base struct {
	code       string
	width      int
	height     int
	reDraw     func(img image.Image)
	randomLine int // 随机线条数
}

func (b *base) SetSize(width int, height int) {
	b.width = width
	b.height = height
}

func (b *base) SetCode(code string) {
	b.code = code
}

func (b *base) Code() string {
	return b.code
}

func (b *base) getSize(srcWidth int, srcHeight int) (width, height int, ok bool) {
	if b.width == 0 && b.height == 0 {
		return 0, 0, false
	}
	if b.width > 0 && b.height > 0 {
		return b.width, b.height, true
	}
	if b.width > 0 {
		return b.width, srcHeight * b.width / srcWidth, true
	}
	return srcWidth * srcHeight / b.height, b.height, true
}

// SetReDraw 调整完大小，待输出前加工图片
func (b *base) SetReDraw(fn func(img image.Image)) {
	b.reDraw = fn
}

// SetRandomLine 设置干扰线数量，若为 -1 则不绘制
// 默认值为 0 - 自动依据 width 大小比例绘制 width / 12 条
func (b *base) SetRandomLine(num int) {
	b.randomLine = num
}

func (b *base) drawRandomLine(img *image.RGBA) {
	if b.randomLine < 0 {
		return
	}
	num := b.randomLine
	if num == 0 { // auto number
		num = img.Bounds().Dx() / 12
	}
	for i := 0; i < num; i++ {
		ximage.DrawRandomLine(img)
	}
}

func (b *base) tryReDraw(img image.Image) {
	if b.reDraw != nil {
		b.reDraw(img)
	}
}
