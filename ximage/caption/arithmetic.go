//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package caption

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/xanygo/anygo/ximage"
	"github.com/xanygo/anygo/ximage/pixelfont"
)

func NewArithmetic() *Arithmetic {
	a := &Arithmetic{}
	a.genCode()
	return a
}

var _ Caption = (*Arithmetic)(nil)

// Arithmetic 算数表达式验证码
type Arithmetic struct {
	code string
	exp  string
}

func (a *Arithmetic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	img := a.Image()
	w.Header().Set("Content-Type", "image/png")
	_ = png.Encode(w, img)
}

func (a *Arithmetic) genCode() {
	a.exp, a.code = a.newExp()
}

func (a *Arithmetic) Image() image.Image {
	bs := pixelfont.GetBytes([]byte(a.exp))
	img := image.NewRGBA(image.Rect(0, 0, bs.Width()+len(a.exp)*5, pixelfont.MaxHeight()))
	bs.DrawTo(img, 0, 0, 5, color.Black)

	for i := 0; i < 10; i++ {
		ximage.DrawRandomLine(img)
	}
	return img
}

func (a *Arithmetic) Code() string {
	return a.code
}

func (a *Arithmetic) newExp() (string, string) {
	const opTable = "+-*/"
	op := string(opTable[rand.IntN(len(opTable))])
	var x, y, result int
	switch op {
	case "+":
		x = intRange(1, 99)
		y = rand.IntN(99)
		result = x + y
	case "-":
		x = intRange(10, 99)
		y = intRange(1, x)
		result = x - y
	case "*":
		x = intRange(1, 99)
		y = intRange(2, 5)
		result = x * y
	case "/":
		y = intRange(2, 9)
		for result == 0 {
			result = intRange(1, 200) / y
		}
		x = result * y
	}
	exp := fmt.Sprintf("%d%s%d=?", x, op, y)
	return exp, strconv.Itoa(result)
}

func intRange(min int, max int) int {
	var value int
	for value < min {
		value = rand.IntN(max)
	}
	return value
}
