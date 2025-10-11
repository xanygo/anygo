//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package ximage

import (
	"image"
	"image/color"
	"math/rand/v2"
)

type SetAble interface {
	Set(x, y int, c color.Color)
}

func DrawLine(img SetAble, start image.Point, end image.Point, c color.Color) {
	x1, y1 := start.X, start.Y
	x2, y2 := end.X, end.Y
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}
	sy := 1
	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}
	dxy := dx - dy
	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		dxy2 := dxy * 2
		if dxy2 > -dy {
			dxy -= dy
			x1 += sx
		}
		if dxy2 < dx {
			dxy += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func RandomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.IntN(256)),
		G: uint8(rand.IntN(256)),
		B: uint8(rand.IntN(256)),
		A: 255,
	}
}

func RandomPoint(size image.Point) image.Point {
	return image.Point{
		X: rand.IntN(size.X),
		Y: rand.IntN(size.Y),
	}
}

func DrawRandomLine(img *image.RGBA) {
	size := img.Rect.Size()
	DrawLine(img, RandomPoint(size), RandomPoint(size), RandomColor())
}
