//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package caption

import (
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"net/http"

	"github.com/xanygo/anygo/ximage"
	"github.com/xanygo/anygo/ximage/pixelfont"
)

var _ Caption = (*Random)(nil)

func NewRandom(length int) *Random {
	rd := &Random{}
	rd.genCode(length)
	return rd
}

const digitsTable = "0123456789"

func NewRandomDigits(length int) *Random {
	rd := &Random{
		Table: digitsTable,
	}
	rd.genCode(length)
	return rd
}

type Random struct {
	base

	// Table 可选，默认为所有字母和数字
	Table string
}

func (p *Random) genCode(length int) {
	if length == 0 {
		length = 4
	}
	const strTable = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	table := p.Table
	if table == "" {
		table = strTable
	}
	bf := make([]byte, length)
	for i := 0; i < length; i++ {
		bf[i] = table[rand.IntN(len(table))]
	}
	p.code = string(bf)
}

func (p *Random) Image() image.Image {
	bs := pixelfont.GetBytes([]byte(p.code))
	width := bs.Width() + len(p.code)*5
	height := pixelfont.MaxHeight()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bs.DrawTo(img, 0, 0, 5, color.Black)

	nw, nh, ok := p.getSize(width, height)
	if ok {
		img = ximage.Resize(img, nw, nh)
	}
	p.drawRandomLine(img)
	p.tryReDraw(img)

	return img
}

func (p *Random) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	img := p.Image()
	w.Header().Set("Content-Type", "image/png")
	_ = png.Encode(w, img)
}
