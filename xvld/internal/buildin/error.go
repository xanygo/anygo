package buildin

import "github.com/xanygo/anygo/xi18n"

var Errs = map[string]*xi18n.Error{
	"required": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "is a required field",
			},
			{
				Lang: xi18n.LangZh,
				Text: "为必填字段",
			},
		},
	},
	"min": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be {0} or greater, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "值应不小于 {0}, 当前为 {1}",
			},
		},
	},
	"max": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be {0} or less, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "值应不大于 {0}, 当前为 {1}",
			},
		},
	},
	"min-len": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "length must be {0} or greater, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "长度应不小于 {0}, 当前为 {1}",
			},
		},
	},
	"max-len": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "legth must be {0} or less, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "长度应不大于 {0}, 当前为 {1}",
			},
		},
	},
	"len": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "length must be {0}, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "长度应为 {0}, 当前为 {1}",
			},
		},
	},
	"eq": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "is not equal to {0}, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "不等于 {0}, 当前为 {1}",
			},
		},
	},
	"ne": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "should not be equal to {0}, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "不应等于 {0}, 当前为 {1}",
			},
		},
	},
	"gt": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be greater than {0}, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "应大于 {0}, 当前为 {1}",
			},
		},
	},
	"gte": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be {1} or greater, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "应大于或等于 {0}, 当前为 {1}",
			},
		},
	},
	"lt": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be less than {0}, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "应小于 {0}, 当前为 {1}",
			},
		},
	},
	"lte": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be {1} or less, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "应小于或等于 {0}, 当前为 {1}",
			},
		},
	},
	"contains": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must contains {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须包含 {0}",
			},
		},
	},
	"not-contains": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must not contains {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "不能包含 {0}",
			},
		},
	},
	"oneof": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be one of the allowed values, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须为允许的值之一, 当前为 {1}",
			},
		},
	},
	"email": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid email address, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的邮箱地址, 当前为 {1}",
			},
		},
	},
	"url": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid URL, got {1}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的 URL, 当前为 {1}",
			},
		},
	},
	"alpha": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "can only contain alphabetic characters, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "只能包含字母, 当前为 {0}",
			},
		},
	},
	"alpha-num": errAlphanum,
	"alphanum":  errAlphanum,
	"uuid": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid UUID, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是有效的 UUID, 当前为 {0}",
			},
		},
	},
	"numeric": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid numeric value, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的数值, 当前为 {0}",
			},
		},
	},
	"hex-color": errHexColor,
	"hexcolor":  errHexColor,
	"base36": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid base36 value, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的 Base36, 当前为 {0}",
			},
		},
	},
	"base58": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid base58 value, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的 Bas58, 当前为 {0}",
			},
		},
	},
	"base64-url": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid base64-url value, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的 Base64-URL, 当前为 {0}",
			},
		},
	},
	"base64-std": {
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid base64-std value, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的 Base64-STD, 当前为 {0}",
			},
		},
	},
}

var (
	errAlphanum = &xi18n.Error{
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "can only contain alphanumeric characters, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "只能包含字母和数字, 当前为 {0}",
			},
		},
	}
	errHexColor = &xi18n.Error{
		Messages: []*xi18n.LangMessage{
			{
				Lang: xi18n.LangEn,
				Text: "must be a valid hex color, got {0}",
			},
			{
				Lang: xi18n.LangZh,
				Text: "必须是一个有效的十六进制颜色, 当前为 {0}",
			},
		},
	}
)
