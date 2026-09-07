package xcipher

import "encoding/base64"

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

func (b Base64) Name() string {
	return "Base64"
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
