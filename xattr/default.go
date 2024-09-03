//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xattr

import (
	"os"
	"path/filepath"
)

// Default (全局)默认的环境信息
var Default *Attribute

func init() {
	doInit()
}

func doInit() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	Init(filepath.Base(pwd), pwd)
}

func Init(appName string, rootDir string) {
	Default = NewAttribute(appName, rootDir)
}

// MustInitWithAppConfPath 传入应用的主配置文件路径来初始化
//
// 假如传入 ./conf/app.toml, 其他路径信息依据此路径自动推断出
func MustInitWithAppConfPath(appConfPath string) {
	p, err := filepath.Abs(appConfPath)
	if err != nil {
		panic(err)
	}
	confDir := filepath.Dir(p)
	rootDir := filepath.Dir(confDir)
	appName := filepath.Base(rootDir)
	Init(appName, rootDir)
	SetConfDir(confDir)
}

func SetAppName(name string) {
	Default.SetAppName(name)
}

func AppName() string {
	return Default.AppName()
}

// RootDir (全局)获取应用根目录
func RootDir() string {
	return Default.RootDir()
}

// SetRootDir (全局)设置应用根目录
func SetRootDir(dir string) {
	Default.SetRootDir(dir)
}

// DataDir (全局)设置应用数据根目录
func DataDir() string {
	return Default.DataDir()
}

// SetDataDir 设置(全局)应用数据根目录
func SetDataDir(dir string) {
	Default.SetDataDir(dir)
}

// LogDir (全局)获取应用日志根目录
func LogDir() string {
	return Default.LogDir()
}

// SetLogDir (全局)设置应用日志根目录
func SetLogDir(dir string) {
	Default.SetLogDir(dir)
}

// ConfDir (全局)获取应用配置根目录
func ConfDir() string {
	return Default.ConfDir()
}

// SetConfDir (全局)设置应用配置根目录
func SetConfDir(dir string) {
	Default.SetConfDir(dir)
}

// TempDir (全局)获取应用临时文件根目录
func TempDir() string {
	return Default.TempDir()
}

// SetTempDir (全局)设置应用临时文件根目录
func SetTempDir(dir string) {
	Default.SetTempDir(dir)
}

// SetIDC (全局) 设置idc
func SetIDC(idc string) {
	Default.SetIDC(idc)
}

// IDC (全局)获取应用的 IDC
func IDC() string {
	return Default.IDC()
}

// RunMode (全局)获取应用的运行模式
func RunMode() Mode {
	return Default.RunMode()
}

// SetRunMode (全局)设置应用的运行模式，是并发安全的
func SetRunMode(mode Mode) {
	Default.SetRunMode(mode)
}

func SetAttr(key any, value any) {
	Default.SetAttr(key, value)
}

func Attr(key any) (any, bool) {
	return Default.Attr(key)
}

func AttrRange(fn func(key any, value any) bool) {
	Default.AttrRange(fn)
}
