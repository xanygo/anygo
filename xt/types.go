package xt

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type StringByte interface {
	~string | ~[]byte
}

type Labeled struct {
	Label    string
	Content  string
	Children Labeleds
}

func (l Labeled) writeIndent(w io.Writer, indent int, labelWitdh int) (int, error) {
	space1 := strings.Repeat(" ", indent)
	format := space1 + "%" + strconv.Itoa(labelWitdh) + "s : "
	n, err := fmt.Fprintf(w, format, l.Label)
	if err != nil {
		return n, err
	}
	lines := strings.Split(l.Content, "\n")
	space2 := strings.Repeat(" ", labelWitdh+3)
	for i, line := range lines {
		if i > 0 {
			line = space1 + space2 + line
		}
		m, err := fmt.Fprintf(w, "%s\n", line)
		n += m
		if err != nil {
			return n, err
		}
	}
	if len(l.Children) > 0 {
		var m int
		m, err = l.Children.WriteIndent(w, indent+labelWitdh)
		n += m
	}
	return n, err
}

type Labeleds []Labeled

func (lb Labeleds) String() string {
	bf := &bytes.Buffer{}
	lb.WriteIndent(bf, 0)
	return bf.String()
}

func (lb Labeleds) WriteIndent(w io.Writer, indent int) (n int, err error) {
	if len(lb) == 0 {
		return 0, nil
	}
	var ml int
	for _, item := range lb {
		ml = max(ml, len(item.Label))
	}
	for _, item := range lb {
		m, err := item.writeIndent(w, indent, ml+2)
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, err
}
