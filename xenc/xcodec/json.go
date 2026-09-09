package xcodec

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"unsafe"

	"github.com/xanygo/anygo/xenc"
)

var JSON = xenc.NewCodec("JSON", jsonv1.Marshal, jsonv1.Unmarshal, "application/json")

var JSONV2 = xenc.NewCodec("JSONV2", jsonMarshal, jsonUnmarshal, "application/json")

func JSONString(obj any) string {
	bf, err := json.Marshal(obj)
	if err != nil {
		return err.Error()
	}
	return unsafe.String(unsafe.SliceData(bf), len(bf))
}

func jsonMarshal(a any) ([]byte, error) {
	return json.Marshal(a)
}

func jsonUnmarshal(b []byte, a any) error {
	return json.Unmarshal(b, a)
}
