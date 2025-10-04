//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xvalidator"
)

// NewDefault 创建一个新的配置解析实例
// 会注册默认的配置解析方法和辅助方法
func NewDefault() *Configure {
	cfg := &Configure{}
	for _, pair := range defaultDecoders {
		if err := cfg.WithDecoder(pair.Name, pair.Fn); err != nil {
			panic(fmt.Sprintf("WithParser(%q) err=%s", pair.Name, err))
		}
	}

	for _, h := range defaultHooks {
		if err := cfg.WithHook(h); err != nil {
			panic(fmt.Sprintf("RegisterInterceptor(%q) err=%s", h.Name(), err))
		}
	}
	return cfg
}

type Configure struct {
	// Dir 配置的根目录，可选，当为空时，会使用 xattr.ConfDir()
	Dir string

	ctx       context.Context
	validator xvalidator.Validator
	decoders  map[string]xcodec.Decoder
	exts      []string // 支持的文件后缀，如 []string{".json",".toml"}
	hooks     hooks
}

func (c *Configure) Parse(path string, obj any) error {
	absPath, err := c.confFileAbsPath(path)
	if err != nil {
		return err
	}
	return c.parseByAbsPath(absPath, obj)
}

func (c *Configure) getDir() string {
	if c.Dir != "" {
		return c.Dir
	}
	return xattr.ConfDir()
}

func (c *Configure) confFileAbsPath(path string) (string, error) {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return filepath.Abs(path)
	}

	if filepath.IsAbs(path) {
		return path, nil
	}

	fp := filepath.Join(c.getDir(), path)

	if !fileExists(fp) {
		if fp1, err := filepath.Abs(path); err == nil && fileExists(fp1) {
			return fp1, nil
		}
	}
	return fp, nil
}

func fileExists(fp string) bool {
	info, err := os.Stat(fp)
	return err == nil && !info.IsDir()
}

func (c *Configure) parseByAbsPath(absPath string, obj any) error {
	if len(c.decoders) == 0 {
		return errors.New("no parser")
	}

	return c.readConfDirect(absPath, obj)
}

func (c *Configure) realConfPath(confPath string) (path string, ext string, err error) {
	fileExt := filepath.Ext(confPath)
	info, err1 := os.Stat(confPath)

	if err1 == nil && !info.IsDir() {
		return confPath, fileExt, nil
	}

	notExist := err1 != nil && errors.Is(err1, fs.ErrNotExist)
	isDir := err1 == nil && info.IsDir()

	// fileExt == "" 是为了兼容存在同名目录的情况
	if (notExist || isDir || fileExt == "") && !slices.Contains(c.exts, fileExt) {
		for i := 0; i < len(c.exts); i++ {
			ext2 := c.exts[i]
			name2 := confPath + ext2
			info2, err2 := os.Stat(name2)
			if err2 == nil && !info2.IsDir() {
				return name2, ext2, nil
			}
		}
	}
	if err1 != nil {
		return "", "", err1
	}
	return "", "", fmt.Errorf("cannot get real path for %q", confPath)
}

func (c *Configure) readConfDirect(path string, obj any) error {
	realPath, fileExt, err := c.realConfPath(path)
	if err != nil {
		return err
	}
	content, errIO := os.ReadFile(realPath)
	if errIO != nil {
		return errIO
	}
	err2 := c.parseBytes(realPath, fileExt, content, obj)
	if err2 == nil {
		return nil
	}
	return fmt.Errorf("parser %q failed: %w", realPath, err2)
}

func (c *Configure) context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *Configure) ParseBytes(ext string, content []byte, obj any) error {
	return c.parseBytes("", ext, content, obj)
}

func (c *Configure) parseBytes(confPath string, fileExt string, content []byte, obj any) error {
	parser, hasParser := c.decoders[fileExt]
	if len(fileExt) == 0 || !hasParser {
		err1 := fmt.Errorf("fileExt %q is not supported yet", fileExt)
		if confPath == "" {
			return err1
		}
		return fmt.Errorf("cannot parser %q: %w", confPath, err1)
	}

	p := &HookParam{
		FileExt:  fileExt,
		Cfg:      c,
		FilePath: confPath,
		Content:  content,
	}

	contentNew, errHook := c.hooks.Execute(c.context(), p)

	if errHook != nil {
		return errHook
	}

	if errParser := xcodec.Decode(parser, contentNew, obj); errParser != nil {
		return fmt.Errorf("%w, config content=\n%s", errParser, string(contentNew))
	}

	return xvalidator.ValidateWith(c.validator, obj)
}

func (c *Configure) Exists(path string) bool {
	p, err := c.confFileAbsPath(path)
	if err != nil {
		return false
	}

	info, err := os.Stat(p)
	if err == nil && !info.IsDir() {
		return true
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return false
	}
	for ext := range c.decoders {
		info1, err1 := os.Stat(p + ext)
		if err1 == nil && !info1.IsDir() {
			return true
		}
	}
	return false
}

func (c *Configure) WithDecoder(ext string, fn xcodec.Decoder) error {
	if c.decoders == nil {
		c.decoders = make(map[string]xcodec.Decoder, len(defaultDecoders))
	}
	if _, has := c.decoders[ext]; has {
		return fmt.Errorf("parser=%q already exists", ext)
	}
	c.decoders[ext] = fn
	c.exts = append(c.exts, ext)
	return nil
}

func (c *Configure) MustWithDecoder(ext string, fn xcodec.Decoder) {
	if err := c.WithDecoder(ext, fn); err != nil {
		panic(err)
	}
}

// WithHook 注册新的 Hook，若出现重名会注册失败
func (c *Configure) WithHook(hooks ...Hook) error {
	for _, h := range hooks {
		if err := c.hooks.Add(h); err != nil {
			return err
		}
	}
	return nil
}

// MusWithHook 注册新的 Hook, 若出现重名等异常情况会 panic
func (c *Configure) MusWithHook(hooks ...Hook) {
	if err := c.WithHook(hooks...); err != nil {
		panic(err)
	}
}

func (c *Configure) Clone() *Configure {
	return &Configure{
		Dir:       c.Dir,
		decoders:  maps.Clone(c.decoders),
		exts:      slices.Clone(c.exts),
		validator: c.validator,
		hooks:     slices.Clone(c.hooks),
	}
}

func (c *Configure) CloneWithContext(ctx context.Context) *Configure {
	c1 := c.Clone()
	c1.ctx = ctx
	return c1
}

func (c *Configure) CloneWithHook(hooks ...Hook) *Configure {
	c1 := c.Clone()
	for _, h := range hooks {
		c1.MusWithHook(h)
	}
	return c1
}
