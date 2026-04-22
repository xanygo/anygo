//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import (
	"fmt"
	"maps"
	"path"
)

// Localize 一个具体的本地化配置，如 zh(中文) 的本地化资源对于一个 Localize 对象
type Localize struct {
	messages map[string]*Message
}

// Add 注册资源，存储的时候以 namespace + msg.Key 作为存储的主键，若重复新的 message 会覆盖旧的
//
// namespace: 名字空间，可选，可以为空字符串
// messages: 本地消息资源
func (l *Localize) Add(namespace string, messages ...*Message) error {
	if l.messages == nil {
		l.messages = make(map[string]*Message)
	}
	for index, msg := range messages {
		if err := msg.doInit(); err != nil {
			return fmt.Errorf("namespace=%q, index=%d, %w", namespace, index, err)
		}
		path := l.keyJoin(namespace, msg.Key)
		l.messages[path] = msg
	}
	return nil
}

// MustAdd 简化的 Add，若有异常会 panic
func (l *Localize) MustAdd(namespace string, messages ...*Message) {
	if err := l.Add(namespace, messages...); err != nil {
		panic(err)
	}
}

func (l *Localize) keyJoin(namespace string, key string) string {
	return path.Join(namespace, key)
}

// FindNS 从 namespace 里查找一条具体的消息，若查找不到会返回 nil
//
// namespace: 命名空间，可以为空
// key: 消息的 key，若包含 "/"，会认为前面部分是 namespace
func (l *Localize) FindNS(namespace string, key string) *Message {
	if len(l.messages) == 0 {
		return nil
	}
	path := l.keyJoin(namespace, key)
	return l.messages[path]
}

// Find 查找一条具体的消息，若查找不到会返回 nil。key 是完整的包含 namespace 的 path
func (l *Localize) Find(key string) *Message {
	if len(l.messages) == 0 {
		return nil
	}
	return l.messages[key]
}

func (l *Localize) Clone() *Localize {
	return &Localize{
		messages: maps.Clone(l.messages),
	}
}

// FindMessage 在 Bundle 中，使用推荐的 Language 列表，查找指定的消息，若查找不到会返回 nil
//
// languages: 待查找的语言列表，优先支持的排在前面，若 len=0 则，使用 Bundle 里所有的语言列表查询
func FindMessage(b *Bundle, languages []Language, namespace string, key string) *Message {
	if namespace != "" {
		key = path.Join(namespace, key)
	}
	if len(languages) == 0 {
		languages = b.Languages()
	}
	for _, lang := range languages {
		if msg := findMessage(b, lang, key); msg != nil {
			return msg
		}
	}
	return nil
}

func findMessage(b *Bundle, lang Language, key string) *Message {
	l := b.Localize(lang)
	if l == nil {
		return nil
	}
	return l.Find(key)
}
