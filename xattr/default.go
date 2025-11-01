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

func IsDebugMode() bool {
	return RunMode() == ModeDebug
}

// SetRunMode (全局)设置应用的运行模式，是并发安全的
func SetRunMode(mode Mode) {
	Default.SetRunMode(mode)
}

func Set(key any, value any) {
	Default.Set(key, value)
}

func Get(key any) (any, bool) {
	return Default.Get(key)
}

func GetAs[T any](key any) (result T, ok bool) {
	val, ok := Get(key)
	if !ok {
		return result, false
	}
	result, ok = val.(T)
	return result, ok
}

func GetDefault[T any](key any, def T) T {
	val, ok := GetAs[T](key)
	if !ok {
		return def
	}
	return val
}

func Range(fn func(key any, value any) bool) {
	Default.Range(fn)
}
