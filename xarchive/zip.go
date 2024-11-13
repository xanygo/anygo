//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xarchive

import (
	"archive/zip"
	"path"
	"strings"
)

func ZipFileNames(rd *zip.Reader, strip uint) []string {
	result := make([]string, 0, len(rd.File))
	for _, f := range rd.File {
		np := stripComponents(f.Name, strip)
		if np != "" {
			result = append(result, np)
		}
	}
	return result
}

func stripComponents(p string, n uint) string {
	if n == 0 {
		return p
	}
	sc := int(n)
	ps := strings.Split(path.Clean(p), "/")
	if len(ps) < sc {
		return ""
	}
	return path.Join(ps[sc:]...)
}
