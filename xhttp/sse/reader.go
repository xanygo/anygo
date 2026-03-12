//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-11

package sse

import (
	"io"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/xio"
)

func ReadEvent(rd xio.StringReader) (Event, error) {
	var e Event
	var dataLines []string
	var comments []string

	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) == 0 {
				return e, io.EOF
			}
			return e, err
		}

		line = strings.TrimRight(line, "\r\n")

		// 空行：事件结束
		if line == "" {
			e.Data = strings.Join(dataLines, "\n")
			e.Comment = strings.Join(comments, "\n")
			return e, nil
		}

		// comment
		if after, found := strings.CutPrefix(line, ": "); found {
			comments = append(comments, after)
			continue
		}

		field := line
		var value string
		if before, after, found := strings.Cut(line, ": "); found {
			field = before
			value = after
		}

		switch field {
		case "id":
			e.ID = value
		case "event":
			e.Event = value
		case "data":
			dataLines = append(dataLines, value)
		case "retry":
			if v, err := strconv.Atoi(value); err == nil {
				e.Retry = v
			}
		}
	}
}
