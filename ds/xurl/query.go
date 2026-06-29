package xurl

import (
	"net/url"
)

func BuildQueryString(p map[string]string) string {
	if len(p) == 0 {
		return ""
	}
	vs := url.Values{}
	for k, v := range p {
		vs.Add(k, v)
	}
	return vs.Encode()
}
