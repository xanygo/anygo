//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xattr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xanygo/anygo/xenc"
)

// FileConfig 应用的主配置文件，一般是 conf/app.yml 或 conf/app.json
type FileConfig struct {
	// Listen 监听的端口信息，可选
	Listen map[string]string `yaml:"Listen"`

	// AppName 应用名称，可选
	AppName string

	// IDC 应用运行机房，可选
	IDC string `yaml:"IDC"`

	// RunMode 运行模式，可选
	// 可选值：product （生成环境），debug （调试模式）
	RunMode string `yaml:"RunMode"`

	// RootDir 应用根目录，可选
	RootDir string `yaml:"RootDir"`

	// DataDir 应用数据目录，可选
	DataDir string `yaml:"DataDir"`

	// LogDir 应用日志目录，可选
	LogDir string `yaml:"LogDir"`

	// ConfDir 应用配置文件目录，可选
	ConfDir string `yaml:"ConfDir"`

	// TempDir 应用临时文件目录，可选
	TempDir string `yaml:"TempDir"`

	// Tags 应用部署及运行标签，可选，可设置多个
	// 例如：prod、web、api、blue、us-west
	Tags []string `yaml:"Tags"`

	// Other 其他项，可选
	Other map[string]any `yaml:"Other"`

	// EnsureDirs 应用启动时需要确保存在的目录，可选
	// 目录不存在时自动创建
	EnsureDirs []string `yaml:"EnsureDirs"`

	// SelfPath 配置自己的路径
	SelfPath string
}

func (c FileConfig) MustGetListen(name string) string {
	value, err := c.GetListen(name)
	if err != nil {
		panic(err)
	}
	return value
}

func (c FileConfig) GetListen(name string) (string, error) {
	if len(c.Listen) == 0 {
		return "", fmt.Errorf("empty Listen in %s", c.SelfPath)
	}
	v, ok := c.Listen[name]
	if ok {
		return strings.TrimSpace(v), nil
	}
	return "", fmt.Errorf("not found Listen[%q] in %s", name, c.SelfPath)
}

func (c FileConfig) GetOther(name string) (value any, ok bool) {
	if len(c.Other) == 0 {
		return nil, false
	}
	value, ok = c.Other[name]
	return value, ok
}

func (c FileConfig) getAppName() string {
	if c.AppName != "" {
		return c.AppName
	}
	root := c.getRootDir()
	if root != "" {
		return filepath.Base(root)
	}
	wd, _ := os.Getwd()
	if wd != "" {
		return filepath.Base(wd)
	}
	return ""
}

func (c FileConfig) getRootDir() string {
	if c.RootDir != "" {
		return c.RootDir
	}
	if c.SelfPath == "" {
		return ""
	}
	dir := filepath.Dir(c.SelfPath)
	// 支持这样的目录结构：
	// xxx/conf/app.yml          <-- 普通
	// xxx/conf/product/app.yml  <-- 子目录中放主配置文件
	//
	// xxx/conf_product/app.yml  <-- 不同环境的配置文件以 conf_xxx 命名
	// xxx/conf_dev/app.yml
	for i := 0; i < 2; i++ {
		base := filepath.Base(dir)
		if base == "conf" || strings.HasPrefix(base, "conf_") {
			return filepath.Dir(dir)
		}
		dir = filepath.Dir(dir)
	}
	return dir
}

func (c FileConfig) getConfDir() string {
	if c.ConfDir != "" {
		return c.ConfDir
	}
	if c.SelfPath != "" {
		return filepath.Dir(c.SelfPath)
	}
	if c.RootDir != "" {
		return filepath.Join(c.RootDir, "conf")
	}
	return ""
}

func (c FileConfig) SetTo(attr *Attribute) {
	if name := c.getAppName(); name != "" {
		attr.SetAppName(c.AppName)
	}
	if c.IDC != "" {
		attr.SetIDC(c.IDC)
	}
	switch c.RunMode {
	case ModeProduct.String():
		attr.SetRunMode(ModeProduct)
	case ModeDebug.String():
		attr.SetRunMode(ModeDebug)
	}
	if root := c.getRootDir(); root != "" {
		attr.SetRootDir(root)
	}
	if dir := c.getConfDir(); dir != "" {
		attr.SetConfDir(dir)
	}
	if c.DataDir != "" {
		attr.SetDataDir(c.DataDir)
	}
	if c.TempDir != "" {
		attr.SetTempDir(c.TempDir)
	}
	if c.LogDir != "" {
		attr.SetLogDir(c.LogDir)
	}
	for key, value := range c.Other {
		attr.SetOther(key, value)
	}
	attr.AppendTags(c.Tags...)
}

func (c FileConfig) SetToDefault() {
	c.SetTo(Default)
}

var _ xenc.UnmarshalExtra = FileConfig{}

// NeedDecodeExtra 将其他未定义的字段全部解析的 Other 这个 map 里去
func (c FileConfig) NeedDecodeExtra() string {
	return "Other"
}

// ensureDirectories 检查 EnsureDirs 中的目录是否存在，不存在则自动创建。
func (c FileConfig) ensureDirectories() error {
	var errs []error
	for _, dir := range c.EnsureDirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dir = c.realDir(dir)
		dir = filepath.Clean(dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			errs = append(errs, fmt.Errorf("create directory %q: %w", dir, err))
		}
	}
	return errors.Join(errs...)
}

func (c FileConfig) realDir(dir string) string {
	if !strings.Contains(dir, "{") {
		return dir
	}
	return strings.NewReplacer(
		"{RootDir}", RootDir(),
		"{DataDir}", DataDir(),
		"{ConfDir}", ConfDir(),
		"{TempDir}", TempDir(),
		"{LogDir}", LogDir(),
		"{IDC}", IDC(),
		"{RunMode}", RunMode().String(),
	).Replace(dir)
}

func ParserFileConfig(path string, parser func(string, any) error) (*FileConfig, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cfg := &FileConfig{}
	if err := parser(path, cfg); err != nil {
		return nil, err
	}
	cfg.SelfPath = path
	return cfg, nil
}

var mainCfg FileConfig
var mainCfgInited bool

// AppMain 应用主配置文件，在使用前，需要先使用 InitAppMainCfg 或者 MustInitAppMain 加载
func AppMain() FileConfig {
	if !mainCfgInited {
		panic("should InitAppMain or MustInitAppMain first")
	}
	return mainCfg
}

// InitAppMain 初始化应用主配置文件,并处理：
//  1. 应用切换到 RootDir()
//  2. 检查 EnsureDirs 目前，若不存在则创建
func InitAppMain(path string, parser func(string, any) error) (FileConfig, error) {
	cfg, err := ParserFileConfig(path, parser)
	if err != nil {
		return FileConfig{}, err
	}
	Init(AppName(), cfg.getRootDir())
	mainCfg = *cfg
	mainCfgInited = true
	cfg.SetToDefault()
	err = os.Chdir(RootDir())
	if err == nil {
		err = cfg.ensureDirectories()
	}
	return mainCfg, err
}

// MustInitAppMain 初始化应用主配置文件，若失败会 panic
func MustInitAppMain(path string, parser func(string, any) error) FileConfig {
	cfg, err := InitAppMain(path, parser)
	if err != nil {
		panic(err)
	}
	return cfg
}
