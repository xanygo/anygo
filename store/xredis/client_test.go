//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package xredis

import (
	"fmt"
	"testing"

	"github.com/fsgo/fst"
)

func TestNewClientByURI(t *testing.T) {
	success := []string{
		"redis://user:password@host:8080/0",
		"redis://user:password@host:8080/1",
		"redis://host:8080/1",
		"rediss://user:password@host:8080/1",
	}
	for i, uri := range success {
		t.Run(fmt.Sprintf("succ_%d", i), func(t *testing.T) {
			s2, c2, err2 := NewClientByURI("demo", uri)
			fst.NoError(t, err2)
			fst.NotEmpty(t, s2)
			fst.NotEmpty(t, c2)
		})
	}

	fail := []string{
		"",
		"redis://user:password@host:8080/NotNum",
		"https://user:password@host:8080/NotNum",
	}
	for i, uri := range fail {
		t.Run(fmt.Sprintf("fail_%d", i), func(t *testing.T) {
			s2, c2, err2 := NewClientByURI("demo", uri)
			fst.Error(t, err2)
			fst.Empty(t, s2)
			fst.Empty(t, c2)
		})
	}
}
