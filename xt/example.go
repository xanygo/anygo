//go:build ignore

package main

import (
	"fmt"
	"io"
	"net"

	"github.com/xanygo/anygo/xt"
)

var _ xt.Testing = (*myTest)(nil)

type myTest struct{}

// Fatalf implements [xt.Testing].
func (m *myTest) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg)
}

var t = &myTest{}

func main() {
	// 打印出各种错误信息
	xt.Equal(t, 1, 2)
	xt.NotEqual(t, 1, 1)

	xt.AnyOf(t, 3, xt.Slice(1, 2))
	xt.NotAnyOf(t, 3, xt.Slice(1, 2, 3, 4, 3))

	xt.Less(t, 3, 1)
	xt.LessOrEqual(t, 3, 1)

	xt.Greater(t, 1, 3)
	xt.GreaterOrEqual(t, 1, 3)

	xt.Error(t, nil)
	xt.NoError(t, io.EOF)

	xt.Nil(t, 123)
	xt.NotNil(t, nil)
	var a *net.Dialer
	xt.NotNil(t, a)

	xt.Empty(t, 123)

	xt.NotEmpty(t, 0)

	xt.HasPrefix(t, "hello ", "world ")
	xt.HasPrefix(t, []byte("hello "), []byte("world "))

	xt.NotPrefix(t, "hello world", "hello")
	xt.NotPrefix(t, []byte("hello world"), []byte("hello"))

	xt.Contains(t, "hello", "world")
	xt.NotContains(t, "hello", "h")

	xt.ErrorIs(t, nil, io.EOF)

	err1 := fmt.Errorf("some error %w", io.EOF)
	xt.ErrorNot(t, err1, io.EOF)
}
