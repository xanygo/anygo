//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/xio/xfs"
	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*FileStore)(nil)

// FileStore 解析文件，如  file@server_list.ns
//
//	文件内部格式如：
//	# user service node list
//	127.0.0.1:8000
//	127.0.0.2:8000
//
//	# backup node
//	10.0.0.1:9000  # comment
type FileStore struct {
	cache *xmap.LRUReader[string, *cachedFile]
	once  sync.Once
}

func (f *FileStore) Scheme() string {
	return "file"
}

func (f *FileStore) init() {
	f.cache = &xmap.LRUReader[string, *cachedFile]{
		New: func(key string) *cachedFile {
			return &cachedFile{
				path: key,
				file: &xfs.CachedReader{
					Path: key,
				},
			}
		},
		Store: xmap.NewLRU[string, *cachedFile](1024),
	}
}

func (f *FileStore) Lookup(ctx context.Context, idc string, filename string) ([]xnet.AddrNode, error) {
	f.once.Do(f.init)
	return f.cache.Get(filename).fetch()
}

func init() {
	MustRegister(&FileStore{})
}

type cachedFile struct {
	path  string
	file  *xfs.CachedReader
	addrs []xnet.AddrNode
	err   error
}

func (cf *cachedFile) fetch() ([]xnet.AddrNode, error) {
	content, fromCache, err := cf.file.ReadFile()
	if err != nil {
		return nil, err
	}
	if !fromCache {
		cf.addrs, cf.err = cf.parser(content)
	}
	return cf.addrs, cf.err
}

func (cf *cachedFile) parser(content []byte) ([]xnet.AddrNode, error) {
	lines := strings.Split(string(content), "\n")
	nodes := make([]xnet.AddrNode, 0, len(lines))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, line := range lines {
		line, _, _ = strings.Cut(line, "#") // 去掉 # 注释的内容
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ns, err := LookupRaw(ctx, "", line)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, ns...)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no hostPort found in file %s", cf.path)
	}
	return nodes, nil
}
