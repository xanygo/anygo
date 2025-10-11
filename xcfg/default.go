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

// Default 默认的 Configure 实例
func Default() *Configure {
	return defaultCfg.Load()
}

func SetDefault(cfg *Configure) (old *Configure) {
	return defaultCfg.Swap(cfg)
}

// Parse 解析配置，配置文件默认认为在 conf/ 目录下,
// 如 有 conf/abc.toml ，则 path = "abc.toml"
func Parse(path string, obj any) error {
	return Default().Parse(path, obj)
}

// MustParse 调用 Parse，若返回 err != nil 则 panic
func MustParse(path string, obj any) {
	if err := Parse(path, obj); err != nil {
		panic(err)
	}
}

// ParseBytes 解析 bytes 内容的配置
// ext: 文件后缀，如.json、.toml
func ParseBytes(ext string, content []byte, obj any) error {
	return Default().ParseBytes(ext, content, obj)
}

// MustParseBytes 调用 ParseBytes，若返回 err!=ni 则 panic
func MustParseBytes(ext string, content []byte, obj any) {
	if err := ParseBytes(ext, content, obj); err != nil {
		panic(err)
	}
}

// Exists 判断配置文件是否存在
//
// path: 配置文件路径，可以是绝对路径，也可以是相对于 ConfDir 的相对路径。
// 文件后缀是可选的，当查找文件不存在时，会添加上支持的后缀依次去判断。
// 如 Exists("app.toml") 会补充为 完整路径 {ConfDir}/app.json、{ConfDir}/app.xml 等去判断
func Exists(path string) bool {
	return Default().Exists(path)
}

// WithDecoder 注册一个解析器
//
// ext: 文件后缀，如 .json
func WithDecoder(ext string, fn xcodec.Decoder) error {
	return Default().WithDecoder(ext, fn)
}

// MustWithDecoder 注册一个解析器，若返回的 err!=nil 则 panic
func MustWithDecoder(ext string, fn xcodec.Decoder) {
	Default().MustWithDecoder(ext, fn)
}

// WithHook 注册辅助类 / Hook
func WithHook(hooks ...Hook) error {
	return Default().WithHook(hooks...)
}

// MustWithHook 注册一个辅助类，若失败会 panic
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
