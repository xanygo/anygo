//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-12

package sse

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestReadEvent(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		const txt = "data: hello\n\n"
		rd := bufio.NewReader(strings.NewReader(txt))
		e1, err1 := ReadEvent(rd)
		xt.NoError(t, err1)
		xt.Equal(t, e1, Event{Data: "hello"})

		e2, err2 := ReadEvent(rd)
		xt.Empty(t, e2)
		xt.ErrorIs(t, err2, io.EOF)
	})

	t.Run("case 2", func(t *testing.T) {
		const txt = "data: hello\n\ndata: world\n\n"
		rd := bufio.NewReader(strings.NewReader(txt))
		e1, err1 := ReadEvent(rd)
		xt.NoError(t, err1)
		xt.Equal(t, e1, Event{Data: "hello"})

		e2, err2 := ReadEvent(rd)
		xt.NoError(t, err2)
		xt.Equal(t, e2, Event{Data: "world"})

		e3, err3 := ReadEvent(rd)
		xt.Empty(t, e3)
		xt.ErrorIs(t, err3, io.EOF)
	})

	t.Run("case 3", func(t *testing.T) {
		const txt = ": hello\n: world\n\n"
		rd := bufio.NewReader(strings.NewReader(txt))
		e1, err1 := ReadEvent(rd)
		xt.NoError(t, err1)
		xt.Equal(t, e1, Event{Comment: "hello\nworld"})

		e2, err2 := ReadEvent(rd)
		xt.Empty(t, e2)
		xt.ErrorIs(t, err2, io.EOF)
	})

	t.Run("case 4", func(t *testing.T) {
		const txt = "data: hello\ndata: world\n\n"
		rd := bufio.NewReader(strings.NewReader(txt))
		e1, err1 := ReadEvent(rd)
		xt.NoError(t, err1)
		xt.Equal(t, e1, Event{Data: "hello\nworld"})

		e2, err2 := ReadEvent(rd)
		xt.Empty(t, e2)
		xt.ErrorIs(t, err2, io.EOF)
	})
}
