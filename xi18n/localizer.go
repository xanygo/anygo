//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import (
	"fmt"
)

type Localize struct {
	messages map[string]*Message
}

func (l *Localize) Add(ns string, messages ...*Message) error {
	if l.messages == nil {
		l.messages = make(map[string]*Message)
	}
	for index, msg := range messages {
		if err := msg.initAndCheck(); err != nil {
			return fmt.Errorf("ns=%q, index=%d, %w", ns, index, err)
		}
		path := l.keyJoin(ns, msg.Key)
		l.messages[path] = msg
	}
	return nil
}

const nameSpaceKeySep = "@"

func (l *Localize) keyJoin(ns string, key string) string {
	return ns + nameSpaceKeySep + key
}

func (l *Localize) Find(ns string, key string) *Message {
	if len(l.messages) == 0 {
		return nil
	}
	path := l.keyJoin(ns, key)
	return l.messages[path]
}

func FindMessage(b *Bundle, languages []Language, ns string, key string) *Message {
	for _, lang := range languages {
		if msg := findMessage(b, lang, ns, key); msg != nil {
			return msg
		}
	}
	return nil
}

func findMessage(b *Bundle, lang Language, ns string, key string) *Message {
	l := b.Localize(lang)
	if l == nil {
		return nil
	}
	return l.Find(ns, key)
}
