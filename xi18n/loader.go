//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-07

package xi18n

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/xanygo/anygo/xcodec"
)

// LoadFS 加载本地化资源到 Bundle 里去
func LoadFS(b *Bundle, f fs.FS, root string, ext string, decoder xcodec.Decoder) error {
	return fs.WalkDir(f, root, func(fileName string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := path.Clean(fileName)
		fileExt := path.Ext(name)
		if fileExt != ext {
			return nil
		}

		content, err := fs.ReadFile(f, fileName)
		if err != nil {
			return err
		}
		content = bytes.TrimSpace(content)
		if len(content) == 0 {
			return nil
		}
		lang, nsName, ok := strings.Cut(name, "/")
		if !ok {
			return fmt.Errorf("invalid path: %q", fileName)
		}
		lz := b.MustLocalize(Language(lang))

		nameSpace, _ := strings.CutSuffix(nsName, fileExt)

		var msgs []*Message
		if err = decoder.Decode(content, &msgs); err != nil {
			return err
		}
		return lz.Add(nameSpace, msgs...)
	})
}
