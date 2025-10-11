//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package pixelfont

import (
	"image/color"
)

// ASCII 获取 ASCII 可见字符 [33,126] 区间的像素描述信息
func ASCII(c byte) Byte {
	return asciiFonts[c]
}

type Byte struct {
	Width  int
	Height int
	Pixel  []byte
}

func (b Byte) DrawTo(img SetAble, startX int, startY int, col color.Color) {
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			index := y*b.Width + x
			if b.Pixel[index] != 1 {
				continue
			}
			img.Set(startX+x, startY+y, col)
		}
	}
}

type SetAble interface {
	Set(x, y int, c color.Color)
}

var maxWidth, maxHeight int

func init() {
	for _, f := range asciiFonts {
		maxWidth = max(maxWidth, f.Width)
		maxHeight = max(maxHeight, f.Height)
	}
}

func MaxWidth() int {
	return maxWidth
}

func MaxHeight() int {
	return maxHeight
}

func GetBytes(str []byte) Bytes {
	bs := make([]Byte, len(str))
	for i, b := range str {
		bs[i] = ASCII(b)
	}
	return bs
}

type Bytes []Byte

func (bs Bytes) Width() int {
	var total int
	for _, b := range bs {
		total += b.Width
	}
	return total
}

func (bs Bytes) DrawTo(img SetAble, startX int, startY int, sep int, col color.Color) {
	for i := 0; i < len(bs); i++ {
		bc := bs[i]
		bc.DrawTo(img, startX, startY, col)
		startX += bc.Width + sep
	}
}
