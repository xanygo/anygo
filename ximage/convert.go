//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package ximage

import (
	"image"
	"image/color"
)

func Resize(src *image.RGBA, width, height int) *image.RGBA {
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
func bilinearInterpolation(img *image.RGBA, srcX, srcY float64, x1, y1, x2, y2 int) (uint8, uint8, uint8, uint8) {
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

// ToGrayImage 将彩色图片转换为黑白图片
func ToGrayImage(img image.Image) *image.Gray {
	if gi, ok := img.(*image.Gray); ok {
		return gi
	}
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	out := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := ToGrayColor(img.At(x, y))
			out.SetGray(x, y, gray)
		}
	}
	return out
}

func ToGrayColor(c color.Color) color.Gray {
	r, g, b, a := c.RGBA()
	if a == 0 {
		// 完全透明，返回白色，更符合一般情况
		return color.Gray{Y: 255}
	}
	gray := uint8(0.2989*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
	if a == 0xffff {
		return color.Gray{Y: gray}
	}
	// 对于半透明像素，按透明度加权计算灰度值
	alpha := uint8(a >> 8)
	weightedGray := uint8((float64(alpha) / 255.0) * float64(gray))
	return color.Gray{Y: weightedGray}
}

// CanvasScale 将宽高调整为指定的比例，在调整时，会保持其中一条边的值不变，让另一条表按照比例放大
func CanvasScale(width int, height int, scale float64) (int, int) {
	if scale < 1 {
		scale = 1.0 / scale
	}
	if width >= height {
		if float64(width)/float64(height) > scale {
			return width, int(float64(width) / scale)
		}
		return int(float64(height) * scale), height
	}
	if float64(height)/float64(width) > scale {
		return int(float64(height) / scale), height
	}
	return width, int(float64(width) * scale)
}
