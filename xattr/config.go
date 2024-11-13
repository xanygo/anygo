//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-13

package xattr

import (
	"path/filepath"
)

// FileConfig 应用的主配置文件，一般是 conf/app.yml 或 conf/app.json
type FileConfig struct {
	// Listen 监听的端口信息，可选
	Listen []string `yaml:"Listen"`

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

func (c *FileConfig) AutoCheck() error {
	if c.SelfPath != "" {
		if err := c.setDir(); err != nil {
			return err
		}
	}

	return nil
}

func (c *FileConfig) setDir() error {
	selfDir := filepath.Dir(c.SelfPath)
	rootDir := filepath.Dir(c.RootDir)
	if c.RootDir == "" {
		c.RootDir = rootDir
	}
	if c.ConfDir == "" {
		c.ConfDir = selfDir
	}
	if c.AppName == "" {
		c.AppName = filepath.Base(rootDir)
	}
	return nil
}

func (c *FileConfig) SetTo(attr *Attribute) {
	if c.AppName != "" {
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
	if c.RootDir != "" {
		attr.SetRootDir(c.RootDir)
	}
	if c.DataDir != "" {
		attr.SetDataDir(c.DataDir)
	}
	if c.ConfDir != "" {
		attr.SetConfDir(c.ConfDir)
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

func (c *FileConfig) SetToDefault() {
	c.SetTo(Default)
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
	if err := cfg.AutoCheck(); err != nil {
		return nil, err
	}
	cfg.SelfPath = path
	return cfg, nil
}

var appMainCfg FileConfig

// AppMainConfig 应用主配置文件，在使用前，需要先使用 InitAppMainCfg 或者 MustInitAppMainCfg 加载
func AppMainConfig() FileConfig {
	return appMainCfg
}

func InitAppMainCfg(path string, parser func(string, any) error) (FileConfig, error) {
	cfg, err := ParserFileConfig(path, parser)
	if err != nil {
		return FileConfig{}, err
	}
	appMainCfg = *cfg
	cfg.SetToDefault()
	return appMainCfg, nil
}

func MustInitAppMainCfg(path string, parser func(string, any) error) FileConfig {
	cfg, err := InitAppMainCfg(path, parser)
	if err != nil {
		panic(err)
	}
	return cfg
}
