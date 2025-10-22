//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-13

package xt

import "net/http"

type HTTPResponseChecker struct {
	resp *http.Response
	t    Testing
}

func WithHTTPResponse(t Testing, resp *http.Response) *HTTPResponseChecker {
	return &HTTPResponseChecker{resp: resp, t: t}
}

func (hc *HTTPResponseChecker) StatusCodeEqual(code int) *HTTPResponseChecker {
	if h, ok := hc.t.(Helper); ok {
		h.Helper()
	}
	Equal(hc.t, code, hc.resp.StatusCode)
	return hc
}

func (hc *HTTPResponseChecker) StatusCodeNotEqual(code int) *HTTPResponseChecker {
	if h, ok := hc.t.(Helper); ok {
		h.Helper()
	}
	NotEqual(hc.t, code, hc.resp.StatusCode)
	return hc
}
