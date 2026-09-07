//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcipher

type (
	// Encryptor 加密
	Encryptor interface {
		Encrypt(src []byte) ([]byte, error)
	}

	// Decrypter 解密
	Decrypter interface {
		Decrypt(src []byte) ([]byte, error)
	}

	Cipher interface {
		Encryptor
		Decrypter
	}
)

type (
	IDEncryptor interface {
		Encryptor
		ID() []byte
	}

	IDDecrypter interface {
		Decrypter
		ID() []byte
	}
)

type EncryptFunc func([]byte) ([]byte, error)

func (fn EncryptFunc) Encrypt(src []byte) ([]byte, error) {
	return fn(src)
}

type DecryptFunc func([]byte) ([]byte, error)

func (fn DecryptFunc) Decrypt(src []byte) ([]byte, error) {
	return fn(src)
}

type Encryptors []Encryptor

func (es Encryptors) Encrypt(src []byte) (out []byte, err error) {
	out = src
	for _, cp := range es {
		out, err = cp.Encrypt(out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

type Decrypters []Decrypter

func (ds Decrypters) Decrypt(src []byte) (out []byte, err error) {
	out = src
	for _, cp := range ds {
		out, err = cp.Decrypt(out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Ciphers 多个 Cipher 的组合。
// 可以联合在一起链式工作，在 Encrypt 的时候，会依次正序调用。在 Decrypt 的时候，会依次倒序调用。
type Ciphers []Cipher

func (cs Ciphers) Encrypt(src []byte) (out []byte, err error) {
	out = src
	for _, cp := range cs {
		out, err = cp.Encrypt(out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (cs Ciphers) Decrypt(src []byte) (out []byte, err error) {
	out = src
	for i := len(cs) - 1; i >= 0; i-- {
		out, err = cs[i].Decrypt(out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func NewCipher(enc EncryptFunc, dec DecryptFunc) Cipher {
	return cipherTPL{enc, dec}
}

type cipherTPL [2]func(src []byte) ([]byte, error)

func (cs cipherTPL) Encrypt(src []byte) ([]byte, error) {
	return cs[0](src)
}

func (cs cipherTPL) Decrypt(src []byte) ([]byte, error) {
	return cs[1](src)
}
