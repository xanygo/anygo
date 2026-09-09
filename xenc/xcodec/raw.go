package xcodec

import (
	"fmt"
	"unsafe"

	"github.com/xanygo/anygo/xenc"
)

var Raw = xenc.NewCodec("Raw", rawEncode, rawDecode, "application/octet-stream")

func rawEncode(obj any) ([]byte, error) {
	switch val := obj.(type) {
	case []byte:
		return val, nil
	case string:
		return unsafe.Slice(unsafe.StringData(val), len(val)), nil
	default:
		return nil, fmt.Errorf("not support type %T for rawEncode", obj)
	}
}

func rawDecode(data []byte, obj any) error {
	switch val := obj.(type) {
	case *[]byte:
		*val = data
		return nil
	case *string:
		*val = string(data)
		return nil
	default:
		return fmt.Errorf("not support type %T for rawDecode", obj)
	}
}
