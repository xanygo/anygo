//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-11

package sse

import (
	"bytes"
	"io"
	"strconv"
)

var _ io.WriterTo = Event{}

// Event  server-sent-event
// https://html.spec.whatwg.org/multipage/server-sent-events.html
type Event struct {
	ID      string `json:"id,omitempty"`
	Event   string `json:"event,omitempty"`
	Data    string `json:"data,omitempty"`
	Retry   int    `json:"retry,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func (e Event) IsEmpty() bool {
	return e.ID == "" && e.Event == "" && e.Data == "" && e.Comment == ""
}

func (e Event) WriteTo(w io.Writer) (n int64, err error) {
	write := func(s string) error {
		m, err1 := io.WriteString(w, s)
		n += int64(m)
		return err1
	}

	if e.ID != "" {
		if err = write("id: " + e.ID + "\n"); err != nil {
			return
		}
	}

	if e.Event != "" {
		if err = write("event: " + e.Event + "\n"); err != nil {
			return
		}
	}

	if e.Data != "" {
		var start = 0
		for i := 0; i < len(e.Data); i++ {
			if e.Data[i] == '\n' {
				if err = write("data: " + e.Data[start:i] + "\n"); err != nil {
					return
				}
				start = i + 1
			}
		}
		if start <= len(e.Data) {
			if err = write("data: " + e.Data[start:] + "\n"); err != nil {
				return
			}
		}
	}
	if e.Retry > 0 {
		if err = write("retry: " + strconv.Itoa(e.Retry) + "\n"); err != nil {
			return
		}
	}

	if e.Comment != "" {
		var start = 0
		for i := 0; i < len(e.Comment); i++ {
			if e.Comment[i] == '\n' {
				if err = write(": " + e.Comment[start:i] + "\n"); err != nil {
					return
				}
				start = i + 1
			}
		}
		if start <= len(e.Comment) {
			if err = write(": " + e.Comment[start:] + "\n"); err != nil {
				return
			}
		}
	}

	if n == 0 {
		return 0, nil
	}

	err = write("\n")
	return n, err
}

func (e Event) Bytes() []byte {
	buf := new(bytes.Buffer)
	_, _ = e.WriteTo(buf)
	return buf.Bytes()
}
