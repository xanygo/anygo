package internal

import (
	"github.com/xanygo/anygo/xi18n"
)

// ErrDefault 没有定义对应的 error 时候的默认错误
var ErrDefault = &xi18n.Error{
	Messages: []*xi18n.LangMessage{
		{
			Text: "validation {0} failed",
			Lang: xi18n.LangEn,
		},
		{
			Text: "校验 {0} 失败",
			Lang: xi18n.LangZh,
		},
	},
}

// ErrBase 基础错误信息模板，添加字段名信息，用于包裹具体的错误作
var ErrBase = &xi18n.Error{
	Messages: []*xi18n.LangMessage{
		{
			Text: "Field '{0}': %w",
			Lang: xi18n.LangEn,
		},
		{
			Text: "字段 {0}: %w",
			Lang: xi18n.LangZh,
		},
	},
}

// ErrNotStruct 输入类型不是 struct 或者 *strcut 的错误
var ErrNotStruct = &xi18n.Error{
	Messages: []*xi18n.LangMessage{
		{
			Text: "Invalid input type: expected struct, got {0}",
			Lang: xi18n.LangEn,
		},
		{
			Text: "输入类型错误：期望 struct，实际为 {0}",
			Lang: xi18n.LangZh,
		},
	},
}
