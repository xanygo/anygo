//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-31

package xcodec

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
)

var _ Cipher = (*Base64)(nil)

type Base64 struct {
	Encoder *base64.Encoding
}

func (b Base64) getEncoder() *base64.Encoding {
	if b.Encoder == nil {
		return base64.StdEncoding
	}
	return b.Encoder
}

func (b Base64) Encrypt(src []byte) ([]byte, error) {
	enc := b.getEncoder()
	buf := make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(buf, src)
	return buf, nil
}

func (b Base64) Decrypt(src []byte) ([]byte, error) {
	enc := b.getEncoder()
	buf := make([]byte, enc.DecodedLen(len(src)))
	n, err := enc.Decode(buf, src)
	return buf[:n], err
}

var _ Cipher = (*Base32)(nil)

type Base32 struct {
	Encoder *base32.Encoding
}

func (b Base32) getEncoder() *base32.Encoding {
	if b.Encoder == nil {
		return base32.StdEncoding
	}
	return b.Encoder
}

func (b Base32) Encrypt(src []byte) ([]byte, error) {
	enc := b.getEncoder()
	buf := make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(buf, src)
	return buf, nil
}

func (b Base32) Decrypt(src []byte) ([]byte, error) {
	enc := b.getEncoder()
	buf := make([]byte, enc.DecodedLen(len(src)))
	n, err := enc.Decode(buf, src)
	return buf[:n], err
}

var _ Cipher = (*HEX)(nil)

type HEX struct{}

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
