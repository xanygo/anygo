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
//
// 参数说明：
//
//	b *Bundle:    将数据加载到此资源包
//	f fs.FS:      本地化资源文件所在目录
//	root string:  文件根目录，一般使用 "."
//	ext:          资源文件后缀，如 .json 或者 .yml
//	decoder:      和 ext 对应的资源文件解析器
//
// 资源文件目录结构：
//
//	.
//	├── en                           <--- 语言：英文。第一级目录为语言名称，如 zh、en 等
//	│   ├── home.json          <---  访问这里的资源 (ns=home key=<key>)
//	│   └── ns1                <--- 可以包含子目录
//	│       ├── index.json
//	│       └── z              <--- 可以包含多层级子目录
//	│           └── z.json
//	└── zh                        <--- 语言：中文。 实际使用时，应确保不同语言下，目录结构一致，并确保所有的所有的 key 都包含
//		└── home.json
//
// en/home.json 文件格式：
//
//	[
//		{
//			"Key":"k1",
//			"Other": "hello"
//		},                     <--- 这是一条国际化资源，字段列表详见 Message 的文档
//		{
//			"Key":"k2",
//			"Other": "world"
//		}
//	]
//
//
//	namespace=目录+文件名(不包含后缀)，如 namespace=home 或者 namespace="ns1/index" .
//	在后续使用的时候，namespace 也可以拼接到 key 里作为前缀，如 key="home/k1",然后 namespace 传空字符串：
//	{{ xi "home/k1" }} 或者  {{ xi "ns1/index/k2" }}
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
