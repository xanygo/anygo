package xenc

import (
	"errors"
)

type Codec interface {
	Namer
	Marshaler
	Unmarshaler
}

type Marshaler interface {
	Marshal(any) ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte, any) error
}

// UnmarshalExtra 当被解析的对象，实现了此接口的时候，并且 NeedDecodeExtra 返回了有效的字段名，
// 则会将未在 struct 中定义的字段，全部解析到指定的字段里。
type UnmarshalExtra interface {
	// NeedDecodeExtra 存储未定义字段的字段名，返回非空为有效。
	// 并且返回的名字必须在 struct 中存在，而且必须是 map[string]any 类型
	NeedDecodeExtra() string
}

// Namer 名字
type Namer interface {
	Name() string
}

func Name(obj any) string {
	if hn, ok := obj.(Namer); ok {
		return hn.Name()
	}
	return ""
}

type MarshalFunc func(any) ([]byte, error)

func (e MarshalFunc) Marshal(v any) ([]byte, error) {
	return e(v)
}

type UnmarshalFunc func([]byte, any) error

func (d UnmarshalFunc) Unmarshal(v []byte, r any) error {
	return d(v, r)
}

func NewCodec(name string, e MarshalFunc, d UnmarshalFunc, ct string) Codec {
	return &codec{name: name, e: e, d: d, ct: ct}
}

var _ Codec = (*codec)(nil)
var _ HasContentType = (*codec)(nil)

type codec struct {
	name string
	e    MarshalFunc
	d    UnmarshalFunc
	ct   string
}

func (c *codec) Marshal(a any) ([]byte, error) {
	return c.e(a)
}

func (c *codec) Unmarshal(bf []byte, a any) error {
	return c.d(bf, a)
}

func (c *codec) Name() string {
	return c.name
}

func (c *codec) ContentType() string {
	return c.ct
}

type HasContentType interface {
	ContentType() string
}

var errNoCt = errors.New("invalid codec: not xcodec.HasContentType")

func ContentType(c Marshaler) (string, error) {
	if hct, ok := c.(HasContentType); ok {
		return hct.ContentType(), nil
	}
	return "", errNoCt
}
