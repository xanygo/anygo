//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-30

package xcodec

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
)

func GZipCompress(src []byte) ([]byte, error) {
	bf := &bytes.Buffer{}
	cp := gzip.NewWriter(bf)
	_, err := cp.Write(src)
	if err != nil {
		return nil, err
	}
	if err = cp.Flush(); err != nil {
		return nil, err
	}
	if err = cp.Close(); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func GZipDecompress(src []byte) ([]byte, error) {
	rd, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	bf, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if err = rd.Close(); err != nil {
		return nil, err
	}
	return bf, nil
}

func ZLibCompress(src []byte) ([]byte, error) {
	bf := &bytes.Buffer{}
	cp := zlib.NewWriter(bf)
	_, err := cp.Write(src)
	if err != nil {
		return nil, err
	}
	if err = cp.Flush(); err != nil {
		return nil, err
	}
	if err = cp.Close(); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func ZLibDecompress(src []byte) ([]byte, error) {
	rd, err := zlib.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	bf, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if err = rd.Close(); err != nil {
		return nil, err
	}
	return bf, nil
}

func FlateCompress(src []byte) ([]byte, error) {
	bf := &bytes.Buffer{}
	cp, err := flate.NewWriter(bf, 1)
	if err != nil {
		return nil, err
	}
	_, err = cp.Write(src)
	if err != nil {
		return nil, err
	}
	if err = cp.Flush(); err != nil {
		return nil, err
	}
	if err = cp.Close(); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func FlateDecompress(src []byte) ([]byte, error) {
	rd := flate.NewReader(bytes.NewReader(src))
	bf, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if err = rd.Close(); err != nil {
		return nil, err
	}
	return bf, nil
}
