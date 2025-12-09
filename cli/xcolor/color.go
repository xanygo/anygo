//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-25

package xcolor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xanygo/anygo/internal/zos"
)

var noColor atomic.Bool

func init() {
	no := os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" || !zos.IsTerminalFile(os.Stdout)
	noColor.Store(no)
}

var output atomic.Value

func SetOutput(out io.Writer) {
	output.Store(&out)
}

func Output() io.Writer {
	val := output.Load()
	if val == nil {
		return os.Stderr
	}
	return val.(io.Writer)
}

func SetColorable(enable bool) {
	noColor.Store(!enable)
}

func New(ids ...Code) *Color {
	c := &Color{
		ids: ids,
	}
	c.init()
	return c
}

type Color struct {
	ids   []Code
	begin []byte
	end   []byte
	endLn []byte
}

const escape = "\x1b"

func (c *Color) init() {
	tmp := make([]string, len(c.ids))
	for i, id := range c.ids {
		tmp[i] = fmt.Sprint(id)
	}
	c.begin = []byte(fmt.Sprintf("%s[%sm", escape, strings.Join(tmp, ";")))

	c.end = []byte(fmt.Sprintf("%s[%dm", escape, reset))
	c.endLn = []byte(string(c.end) + "\n")
}

func (c *Color) Add(code ...Code) {
	c.ids = append(c.ids, code...)
	c.init()
}

func (c *Color) Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	if noColor.Load() {
		if len(a) == 0 {
			return io.WriteString(w, format)
		}
		return fmt.Fprintf(w, format, a...)
	}

	bf := getBP()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if len(a) == 0 {
		bf.WriteString(format)
	} else if _, err = fmt.Fprintf(bf, format, a...); err != nil {
		return 0, err
	}
	bf.Write(c.end)

	return w.Write(bf.Bytes())
}

func (c *Color) String(txt string) string {
	if noColor.Load() {
		return txt
	}
	bf := getBP()
	defer bp.Put(bf)
	bf.Write(c.begin)
	bf.WriteString(txt)
	bf.Write(c.end)
	return bf.String()
}

var bp = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func getBP() *bytes.Buffer {
	bf := bp.Get().(*bytes.Buffer)
	bf.Reset()
	return bf
}

func (c *Color) Fprintln(w io.Writer, a ...any) (n int, err error) {
	if len(a) == 0 {
		return w.Write([]byte("\n"))
	}
	if noColor.Load() {
		return fmt.Fprintln(w, a...)
	}
	bf := getBP()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if _, err = fmt.Fprint(bf, a...); err != nil {
		return 0, err
	}
	bf.Write(c.endLn)
	return w.Write(bf.Bytes())
}

func (c *Color) Fprint(w io.Writer, a ...any) (n int, err error) {
	if len(a) == 0 {
		return 0, nil
	}
	if noColor.Load() {
		return fmt.Fprint(w, a...)
	}

	bf := getBP()
	defer bp.Put(bf)

	bf.Write(c.begin)
	if _, err = fmt.Fprint(bf, a...); err != nil {
		return 0, err
	}
	bf.Write(c.end)
	return w.Write(bf.Bytes())
}

func (c *Color) Printf(format string, a ...any) (n int, err error) {
	out := Output()
	return c.Fprintf(out, format, a...)
}

func (c *Color) Println(a ...any) (n int, err error) {
	out := Output()
	return c.Fprintln(out, a...)
}

func (c *Color) Sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return c.String(format)
	}
	bf := getBP()
	defer bp.Put(bf)
	c.Fprintf(bf, format, a...)
	return bf.String()
}

func (c *Color) Sprint(a ...any) string {
	bf := getBP()
	defer bp.Put(bf)
	c.Fprint(bf, a...)
	return bf.String()
}

func (c *Color) SprintFunc() func(a ...any) string {
	return func(a ...any) string {
		if len(a) == 0 {
			return ""
		}
		return c.String(fmt.Sprint(a...))
	}
}

func (c *Color) SprintfFunc(format string) func(a ...any) string {
	return func(a ...any) string {
		if len(a) == 0 {
			return c.String(format)
		}
		return c.String(fmt.Sprintf(format, a...))
	}
}

func (c *Color) FprintlnFunc(w io.Writer) func(a ...any) (n int, err error) {
	return func(a ...any) (n int, err error) {
		return c.Fprintln(w, a...)
	}
}

func (c *Color) PrintfLnFunc(format string) func(a ...any) (n int, err error) {
	return func(a ...any) (n int, err error) {
		if len(a) == 0 {
			return c.Println(format)
		}
		return c.Println(fmt.Sprintf(format, a...))
	}
}

var cache = &sync.Map{}

func getByID(id Code) *Color {
	v, ok := cache.Load(id)
	if ok {
		return v.(*Color)
	}
	c := New(id)
	cache.Store(id, c)
	return c
}

func printfln(c *Color, format string, a ...any) {
	c.PrintfLnFunc(format)(a...)
}
