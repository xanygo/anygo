//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-08

package ximage

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xanygo/anygo/xerror"
)

type DecoderFunc func(io.Reader) (image.Image, error)

var decoders = map[string]DecoderFunc{}

// RegisterDecoder 注册新的解析解析方法
// ext: 文件后缀，如 .bmp
func RegisterDecoder(ext string, fn DecoderFunc) {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		panic("ext should has prefix '.'")
	}
	decoders[ext] = fn
}

// GetDecoderFuncByExt 获取已注册的好默认内置支持的 DecoderFunc
func GetDecoderFuncByExt(ext string) (DecoderFunc, error) {
	switch strings.ToLower(ext) {
	case ".png":
		return png.Decode, nil
	case ".jpeg", ".jpg":
		return jpeg.Decode, nil
	case ".gif":
		return gif.Decode, nil
	default:
		if fn, ok := decoders[ext]; ok {
			return fn, nil
		}
		return nil, fmt.Errorf("decoder for %q %w", ext, xerror.NotFound)
	}
}

// Load 加载图片文件，并依据文件后缀自动选用合适的解析器解析图片
//
// 默认支持：.png、.jpeg、.jpg、.gif
// 其他后缀可以通过 RegisterDecoder 方法注册实现支持
func Load(fp string) (image.Image, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	fn, err := GetDecoderFuncByExt(filepath.Ext(fp))
	if err != nil {
		return nil, err
	}
	return fn(file)
}
