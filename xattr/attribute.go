//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"path/filepath"
	"sync"
	"sync/atomic"
)

// NewAttribute 创建一个新的 Attribute 对象
//
//	appName: 应用名，建议满足正则 [a-zA-Z0-9_-]+
//	rootDir: 应用的根目录，建议传入绝对路径
//
// ConfDir、DataDir、LogDir、TempDir 的默认值均依据 rootDir 推断而来，
// 并且会优先使用环境变量配置的额值，具体如下：
//
//	ConfDir : 优先使用环境变量 ANYGO_CONF 的值，若没有则使用默认值 {rootDir}/conf/
//	DataDir : 优先使用环境变量 ANYGO_DATA 的值，若没有则使用默认值 {rootDir}/data/
//	LogDir  : 优先使用环境变量 ANYGO_LOG 的值，若没有则使用默认值 {rootDir}/log/
//	TempDir : 优先使用环境变量 ANYGO_TEMP 的值，若没有则使用默认值 {rootDir}/temp/
func NewAttribute(appName string, rootDir string) *Attribute {
	attr := &Attribute{}
	attr.SetAppName(appName)
	attr.SetRootDir(osEnvDefault(eKeyRoot, rootDir))
	attr.SetConfDir(osEnvDefault(eKeyConf, "conf"))
	attr.SetDataDir(osEnvDefault(eKeyData, "data"))
	attr.SetTempDir(osEnvDefault(eKeyTemp, "temp"))
	attr.SetLogDir(osEnvDefault(eKeyLog, "log"))
	attr.SetIDC(osEnvDefault(eKeyIDC, IDCOnline))
	attr.SetRunMode(modeFromEnv(ModeProduct))
	return attr
}

const (
	IDCOnline = "online"
	IDCDev    = "dev"
)

type Attribute struct {
	rootDir string
	appName string
	dataDir string
	tempDir string
	logDir  string
	confDir string
	idc     string
	mode    atomic.Int32
	other   sync.Map
}

// SetAppName 设置应用名称，建议满足正则 [a-zA-Z0-9_-]+
func (a *Attribute) SetAppName(name string) {
	a.appName = name
}

// AppName 获取应用名
func (a *Attribute) AppName() string {
	return a.appName
}

// SetRootDir 设置应用根目录
func (a *Attribute) SetRootDir(dir string) {
	a.rootDir = dir
}

// RootDir 获取应用根目录
func (a *Attribute) RootDir() string {
	return a.rootDir
}

// SetDataDir 设置数据目录
// 应在 SetRootDir 调用之后调用
func (a *Attribute) SetDataDir(name string) {
	path, abs := parserDirName(name)
	if abs {
		a.dataDir = path
	} else {
		a.dataDir = filepath.Join(a.RootDir(), path)
	}
}

// DataDir 获取数据目录
func (a *Attribute) DataDir() string {
	return a.dataDir
}

func (a *Attribute) SetTempDir(name string) {
	path, abs := parserDirName(name)
	if abs {
		a.tempDir = path
	} else {
		a.tempDir = filepath.Join(a.RootDir(), path)
	}
}

func (a *Attribute) TempDir() string {
	return a.tempDir
}

func (a *Attribute) SetLogDir(name string) {
	path, abs := parserDirName(name)
	if abs {
		a.logDir = path
	} else {
		a.logDir = filepath.Join(a.RootDir(), path)
	}
}

func (a *Attribute) LogDir() string {
	return a.logDir
}

func (a *Attribute) SetConfDir(name string) {
	path, abs := parserDirName(name)
	if abs {
		a.confDir = path
	} else {
		a.confDir = filepath.Join(a.RootDir(), path)
	}
}

func (a *Attribute) ConfDir() string {
	return a.confDir
}

func (a *Attribute) SetIDC(idc string) {
	a.idc = idc
}

func (a *Attribute) IDC() string {
	return a.idc
}

// SetRunMode 设置运行模式，是并发安全的
func (a *Attribute) SetRunMode(mode Mode) {
	a.mode.Store(int32(mode))
}

// RunMode 读取运行模式
func (a *Attribute) RunMode() Mode {
	return Mode(a.mode.Load())
}

func (a *Attribute) Set(key any, value any) {
	a.other.Store(key, value)
}

func (a *Attribute) Get(key any) (any, bool) {
	return a.other.Load(key)
}

func (a *Attribute) Range(fn func(key any, value any) bool) {
	a.other.Range(fn)
}
