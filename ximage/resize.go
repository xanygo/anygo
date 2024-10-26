//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package ximage

import (
	"image"
	"image/color"
)

func Resize(src image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算对应的源图像位置
			srcX := float64(x) * float64(srcWidth) / float64(width)
			srcY := float64(y) * float64(srcHeight) / float64(height)

			// 获取周围四个像素点
			x1, y1 := int(srcX), int(srcY)
			x2, y2 := min(x1+1, srcWidth-1), min(y1+1, srcHeight-1)

			// 进行双线性插值
			r, g, b, a := bilinearInterpolation(src, srcX, srcY, x1, y1, x2, y2)
			dst.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

// 双线性插值函数
func bilinearInterpolation(img image.Image, srcX, srcY float64, x1, y1, x2, y2 int) (uint8, uint8, uint8, uint8) {
	q11 := img.At(x1, y1).(color.RGBA)
	q21 := img.At(x2, y1).(color.RGBA)
	q12 := img.At(x1, y2).(color.RGBA)
	q22 := img.At(x2, y2).(color.RGBA)

	r := interpolate(srcX, srcY, float64(x1), float64(y1), float64(x2), float64(y2), q11.R, q21.R, q12.R, q22.R)
	g := interpolate(srcX, srcY, float64(x1), float64(y1), float64(x2), float64(y2), q11.G, q21.G, q12.G, q22.G)
	b := interpolate(srcX, srcY, float64(x1), float64(y1), float64(x2), float64(y2), q11.B, q21.B, q12.B, q22.B)
	a := interpolate(srcX, srcY, float64(x1), float64(y1), float64(x2), float64(y2), q11.A, q21.A, q12.A, q22.A)

	return uint8(r), uint8(g), uint8(b), uint8(a)
}

// 插值计算
func interpolate(x, y, x1, y1, x2, y2 float64, q11, q21, q12, q22 uint8) float64 {
	return float64(q11)*(x2-x)*(y2-y) +
		float64(q21)*(x-x1)*(y2-y) +
		float64(q12)*(x2-x)*(y-y1) +
		float64(q22)*(x-x1)*(y-y1)
}
