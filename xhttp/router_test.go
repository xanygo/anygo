//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-13

package xhttp

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"testing/iotest"

	"github.com/xanygo/anygo/xt"
)

func TestRouter_ServeHTTP(t *testing.T) {
	router := NewRouter()
	var called atomic.Int64
	checkCalled := func(t *testing.T) {
		xt.Equal(t, called.Load(), 1)
		called.Store(0)
	}
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			num := called.Add(1)
			t.Logf("middleware called, num=%d, uri=%s", num, r.RequestURI)
			handler.ServeHTTP(w, r)
		})
	})
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xt.Equal(t, called.Load(), 1)
			handler.ServeHTTP(w, r)
		})
	})
	router.NotFoundFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(r.Method + " " + r.RequestURI + " NOT Found"))
	})
	router.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + " index " + r.RequestURI))
	})
	router.PostFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("POST index " + r.RequestURI))
	})
	router.PutFunc("/user/{id}.html", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + " user.html " + r.RequestURI + ", id=" + r.PathValue("id")))
	})
	router.PutFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + " user " + r.RequestURI + ", id=" + r.PathValue("id")))
	})

	g1 := router.Prefix("/index/")
	g1.GetFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method + " index.list " + r.RequestURI))
	})
	g1.NotFoundFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteTextStatus(w, http.StatusNotFound, []byte("Not-Found "+r.RequestURI))
	})
	g1.GetFunc("/routeInfo meta|id=1,k1=v1", func(w http.ResponseWriter, r *http.Request) {
		info := ReadRouteInfo(r.Context())
		WriteJSON(w, info)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("GET /", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/")
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("GET index /")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("GET /index/list", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/list")
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("GET index.list /index/list")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("GET /index/404", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/404")
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("Not-Found /index/404")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusNotFound)
		checkCalled(t)
	})

	t.Run("get /", func(t *testing.T) {
		req, _ := http.NewRequest("gEt", ts.URL, nil)
		res, err := ts.Client().Do(req)
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("GET index /")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("POST /index", func(t *testing.T) {
		res, err := ts.Client().Post(ts.URL+"/index", "", nil)
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("POST index /index")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("PUT /user/1", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", ts.URL+"/user/1", nil)
		res, err := ts.Client().Do(req)
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("PUT user /user/1, id=1")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("PUT /user/1.html", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", ts.URL+"/user/1.html", nil)
		res, err := ts.Client().Do(req)
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("PUT user.html /user/1.html, id=1")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})

	t.Run("Delete /user/1", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/user/1", nil)
		res, err := ts.Client().Do(req)
		xt.NoError(t, err)
		xt.NoError(t, iotest.TestReader(res.Body, []byte("DELETE /user/1 NOT Found")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusNotFound)
		checkCalled(t)
	})

	t.Run("get /index/routeInfo", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/routeInfo")
		xt.NoError(t, err)
		const body = `{"Method":"GET","Pattern":"/index/routeInfo","Path":"/index/routeInfo","MetaID":"1","MetaOther":{"k1":"v1"}}`
		xt.NoError(t, iotest.TestReader(res.Body, []byte(body+"\n")))
		xt.NoError(t, res.Body.Close())
		xt.Equal(t, res.StatusCode, http.StatusOK)
		checkCalled(t)
	})
}

// 测试中间件的执行顺序
func TestRouter_Use(t *testing.T) {
	router := NewRouter()
	var num atomic.Int64
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xt.Equal(t, num.Add(1), 1) // 执行顺序 1
			handler.ServeHTTP(w, r)
			xt.Equal(t, num.Load(), 1+3+5+7) // 执行顺序 7
		})
	})
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xt.Equal(t, num.Add(3), 1+3) // 执行顺序 2
			handler.ServeHTTP(w, r)
			xt.Equal(t, num.Load(), 1+3+5+7) // 执行顺序 6
		})
	})
	router.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		xt.Equal(t, num.Add(7), 1+3+5+7) // 执行顺序 4
		_, _ = w.Write([]byte("ok"))
	}, func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xt.Equal(t, num.Add(5), 1+3+5) // 执行顺序 3
			handler.ServeHTTP(w, r)
			xt.Equal(t, num.Load(), 1+3+5+7) // 执行顺序 5
		})
	})
	ts := httptest.NewServer(router)
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL)
	xt.NoError(t, err)
	xt.NoError(t, iotest.TestReader(resp.Body, []byte("ok")))
	defer resp.Body.Close()
}

func TestRouter_Prefix(t *testing.T) {
	router := NewRouter()
	var num atomic.Int64
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("run md-1-1")
			xt.Equal(t, num.Add(1), 1) // 执行顺序 1

			ri := ReadRouteInfo(r.Context())
			xt.Equal(t, ri.Method, http.MethodGet)
			xt.Equal(t, r.Method, http.MethodGet)
			xt.Equal(t, ri.Path, "/api/index")

			session, _ := ri.GetMeta("session")
			xt.Equal(t, session, "no")

			handler.ServeHTTP(w, r)
			log.Println("run md-1-2")
			xt.Equal(t, num.Load(), 1+3+5) // 执行顺序 5
		})
	})
	p := router.Prefix("/api/", func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("run md-2-1") // 执行顺序 2
			xt.Equal(t, num.Load(), 1)
			handler.ServeHTTP(w, r)
			log.Println("run md-2-2") // 执行顺序 4
			xt.Equal(t, num.Add(5), 1+3+5)
		})
	})
	p.GetFunc("/index meta|session=no", func(w http.ResponseWriter, r *http.Request) {
		log.Println("run handler-3") // 执行顺序 3
		xt.Equal(t, num.Load(), 1)
		xt.Equal(t, num.Add(3), 1+3)
		_, _ = w.Write([]byte("ok"))
	})
	ts := httptest.NewServer(router)
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL + "/api/index")
	xt.NoError(t, err)
	xt.NoError(t, iotest.TestReader(resp.Body, []byte("ok")))
	defer resp.Body.Close()
}

func TestRouter_PrefixPrefix(t *testing.T) {
	router := NewRouter()
	var num1 atomic.Int64
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			num1.Add(1)
			handler.ServeHTTP(w, r)
		})
	})
	// 测试多层级 Prefix
	p1 := router.Prefix("/api/")
	p2 := p1.Prefix("/v1/")

	var num2 atomic.Int64
	p2.GetFunc("/hello", func(w http.ResponseWriter, request *http.Request) {
		num2.Add(1)
		w.WriteHeader(200)
		w.Write([]byte("world"))
	})

	p3 := router.Prefix("/v3-")
	p3.GetFunc("info", func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello-info"))
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("api-v1-hello", func(t *testing.T) {
		fullURL := ts.URL + "/api/v1/hello"
		t.Logf("ts.Client().Get %q", fullURL)
		resp, err := ts.Client().Get(fullURL)
		xt.NoError(t, err)
		xt.Equal(t, http.StatusOK, resp.StatusCode)
		xt.Equal(t, num1.Load(), 1)
		xt.Equal(t, num2.Load(), 1)

		xt.NoError(t, iotest.TestReader(resp.Body, []byte("world")))
		defer resp.Body.Close()
	})

	t.Run("v3-info", func(t *testing.T) {
		fullURL := ts.URL + "/v3-info"
		t.Logf("ts.Client().Get %q", fullURL)
		resp, err := ts.Client().Get(fullURL)
		xt.NoError(t, err)
		xt.Equal(t, http.StatusOK, resp.StatusCode)
		xt.Equal(t, num1.Load(), 2)

		xt.NoError(t, iotest.TestReader(resp.Body, []byte("hello-info")))
		defer resp.Body.Close()
	})
}
