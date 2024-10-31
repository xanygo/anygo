//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcodec

type TranscoderFunc func([]byte) ([]byte, error)

type TranscoderFuncs []func([]byte) ([]byte, error)

func (ts TranscoderFuncs) Transcoding(data []byte) (result []byte, err error) {
	result = data
	for _, f := range ts {
		result, err = f(result)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func EncoderWithTranscoder(enc Encoder, trans TranscoderFunc) Encoder {
	return EncodeFunc(func(a any) ([]byte, error) {
		data, err := enc.Encode(a)
		if err != nil {
			return nil, err
		}
		return trans(data)
	})
}

func DecoderWithTranscoder(dec Decoder, trans TranscoderFunc) Decoder {
	return DecodeFunc(func(data []byte, obj any) error {
		nd, err := trans(data)
		if err != nil {
			return err
		}
		return dec.Decode(nd, obj)
	})
}
