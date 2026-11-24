//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xcompress

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xanygo/anygo/xenc"
)

// ZipFileNames 获取文件名列表
func ZipFileNames(rd *zip.Reader, strip uint) []string {
	result := make([]string, 0, len(rd.File))
	for _, f := range rd.File {
		np := stripComponents(f.Name, strip)
		if np != "" {
			result = append(result, np)
		}
	}
	return result
}

func stripComponents(p string, n uint) string {
	if n == 0 {
		return p
	}
	sc := int(n)
	ps := strings.Split(path.Clean(p), "/")
	if len(ps) < sc {
		return ""
	}
	return path.Join(ps[sc:]...)
}

// DecryptZip 从加密的 zip 字节流中解析出 zip.Reader 信息
// 该内容，可以使用 cmd/anygo-encrypt-zip 创建
func DecryptZip(b []byte, dc xenc.IDDecrypter) (*zip.Reader, error) {
	if len(b) < 16 {
		return nil, fmt.Errorf("file too short %d bytes", len(b))
	}
	content := b[:len(b)-16]
	xm := md5.New()
	xm.Write(content)
	xm.Write(dc.ID())
	sign := xm.Sum(nil)
	expect := b[len(b)-16:]
	if !bytes.Equal(expect, sign) {
		return nil, errors.New("invalid signature")
	}
	zipContent, err := dc.Decrypt(b)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
}

func MustDecryptZip(b []byte, dc xenc.IDDecrypter) *zip.Reader {
	r, err := DecryptZip(b, dc)
	if err != nil {
		panic(err)
	}
	return r
}

type ZipFileStrip struct {
	Input string // 输入的 zip 文件，必填

	Output string // 输出的 zip 文件，必填

	// StripComponents 移除 ZIP 内部路径开头的目录层级，比条，应传入 >0 的值
	StripComponents uint

	// Overwrite 是否覆盖已存在的输出文件
	Overwrite bool

	// Filter 文件列表过滤，可选，返回 true 表示保留文件，false 表示跳过文件
	Filter func(name string) bool
}

func (opts *ZipFileStrip) Run() error {
	if opts.Input == "" {
		return errors.New("input is required")
	}
	if opts.Output == "" {
		return errors.New("output is required")
	}
	if opts.StripComponents == 0 {
		return errors.New("strip must be greater than 0")
	}
	if err := os.MkdirAll(filepath.Dir(opts.Output), 0755); err != nil {
		return err
	}

	if !opts.Overwrite {
		if _, err := os.Stat(opts.Output); err == nil {
			return os.ErrExist
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	src, err := zip.OpenReader(opts.Input)
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(opts.Output)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(out)

	var result error

	for _, file := range src.File {
		if opts.Filter != nil && !opts.Filter(file.Name) {
			continue
		}

		parts := strings.Split(file.Name, "/")
		if uint(len(parts)) <= opts.StripComponents {
			continue
		}

		header := file.FileHeader
		header.Name = strings.Join(parts[opts.StripComponents:], "/")

		reader, err := file.OpenRaw()
		if err != nil {
			result = err
			break
		}

		writer, err := zw.CreateRaw(&header)
		if err != nil {
			result = err
			break
		}

		if _, err := io.Copy(writer, reader); err != nil {
			result = err
			break
		}
	}

	if err := zw.Close(); result == nil {
		result = err
	}

	if err := out.Close(); result == nil {
		result = err
	}

	if result != nil {
		os.Remove(opts.Output)
	}

	return result
}
