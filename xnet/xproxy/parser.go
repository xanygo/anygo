//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-08

package xproxy

import (
	"fmt"
	"net/url"
)

func ParserProxyURL(pu string) (*Config, error) {
	u, err := url.Parse(pu)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	switch u.Scheme {
	case "http":
		cfg.Protocol = "HTTP"
	case "https":
		cfg.Protocol = "HTTPS"
	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %q", u.Scheme)
	}
	if u.User != nil {
		cfg.Username = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	cfg.Host = u.Hostname()
	cfg.Port = u.Port()
	return cfg, nil
}
