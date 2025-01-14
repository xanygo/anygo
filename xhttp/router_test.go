//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-13

package xhttp

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"testing/iotest"

	"github.com/fsgo/fst"
)

func TestRouter_ServeHTTP(t *testing.T) {
	router := NewRouter()
	var called atomic.Int64
	checkCalled := func(t *testing.T) {
		fst.Equal(t, 1, called.Load())
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
			fst.Equal(t, 1, called.Load())
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
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("GET index /")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("GET /index/list", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/list")
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("GET index.list /index/list")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("GET /index/404", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/404")
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("Not-Found /index/404")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusNotFound, res.StatusCode)
		checkCalled(t)
	})

	t.Run("get /", func(t *testing.T) {
		req, _ := http.NewRequest("gEt", ts.URL, nil)
		res, err := ts.Client().Do(req)
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("GET index /")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("POST /index", func(t *testing.T) {
		res, err := ts.Client().Post(ts.URL+"/index", "", nil)
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("POST index /index")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("PUT /user/1", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", ts.URL+"/user/1", nil)
		res, err := ts.Client().Do(req)
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("PUT user /user/1, id=1")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("PUT /user/1.html", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", ts.URL+"/user/1.html", nil)
		res, err := ts.Client().Do(req)
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("PUT user.html /user/1.html, id=1")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})

	t.Run("Delete /user/1", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/user/1", nil)
		res, err := ts.Client().Do(req)
		fst.NoError(t, err)
		fst.NoError(t, iotest.TestReader(res.Body, []byte("DELETE /user/1 NOT Found")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusNotFound, res.StatusCode)
		checkCalled(t)
	})

	t.Run("get /index/routeInfo", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/index/routeInfo")
		fst.NoError(t, err)
		const body = `{"Method":"GET","Pattern":"/index/routeInfo","Path":"/index/routeInfo","MetaID":"1","MetaOther":{"k1":"v1"}}`
		fst.NoError(t, iotest.TestReader(res.Body, []byte(body+"\n")))
		fst.NoError(t, res.Body.Close())
		fst.Equal(t, http.StatusOK, res.StatusCode)
		checkCalled(t)
	})
}

// 测试中间件的执行顺序
func TestRouter_Use(t *testing.T) {
	router := NewRouter()
	var num atomic.Int64
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fst.Equal(t, 1, num.Add(1)) // 执行顺序 1
			handler.ServeHTTP(w, r)
			fst.Equal(t, 1+3+5+7, num.Load()) // 执行顺序 7
		})
	})
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fst.Equal(t, 1+3, num.Add(3)) // 执行顺序 2
			handler.ServeHTTP(w, r)
			fst.Equal(t, 1+3+5+7, num.Load()) // 执行顺序 6
		})
	})
	router.GetFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fst.Equal(t, 1+3+5+7, num.Add(7)) // 执行顺序 4
		_, _ = w.Write([]byte("ok"))
	}, func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fst.Equal(t, 1+3+5, num.Add(5)) // 执行顺序 3
			handler.ServeHTTP(w, r)
			fst.Equal(t, 1+3+5+7, num.Load()) // 执行顺序 5
		})
	})
	ts := httptest.NewServer(router)
	resp, err := ts.Client().Get(ts.URL)
	fst.NoError(t, err)
	fst.NoError(t, iotest.TestReader(resp.Body, []byte("ok")))
	defer resp.Body.Close()
}
