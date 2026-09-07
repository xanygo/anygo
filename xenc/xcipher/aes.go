package xcipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
)

var _ Cipher = (*AesBlock)(nil)

// AesBlock AES 加解密
type AesBlock struct {
	// Key 秘钥，必填，若长度为 16, 24, 32 则直接使用，否则使用 md5 值
	Key string

	// IV 初始化向量，可选，当不为空时，长度应为 16
	// 当为空时，会基于 key 生成
	IV string

	base *cryptoBlockBase

	once sync.Once
}

func (a *AesBlock) init() {
	a.base = &cryptoBlockBase{
		Key:          a.Key,
		IV:           a.IV,
		NewCipher:    aes.NewCipher,
		NewEncrypter: cipher.NewCBCEncrypter,
		NewDecrypter: cipher.NewCBCDecrypter,
		BlockSize:    aes.BlockSize,
	}
	a.base.init()
}

func (a *AesBlock) Name() string {
	return "AesBlock"
}

func (a *AesBlock) ID() []byte {
	a.once.Do(a.init)
	return a.base.ID()
}

func (a *AesBlock) Encrypt(src []byte) ([]byte, error) {
	a.once.Do(a.init)
	return a.base.Encrypt(src)
}

func (a *AesBlock) Decrypt(src []byte) ([]byte, error) {
	a.once.Do(a.init)
	return a.base.Decrypt(src)
}

type cryptoBlockBase struct {
	Key          string
	IV           string
	NewCipher    func([]byte) (cipher.Block, error)
	NewEncrypter func(b cipher.Block, iv []byte) cipher.BlockMode
	NewDecrypter func(b cipher.Block, iv []byte) cipher.BlockMode
	BlockSize    int

	key []byte
	iv  []byte
	id  []byte // key 和 iv 的签名
	err error
}

var errNoKey = errors.New("key is required")

func (base *cryptoBlockBase) init() {
	if len(base.Key) == 0 {
		base.err = errNoKey
		return
	}
	base.key = []byte(base.Key)
	switch len(base.key) {
	case 16, 24, 32:
		// 直接使用设置的 Key
	default:
		by1 := md5.Sum([]byte("anygo#" + base.Key))
		base.key = []byte(hex.EncodeToString(by1[:]))
	}

	if len(base.IV) == base.BlockSize {
		base.iv = []byte(base.IV)
	} else {
		by2 := md5.Sum([]byte(base.Key + "|xanygo|" + base.IV))
		base.iv = by2[:base.BlockSize]
	}

	hh := hmac.New(sha256.New, base.key)
	_, err := hh.Write(base.iv)
	if err != nil {
		base.err = err
		return
	}
	base.id = hh.Sum(nil)
}

func (base *cryptoBlockBase) ID() []byte {
	return base.id
}

func (base *cryptoBlockBase) Encrypt(src []byte) ([]byte, error) {
	if base.err != nil {
		return nil, base.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	block, err := base.NewCipher(base.key)
	if err != nil {
		panic(err)
	}
	enc := base.NewEncrypter(block, base.iv)
	bf1, bf2, padding := base.padding(src)
	cipherText := make([]byte, len(src)+padding)
	if len(bf1) > 0 {
		enc.CryptBlocks(cipherText, bf1)
	}
	if padding == 0 {
		return cipherText, nil
	}
	enc.CryptBlocks(cipherText[len(bf1):], bf2)
	return cipherText, nil
}

func (base *cryptoBlockBase) padding(src []byte) ([]byte, []byte, int) {
	dlt := len(src) % base.BlockSize
	if dlt == 0 {
		return src, nil, 0
	}
	index := len(src) / base.BlockSize * base.BlockSize
	pad := make([]byte, base.BlockSize)
	copy(pad, src[index:])
	return src[:index], pad, base.BlockSize - dlt
}

func (base *cryptoBlockBase) Decrypt(src []byte) ([]byte, error) {
	if base.err != nil {
		return nil, base.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	if len(src)%base.BlockSize != 0 {
		return nil, errors.New("invalid encrypt data len")
	}
	block, err := base.NewCipher(base.key)
	if err != nil {
		panic(err)
	}
	dec := base.NewDecrypter(block, base.iv)
	plainText := make([]byte, len(src))
	dec.CryptBlocks(plainText, src)
	for i := len(plainText) - 1; i >= 0; i-- {
		if plainText[i] != 0 {
			return plainText[:i+1], nil
		}
	}
	return plainText, nil
}

var _ Cipher = (*AesOFB)(nil)

type AesOFB struct {
	// Key 必填，加密秘钥，若长度为 16, 24, 32 会直接使用，否则会使用 md5 值
	Key string

	// IV 初始化向量，可选，当不为空时，长度应为 16
	// 当为空时，会基于 key 生成
	IV string

	key []byte
	iv  []byte
	id  []byte // key 和 iv 的签名

	once sync.Once
	err  error
}

func (a *AesOFB) Name() string {
	return "AesOFB"
}

func (a *AesOFB) init() {
	if len(a.Key) == 0 {
		a.err = errNoKey
		return
	}
	key := []byte(a.Key)
	switch len(key) {
	case 16, 24, 32:
		// 直接使用设置的 Key
	default:
		by1 := md5.Sum([]byte("anygo#" + a.Key))
		key = []byte(hex.EncodeToString(by1[:]))
	}
	a.key = key

	if len(a.IV) == aes.BlockSize {
		a.iv = []byte(a.IV)
	} else {
		by2 := md5.Sum([]byte(a.Key + "|xanygo|" + a.IV))
		a.iv = by2[:aes.BlockSize]
	}

	hh := hmac.New(sha256.New, a.key)
	_, err := hh.Write(a.iv)
	if err != nil {
		a.err = err
		return
	}
	a.id = hh.Sum(nil)
}

func (a *AesOFB) ID() []byte {
	a.once.Do(a.init)
	return a.id
}

func (a *AesOFB) Encrypt(src []byte) ([]byte, error) {
	if a.err != nil {
		return nil, a.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	a.once.Do(a.init)
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, a.iv)
	bf := &bytes.Buffer{}
	w := &cipher.StreamWriter{S: stream, W: bf}
	_, err = w.Write(src)
	if err != nil {
		_ = w.Close()
		return nil, err
	}
	err = w.Close()
	return bf.Bytes(), err
}

func (a *AesOFB) Decrypt(src []byte) ([]byte, error) {
	if a.err != nil {
		return nil, a.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	a.once.Do(a.init)
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, a.iv)
	rd := &cipher.StreamReader{S: stream, R: bytes.NewReader(src)}
	return io.ReadAll(rd)
}

var _ Cipher = (*AesGCM)(nil)

type AesGCM struct {
	// Key 必填，加密秘钥，若长度为 16, 24, 32 会直接使用，否则会使用 md5 值
	Key string

	key []byte

	once sync.Once
	err  error
}

func (a *AesGCM) Name() string {
	return "AesGCM"
}

func (a *AesGCM) init() {
	if len(a.Key) == 0 {
		a.err = errNoKey
		return
	}
	key := []byte(a.Key)
	switch len(key) {
	case 16, 24, 32:
		// 直接使用设置的 Key
	default:
		by1 := md5.Sum([]byte("anygo#" + a.Key))
		key = []byte(hex.EncodeToString(by1[:]))
	}
	a.key = key
}

func (a *AesGCM) Encrypt(src []byte) ([]byte, error) {
	if a.err != nil {
		return nil, a.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	a.once.Do(a.init)
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, src, nil)

	return append(nonce, ciphertext...), nil
}

func (a *AesGCM) Decrypt(src []byte) ([]byte, error) {
	if a.err != nil {
		return nil, a.err
	}
	if len(src) == 0 {
		return nil, nil
	}
	a.once.Do(a.init)
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(src) < nonceSize {
		return nil, fmt.Errorf("invalid data, input.len=%d, expect >= %d", len(src), nonceSize)
	}

	nonce := src[:nonceSize]
	ciphertext := src[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
