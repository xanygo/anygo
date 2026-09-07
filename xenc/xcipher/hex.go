//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xcipher

import (
	"encoding/hex"
)

var _ Cipher = (*HEX)(nil)

type HEX struct{}

func (h HEX) Name() string {
	return "HEX"
}

func (h HEX) Encrypt(src []byte) ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst, nil
}

func (h HEX) Decrypt(src []byte) ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	return dst[:n], err
}
