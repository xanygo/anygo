//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xnaming

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*FileStore)(nil)

// FileStore 解析文件，如  file://server_list.ns
//
//	文件内部格式如：
//	# user service node list
//	127.0.0.1:8000
//	127.0.0.2:8000
//
//	# backup node
//	10.0.0.1:9000  # comment
type FileStore struct {
}

func (f *FileStore) Scheme() string {
	return "file"
}

func (f *FileStore) Lookup(ctx context.Context, idc string, fileName string, param url.Values) ([]xnet.AddrNode, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	nodes := make([]xnet.AddrNode, 0, len(lines))
	for _, line := range lines {
		line, _, _ = strings.Cut(line, "#") // 去掉 # 注释的内容
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		_, _, err = net.SplitHostPort(line)
		if err != nil {
			return nil, err
		}
		node := xnet.AddrNode{
			HostPort: line,
			Addr:     xnet.NewAddr("tcp", line),
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no hostPort found in file %s", fileName)
	}
	return nodes, nil
}

func init() {
	MustRegister(&FileStore{})
}
