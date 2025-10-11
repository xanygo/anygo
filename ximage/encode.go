//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-25

package ximage

import (
	"encoding/base64"
	"image"
	"image/png"

	"github.com/xanygo/anygo/ds/xsync"
)

var pool = xsync.NewBytesBufferPool(0)

func EncodeEmbed(img image.Image) (string, error) {
	const prefix = "data:image/jpeg;base64,"
	w := pool.Get()
	defer pool.Put(w)

	err := png.Encode(w, img)
	if err != nil {
		return "", err
	}
	bf := make([]byte, len(prefix)+base64.StdEncoding.EncodedLen(w.Len()))
	base64.StdEncoding.AppendEncode(bf, w.Bytes())
	return string(bf), nil
}
