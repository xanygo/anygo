package xcipher

import "encoding/base32"

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

func (b Base32) Name() string {
	return "Base32"
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
