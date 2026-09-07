//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-06

package xbase

import (
	"bytes"
	"encoding/base64"
	"os"
)

func ReadBase64File(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content = bytes.TrimSpace(content)
	content = bytes.ReplaceAll(content, []byte("\n"), nil)
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(content)))
	n, err1 := base64.StdEncoding.Decode(dbuf, content)
	return dbuf[:n], err1
}
