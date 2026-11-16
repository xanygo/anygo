package xi18n

import "fmt"

var _ error = (*Error)(nil)

type Error struct {
	Messages []*LangMessage
	Params   []any
}

// Error implements [error].
func (e *Error) Error() string {
	if len(e.Messages) == 0 {
		return fmt.Sprintf("invalid xi18n.Error without Messages, Params=%v", e.Params)
	}
	return e.Messages[0].render(e.Params...)
}

// WithParams 设置消息格式化参数，返回新的 Error，不修改原对象。
func (e *Error) WithParams(params ...any) *Error {
	return &Error{
		Messages: e.Messages,
		Params:   params,
	}
}

// WithLangParams 设置首选语言和消息格式化参数，返回新的 Error。
//
// 按照语言优先级选择可用的消息，并替换原有参数。
func (e *Error) WithLangParams(langs []Language, params []any) *Error {
	return &Error{
		Messages: e.langs(langs),
		Params:   params,
	}
}

// WithLang 设置首选语言，返回新的 Error。
//
// 按照语言优先级选择可用的消息，并保留原有参数。
func (e *Error) WithLang(langs ...Language) *Error {
	return &Error{
		Messages: e.langs(langs),
		Params:   e.Params,
	}
}

func (e *Error) langs(langs []Language) []*LangMessage {
	msgs := make([]*LangMessage, 0, len(e.Messages))
	ls := make(map[Language]*LangMessage, len(e.Messages))
	for _, l := range e.Messages {
		ls[l.Lang] = l
	}
	for _, lang := range langs {
		if lm, ok := ls[lang]; ok {
			msgs = append(msgs, lm)
			delete(ls, lang)
		} else {
			for k, kv := range ls {
				if lang.SameFamily(k) {
					msgs = append(msgs, kv)
					delete(ls, lang)
				}
			}
		}
	}
	if len(msgs) == 0 {
		msgs = e.Messages
	}
	return msgs
}

type LangMessage struct {
	Lang    Language
	Text    string
	Message *Message
}

func (lm *LangMessage) render(args ...any) string {
	const prefix = "render xi18n.Error failed: "
	if lm.Message != nil {
		txt, err := lm.Message.Render(args...)
		if err == nil {
			return txt
		}
		return fmt.Sprintf(prefix+"%s", err.Error())
	}
	if lm.Text == "" {
		return prefix + "empty Text"
	}
	txt, err := renderMsgSlice(lm.Text, args...)
	if err == nil {
		return txt
	}
	return fmt.Sprintf(prefix+"%s", err.Error())
}
