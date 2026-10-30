package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xanygo/anygo/xhttp/trustheader"
	"github.com/xanygo/anygo/xt"
)

func TestIsAjax(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	xt.False(t, IsAjax(req))

	req.Header.Set("Hx-Request", "true")
	xt.True(t, IsAjax(req))
	req.Header.Del("Hx-Request")

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	xt.True(t, IsAjax(req))
	req.Header.Del("X-Requested-With")
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	req.Header.Set("Cf-Connecting-Ip", "127.0.0.3")

	xt.Equal(t, ClientIP(req), "192.0.2.1")

	trustheader.Add("Cf-Connecting-Ip")
	xt.Equal(t, ClientIP(req), "127.0.0.3")

	trustheader.Clear()
	xt.Equal(t, ClientIP(req), "192.0.2.1")
}

func TestRequestScheme(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	xt.Equal(t, RequestScheme(req1), "http")

	req2 := httptest.NewRequest(http.MethodGet, "https://127.0.0.1/", nil)
	xt.Equal(t, RequestScheme(req2), "https")
}

func TestRequestPort(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	xt.Equal(t, RequestPort(req1), 80)

	req2 := httptest.NewRequest(http.MethodGet, "https://127.0.0.1/", nil)
	xt.Equal(t, RequestPort(req2), 443)

	req3 := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	xt.Equal(t, RequestPort(req3), 8080)

	req4 := httptest.NewRequest(http.MethodGet, "https://127.0.0.1:8443/", nil)
	xt.Equal(t, RequestPort(req4), 8443)

	req5 := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	req5.Header.Set("X-Forwarded-Port", "443")
	trustheader.Add("X-Forwarded-Port")
	xt.Equal(t, RequestPort(req5), 443)
	trustheader.Clear()
}
