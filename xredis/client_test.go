//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package xredis

import (
	"fmt"
	"testing"

	"github.com/xanygo/anygo/xt"
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
			xt.NoError(t, err2)
			xt.NotEmpty(t, s2)
			xt.NotEmpty(t, c2)
		})
	}

	fail := []string{
		"",
		"redis://user:password@host:8080/NotNum",
		"https://user:password@host:8080/NotNum",
	}
	for i, uri := range fail {
		t.Run(fmt.Sprintf("err_%d", i), func(t *testing.T) {
			s2, c2, err2 := NewClientByURI("demo", uri)
			xt.Error(t, err2)
			xt.Empty(t, s2)
			xt.Empty(t, c2)
		})
	}
}
