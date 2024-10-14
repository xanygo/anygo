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
}
