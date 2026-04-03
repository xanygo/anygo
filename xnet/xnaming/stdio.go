//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-01

package xnaming

import (
	"context"

	"github.com/xanygo/anygo/internal/ztypes"
	"github.com/xanygo/anygo/xnet"
)

var _ Naming = (*Stdio)(nil)

// Stdio 和 Command 子进程的 stdin/stdout 通行模式
type Stdio struct{}

func (c *Stdio) Scheme() string {
	return xnet.NetworkStdio
}

// Lookup 解析 cmd 配置
//
// address 是编码后的 cmd 和 args，具体格式为：
// json_encode({Path:"cmd_path",Args:["arg1","arg2"],Dir:"工作目录")
func (c *Stdio) Lookup(ctx context.Context, idc string, address string) ([]xnet.AddrNode, error) {
	cmd := &ztypes.ServiceCommand{}
	err := cmd.LoadFromStr(address)
	if err != nil {
		return nil, err
	}
	return []xnet.AddrNode{
		{
			HostPort: address,
			Addr:     xnet.NewAddr(xnet.NetworkStdio, address),
		},
	}, nil
}

func init() {
	MustRegister(&Stdio{})
}
