package zloader

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xenc"
	"github.com/xanygo/anygo/xenc/xbase"
	"github.com/xanygo/anygo/xenc/xcipher"
	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xenc/xcompress"
	"github.com/xanygo/anygo/xmap"
)

func ParserCodec(cfg map[string]any, name string, defaultValue xenc.Codec) (xenc.Codec, error) {
	sub, err := xmap.GetMap(cfg, name)
	if err != nil {
		return nil, err
	}
	if len(sub) == 0 {
		return defaultValue, nil
	}
	typ, err := xmap.GetString(sub, "Type")
	if err != nil {
		return nil, err
	}
	var codec xenc.Codec
	if typ == "" {
		codec = defaultValue
	} else {
		codec, err = xcodec.Find(typ)
		if err != nil {
			return nil, err
		}
	}
	cipher, err := ParserCipher(sub)
	if err != nil {
		return nil, err
	}
	if cipher == nil {
		return codec, nil
	}
	return xenc.CodecWithCipher(codec, cipher), nil
}

func ParserCipher(cfg map[string]any) (xenc.Cipher, error) {
	value, err1 := xmap.GetMap(cfg, "Cipher")
	if err1 == nil {
		if len(value) == 0 {
			return nil, nil
		}
		return parserOneCipher(value)
	}
	items, err2 := xmap.GetSlice[string, any, any](cfg, "Cipher")
	if err2 != nil {
		return nil, fmt.Errorf("invalid Cipher Type %w, should slice or map", err2)
	}
	if len(items) == 0 {
		return nil, nil
	}
	var cs xcipher.Ciphers
	err3 := zreflect.RangeSlice(items, func(item map[string]any) error {
		cp, err := parserOneCipher(item)
		if err != nil {
			return err
		}
		cs = append(cs, cp)
		return nil
	})
	return cs, err3
}

func parserOneCipher(cfg map[string]any) (xenc.Cipher, error) {
	typ, err := xmap.GetString(cfg, "Type")
	if err != nil {
		return nil, err
	}
	if typ == "" {
		return nil, errors.New("miss cipher 'Type'")
	}
	if strings.EqualFold(typ, "No") {
		return nil, nil
	}

	var key string
	if strings.HasPrefix(typ, "Aes") {
		key, err = xmap.GetString(cfg, "Key")
		if err != nil {
			return nil, err
		}
		if key == "" {
			return nil, errors.New("missing cipher 'Key'")
		}
	}

	iv, err3 := xmap.GetString(cfg, "IV")
	if err3 != nil {
		return nil, err3
	}

	switch typ {
	case "AesOFB":
		return &xcipher.AesOFB{
			Key: key,
			IV:  iv,
		}, nil
	case "AesGCM":
		return &xcipher.AesGCM{
			Key: key,
		}, nil
	case "AesBlock":
		return &xcipher.AesBlock{
			Key: key,
			IV:  iv,
		}, nil
	case "GZip":
		// 将压缩也当作压缩算法
		cp := xenc.NewCipher(xenc.EncryptFunc(xcompress.GZipCompress), xenc.DecryptFunc(xcompress.GZipDecompress))
		return cp, nil
	case "Base64":
		return &xcipher.Base64{
			Encoder: base64.RawURLEncoding,
		}, nil
	case "Base62":
		cp := xenc.NewCipher(xenc.TextMapperFunc(xbase.Base62.Encode).Encode, xenc.DecryptFunc(xbase.Base62.Decode))
		return cp, nil
	case "Base58":
		cp := xenc.NewCipher(xenc.TextMapperFunc(xbase.Base58.Encode).Encrypt, xenc.DecryptFunc(xbase.Base58.Decode))
		return cp, nil
	case "Base36":
		cp := xenc.NewCipher(xenc.TextMapperFunc(xbase.Base36.Encode).Encrypt, xenc.DecryptFunc(xbase.Base36.Decode))
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupport Cipher.Type %q", typ)
	}
}
