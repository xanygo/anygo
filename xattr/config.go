//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xattr

import (
	"fmt"
	"path/filepath"

	"github.com/xanygo/anygo/xcodec"
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

	// Other 其他项，可选
	Other map[string]any `yaml:"Other"`

	// SelfPath 配置自己的路径
	SelfPath string
}

func (c FileConfig) GetListen(name string) string {
	if len(c.Listen) == 0 {
		panic("empty Listen in " + c.SelfPath)
	}
	v, ok := c.Listen[name]
	if ok {
		return v
	}
	panic(fmt.Sprintf("not found Listen[%q] in %s", name, c.SelfPath))
}

func (c FileConfig) getAppName() string {
	if c.AppName != "" {
		return c.AppName
	}
	if c.SelfPath != "" {
		return filepath.Dir(filepath.Dir(c.SelfPath))
	}
	root := c.getRootDir()
	if root != "" {
		return filepath.Dir(root)
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
	return filepath.Dir(filepath.Dir(c.SelfPath))
}

func (c FileConfig) getConfDir() string {
	if c.ConfDir != "" {
		return c.ConfDir
	}
	if c.SelfPath == "" {
		return ""
	}
	return filepath.Dir(c.SelfPath)
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
		attr.Set(key, value)
	}
}

func (c FileConfig) SetToDefault() {
	c.SetTo(Default)
}

var _ xcodec.DecodeExtra = FileConfig{}

func (c FileConfig) NeedDecodeExtra() string {
	return "Other"
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

// InitAppMain 初始化应用主配置文件
func InitAppMain(path string, parser func(string, any) error) (FileConfig, error) {
	cfg, err := ParserFileConfig(path, parser)
	if err != nil {
		return FileConfig{}, err
	}
	mainCfg = *cfg
	mainCfgInited = true
	cfg.SetToDefault()
	return mainCfg, nil
}

// MustInitAppMain 初始化应用主配置文件，若失败会 panic
func MustInitAppMain(path string, parser func(string, any) error) FileConfig {
	cfg, err := InitAppMain(path, parser)
	if err != nil {
		panic(err)
	}
	return cfg
}
