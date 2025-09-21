//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xzip

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/xanygo/anygo/xcodec"
)

func FileNames(rd *zip.Reader, strip uint) []string {
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

// Decrypt 从加密的 zip 字节流中解析出 zip.Reader 信息
// 该内容，可以使用 cmd/anygo-encrypt-zip 创建
func Decrypt(b []byte, dc xcodec.IDDecrypter) (*zip.Reader, error) {
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

func MustDecrypt(b []byte, dc xcodec.IDDecrypter) *zip.Reader {
	r, err := Decrypt(b, dc)
	if err != nil {
		panic(err)
	}
	return r
}
