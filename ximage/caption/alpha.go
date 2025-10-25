//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-24

package caption

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"sync"

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ximage"
	"github.com/xanygo/anygo/ximage/pixelfont"
)

func NewAlphaNumber(code string) *AlphaNumber {
	a := &AlphaNumber{}
	a.SetCode(code)
	return a
}

func NewRandom(length int) *AlphaNumber {
	code := xstr.TableAlphaNum.RandNChar(length)
	return NewAlphaNumber(code)
}

func NewRandomNumber(length int) *AlphaNumber {
	code := xstr.TableNum.RandNChar(length)
	return NewAlphaNumber(code)
}

var _ Caption = (*AlphaNumber)(nil)

// AlphaNumber 支持使用ASCII 区间 33-126的字符的验证码
//
// 包括：大小写字母、数字、+-*/=、"、‘()、#！等
type AlphaNumber struct {
	base base
	img  image.Image
	once sync.Once
}

func (p *AlphaNumber) SetCode(str string) {
	p.base.SetCode(str)
}

func (p *AlphaNumber) Code() string {
	return p.base.Code()
}

func (p *AlphaNumber) Image() image.Image {
	p.once.Do(p.drawImage)
	return p.img
}

func (p *AlphaNumber) SetSize(width int, height int) {
	p.base.SetSize(width, height)
}

func (p *AlphaNumber) drawImage() {
	bs := pixelfont.GetBytes([]byte(p.base.code))
	width := bs.Width() + len(p.base.code)*5
	height := pixelfont.MaxHeight()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bs.DrawTo(img, 0, 0, 5, color.Black)

	nw, nh, ok := p.base.getSize(width, height)
	if ok {
		img = ximage.Resize(img, nw, nh)
	}
	p.base.drawRandomLine(img)
	p.base.tryReDraw(img)

	p.img = img
}

func (p *AlphaNumber) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	img := p.Image()

	w.Header().Set("Content-Type", "image/png")
	_ = png.Encode(w, img)
}
