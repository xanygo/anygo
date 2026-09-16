package xcookiejar

import (
	"fmt"
	"net/http/cookiejar"
	"strings"

	"github.com/xanygo/anygo/xerror"
)

var pslistMap = map[string]cookiejar.PublicSuffixList{}

// RegisterPSList 注册 PublicSuffixList, 以供 Loader 使用
//
// 默认一般可以使用  golang.org/x/net/publicsuffix
func RegisterPSList(name string, ps cookiejar.PublicSuffixList) error {
	raw := name
	name = strings.ToLower(name)
	if _, has := pslistMap[name]; has {
		return fmt.Errorf("already exists pslist %q", raw)
	}
	pslistMap[name] = ps
	return nil
}

func FindPsList(name string) (cookiejar.PublicSuffixList, error) {
	raw := name
	name = strings.ToLower(name)
	v, has := pslistMap[name]
	if !has {
		return nil, fmt.Errorf("%w pslist=%q", xerror.NotFound, raw)
	}
	return v, nil
}
