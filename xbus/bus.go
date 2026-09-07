//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-05

package xbus

import (
	"context"
	"fmt"
	"sync/atomic"
)

// AnyTopic 表示任意的 Topic，可用于 Consumer 的注册和判断
var AnyTopic = newTopic(0, "any")

// Topic 用于定义消息类型
type Topic struct {
	id   int64
	name string
	str  string
}

func (t Topic) String() string {
	return t.str
}

func (t Topic) Name() string {
	return t.name
}

func (t Topic) Match(b Topic) bool {
	// ID 为 0 时表示任意，所以也可以匹配
	return t.id == b.id || t.id == 0 || b.id == 0
}

var globalTopicID atomic.Int64

func NewTopic(name string) Topic {
	return newTopic(globalTopicID.Add(1), name)
}

func newTopic(id int64, name string) Topic {
	return Topic{
		id:   id,
		name: name,
		str:  fmt.Sprintf("topic-%d-%s", id, name),
	}
}

type Message struct {
	Topic   Topic
	Key     any // 可选，消息的名称
	Payload any // 必填，消息体
}

// Producer 消息生产者
type Producer interface {
	Messages() <-chan Message
}

// Consumer 消息消费者
type Consumer interface {
	Consume(ctx context.Context, msg Message) error
}

type Named interface {
	Name() string
}

func Name(item any) string {
	if v, ok := item.(Named); ok {
		return v.Name()
	}
	return ""
}
