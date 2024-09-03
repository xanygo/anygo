//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package xcfg

import (
	"context"
	"sync/atomic"

	"github.com/xanygo/anygo/xcodec"
)

var defaultCfg atomic.Pointer[Configure]

func init() {
	defaultCfg.Store(NewDefault())
}

// Default 默认的实例
func Default() *Configure {
	return defaultCfg.Load()
}

func SetDefault(cfg *Configure) (old *Configure) {
	return defaultCfg.Swap(cfg)
}

// Parse 解析配置，配置文件默认认为在 conf/目录下,
// 如 有 conf/abc.toml ，则 confName="abc.toml"
func Parse(confName string, obj any) error {
	return Default().Parse(confName, obj)
}

// MustParse 调用 Parse，若返回 err!=ni 则 panic
func MustParse(confName string, obj any) {
	if err := Parse(confName, obj); err != nil {
		panic(err)
	}
}

// ParseByAbsPath 解析绝对路径的配置
func ParseByAbsPath(confAbsPath string, obj any) error {
	return Default().ParseByAbsPath(confAbsPath, obj)
}

// MustParseByAbsPath 调用 ParseByAbsPath，若返回 err!=ni 则 panic
func MustParseByAbsPath(confAbsPath string, obj any) {
	if err := ParseByAbsPath(confAbsPath, obj); err != nil {
		panic(err)
	}
}

// ParseBytes （全局）解析 bytes
// fileExt 是文件后缀，如.json、.toml
func ParseBytes(fileExt string, content []byte, obj any) error {
	return Default().ParseBytes(fileExt, content, obj)
}

// MustParseBytes 调用 ParseBytes，若返回 err!=ni 则 panic
func MustParseBytes(fileExt string, content []byte, obj any) {
	if err := ParseBytes(fileExt, content, obj); err != nil {
		panic(err)
	}
}

// Exists  （全局）判断是否存在
//
//	confName 的文件后缀是可选的，当查找文件不存在时，会添加上支持的后缀依次去判断。
//	如 Exists("app.toml") 会去 {ConfDir}/app.toml 判断
func Exists(confName string) bool {
	return Default().Exists(confName)
}

// RegisterParser （全局）注册一个解析器
// fileExt 是文件后缀，如 .json
func RegisterParser(fileExt string, fn xcodec.Decoder) error {
	return Default().WithParser(fileExt, fn)
}

// MustWithParser 调用 RegisterParser，若返回的 err!=nil 则 panic
func MustWithParser(fileExt string, fn xcodec.Decoder) {
	Default().MustWithParser(fileExt, fn)
}

// WithHook （全局）注册一个辅助类
func WithHook(hooks ...Hook) error {
	return Default().WithHook(hooks...)
}

// MustWithHook （全局）注册一个辅助类，若失败会 panic
func MustWithHook(hooks ...Hook) {
	Default().MusWithHook(hooks...)
}

// CloneWithContext （全局）返回新的对象,并设置新的 ctx
func CloneWithContext(ctx context.Context) *Configure {
	return Default().CloneWithContext(ctx)
}

// CloneWithHook （全局）返回新的对象,并注册 Hook
func CloneWithHook(hooks ...Hook) *Configure {
	return Default().CloneWithHook(hooks...)
}
