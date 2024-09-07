//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-06

package xi18n

import (
	"fmt"
)

type Localize struct {
	messages map[string]Messages
}

func (l *Localize) Add(ns string, messages ...*Message) error {
	if l.messages == nil {
		l.messages = make(map[string]Messages)
	}
	msgs, ok := l.messages[ns]
	if !ok {
		msgs = Messages{}
	}
	for index, msg := range messages {
		if err := msg.initAndCheck(); err != nil {
			return fmt.Errorf("ns=%q, index=%d, %w", ns, index, err)
		}
		msgs[msg.Key] = msg
	}
	l.messages[ns] = msgs
	return nil
}

func (l *Localize) Find(ns string, key string) *Message {
	ms, ok := l.messages[ns]
	if !ok {
		return nil
	}
	return ms[key]
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

func Render(b *Bundle, languages []Language, ns string) func(key string, args ...any) string {
	return func(key string, args ...any) string {
		msg := FindMessage(b, languages, ns, key)
		if msg == nil {
			return "cannot find i18n message " + ns + "#" + key
		}
		result, err := msg.Render1(args)
		if err == nil {
			return result
		}
		return fmt.Sprintf("render i18n message %q#%q: %s", ns, key, err.Error())
	}
}
