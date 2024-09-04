//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/xanygo/anygo/xcfg/internal/hook"
)

// Hook 辅助类，用于在解析配置文件内容是，提前先对配置的内容进行预处理
type Hook interface {
	// Name 名称，不可为空
	// 每个 Hook 应返回唯一的名称，若重名会注册失败
	Name() string

	// Execute 对读取的配置内容加工的逻辑
	Execute(ctx context.Context, p *HookParam) (output []byte, err error)
}

// HookParam Hook 的参数
type HookParam struct {
	Cfg      *Configure // 当前 Cfg 对象
	FileExt  string     // 文件类型后缀，如 .toml,.json
	FilePath string     // 文件路径。当直接解析内容时，为空字符串
	Content  []byte     // 文件内容
}

var defaultHooks hooks = []Hook{
	newHook("template", (&hook.Template{}).Hook),
	newHook("env", hook.OsEnvVars),
	newHook("xattr", hook.XAttrVars),
}

type hooks []Hook

func (hs *hooks) Add(h Hook) error {
	if len(h.Name()) == 0 {
		return errors.New("hook.Name is empty, not allow")
	}

	for _, h1 := range *hs {
		if h.Name() == h1.Name() {
			return fmt.Errorf("hook=%q already exists", h.Name())
		}
	}
	*hs = append(*hs, h)
	return nil
}

func (hs hooks) Execute(ctx context.Context, p *HookParam) (output []byte, err error) {
	if len(hs) == 0 {
		return p.Content, nil
	}
	content := bytes.Clone(p.Content)

	for _, hk := range hs {
		p.Content = content
		content, err = hk.Execute(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("hook=%q has error:%w", hk.Name(), err)
		}
	}
	return content, err
}

type hookTpl struct {
	fn   hook.Func
	name string
}

func (h *hookTpl) Name() string {
	return h.name
}

func (h *hookTpl) Execute(ctx context.Context, p *HookParam) (output []byte, err error) {
	return h.fn(ctx, p.FilePath, p.Content)
}

func newHook(name string, fn hook.Func) Hook {
	return &hookTpl{
		name: name,
		fn:   fn,
	}
}
