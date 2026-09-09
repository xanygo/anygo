package xcodec

import (
	"fmt"
	"net/url"
	"unsafe"

	"github.com/xanygo/anygo/xenc"
)

var Form = &FormCodec{}

var _ Codec = (*FormCodec)(nil)
var _ xenc.HasContentType = (*FormCodec)(nil)

type FormCodec struct {
}

func (f FormCodec) Name() string {
	return "Form"
}

func (f FormCodec) ContentType() string {
	return "application/x-www-form-urlencoded"
}

func (f FormCodec) Marshal(a any) ([]byte, error) {
	switch vv := a.(type) {
	case url.Values:
		str := vv.Encode()
		return unsafe.Slice(unsafe.StringData(str), len(str)), nil
	case map[string]string:
		uv := make(url.Values, len(vv))
		for k, v := range vv {
			uv.Set(k, v)
		}
		str := uv.Encode()
		return unsafe.Slice(unsafe.StringData(str), len(str)), nil
	default:
		return nil, fmt.Errorf("not support type %T for FormEncode", a)
	}
}

func (f FormCodec) Unmarshal(bf []byte, a any) error {
	if len(bf) == 0 {
		return nil
	}
	values, err := url.ParseQuery(string(bf))
	if err != nil {
		return err
	}

	switch vv := a.(type) {
	case *url.Values:
		*vv = values
		return nil
	case *map[string]string:
		m := make(map[string]string, len(values))
		for k, v := range values {
			if len(v) > 0 {
				m[k] = v[0] // 只取第一个
			}
		}
		*vv = m
		return nil
	default:
		return fmt.Errorf("not support type %T for FormDecode", a)
	}
}
