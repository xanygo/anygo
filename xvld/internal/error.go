package internal

import (
	"github.com/xanygo/anygo/xi18n"
)

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
