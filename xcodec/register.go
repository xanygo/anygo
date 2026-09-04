package xcodec

import (
	"fmt"
	"strings"

	"github.com/xanygo/anygo/xerror"
)

var register = map[string]Codec{}

func Register(name string, codec Codec) {
	name = strings.ToLower(name)
	register[name] = codec
}

func Find(name string) (Codec, error) {
	value, ok := register[strings.ToLower(name)]
	if ok {
		return value, nil
	}
	return nil, fmt.Errorf("name=%q %w", name, xerror.NotFound)
}
