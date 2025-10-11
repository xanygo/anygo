//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-27

package xurl

import (
	"net/url"
)

func NewLink(base *url.URL, param string) string {
	query := base.Query()
	p, _ := url.ParseQuery(param)
	for name, values := range p {
		value := values[0]
		if value == "" {
			query.Del(name)
		} else {
			query.Set(name, value)
		}
	}
	if len(query) == 0 {
		return base.Path
	}
	return base.Path + "?" + query.Encode()
}
