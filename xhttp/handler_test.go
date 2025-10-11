//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-23

package xhttp

import (
	"net/http"
	"slices"
	"testing"

	"github.com/fsgo/fst"
)

func TestRegisterGroup(t *testing.T) {
	router := NewRouter()
	RegisterGroup(router, "/user", &testUserHandler{})
	fst.Len(t, router.subRoute, 11)

	wantKeys := []string{
		"GET|/user",
		"GET|/user/Index",

		"DELETE|/user/",
		"DELETE|/user/ByID",

		"GET|/user/Add",
		"GET|/user/ByID",
		"GET|/user/Edit",

		"GET|/user/Search",
		"POST|/user/",
		"POST|/user/Save",
		"PUT|/user/UpdateStatus",
	}
	slices.Sort(wantKeys)
	var gotKeys []string
	for _, sr := range router.subRoute {
		gotKeys = append(gotKeys, sr.UniqKey())
	}
	slices.Sort(gotKeys)
	fst.Equal(t, wantKeys, gotKeys)
}

var _ GroupHandler = (*testUserHandler)(nil)

type testUserHandler struct{}

func (h testUserHandler) GroupHandler() map[string]PatternHandler {
	return nil
}

func (h testUserHandler) Index(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Delete(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Post(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Search(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Add(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Edit(w http.ResponseWriter, r *http.Request) {}

func (h testUserHandler) Save(w http.ResponseWriter, r *http.Request) {}
