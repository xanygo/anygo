// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/8

package xt

import (
	"fmt"
	"io"
	"testing"
)

func TestEqual(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Equal(t, 1, 1)
		Equal(t, "a", "a")
		Equal(t, []int{1}, []int{1})
		NotEqual(t, "a", "b")
		NotEqual(t, 1, 2)
	})

	mt.Fail(func(t Testing) {
		Equal(t, 1, 2)
		Equal(t, "a", "b")
		NotEqual(t, 1, 1)
	})
}

func TestError(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Error(t, io.EOF)
		NoError(t, nil)
		var err1 error
		NoError(t, err1)
	})
	mt.Fail(func(t Testing) {
		NoError(t, io.EOF)
		Error(t, nil)
	})
}

func TestTrue(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		True(t, true)
		False(t, false)
	})
	mt.Fail(func(t Testing) {
		True(t, false)
		False(t, true)
	})
}

func TestNil(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Nil(t, nil)
		NotNil(t, 1)
	})
	mt.Fail(func(t Testing) {
		Nil(t, 1)
		NotNil(t, nil)
	})
}

func TestEmpty(t *testing.T) {
	type TStruct struct {
		x int
	}
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Empty(t, nil)
		Empty(t, 0)
		Empty(t, false)
		Empty(t, make(chan int))
		Empty(t, "")
		Empty(t, [1]int{})
		Empty(t, TStruct{})

		NotEmpty(t, 1)
		NotEmpty(t, true)
		NotEmpty(t, [1]int{1})
		NotEmpty(t, "ok")
		NotEmpty(t, TStruct{x: 1})

		var v1 *TStruct
		Empty(t, v1)

		Empty(t, &TStruct{})
	})
	mt.Fail(func(t Testing) {
		Empty(t, 1)
		Empty(t, true)

		NotEmpty(t, false)
		NotEmpty(t, 0)
	})
}

func TestContains(t *testing.T) {
	mt := newMyTesting(t)
	type str1 string
	type str2 []byte

	mt.Success(func(t Testing) {
		Contains(t, "hello", "h")
		Contains[str1](t, "hello", "h")
		NotContains(t, "hello", "a")
		Contains(t, []byte("hello"), []byte("h"))
		Contains(t, str2("hello"), str2("h"))
		NotContains(t, []byte("hello"), []byte("a"))
	})
	mt.Fail(func(t Testing) {
		Contains(t, "hello", "a")
		NotContains(t, "hello", "h")
		Contains(t, []byte("hello"), []byte("a"))
		NotContains(t, []byte("hello"), []byte("h"))
	})
}

func TestSliceContains(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		SliceContains(t, []int{1, 2}, 1)
		SliceNotContains(t, []int{1, 2}, 3)
	})
	mt.Fail(func(t Testing) {
		SliceContains(t, []int{1, 2}, 3)
		SliceNotContains(t, []int{1, 2}, 1)
	})
}

func TestSamePtr(t *testing.T) {
	type TStruct struct {
		_ int
	}
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		v1 := &TStruct{}
		v2 := v1
		SamePtr(t, v1, v2)
		v3 := &TStruct{}
		NotSamePtr(t, v1, v3)
	})
	mt.Fail(func(t Testing) {
		SamePtr(t, &TStruct{}, &TStruct{})
		v1 := &TStruct{}
		v2 := v1
		NotSamePtr(t, v1, v2)
	})
}

func TestLess(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Less(t, 1, 2)
		Less(t, "a", "b")
		Less(t, 0.1, 0.2)
		Less(t, uint32(1), uint32(2))
	})
	mt.Fail(func(t Testing) {
		Less(t, 3, 2)
		Less(t, "c", "b")
		Less(t, 0.3, 0.2)
		Less(t, uint32(3), uint32(2))
	})
}

func TestLessOrEqual(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		LessOrEqual(t, 1, 2)
		LessOrEqual(t, 2, 2)

		LessOrEqual(t, "a", "b")
		LessOrEqual(t, "b", "b")

		LessOrEqual(t, 0.1, 0.2)
		LessOrEqual(t, 0.2, 0.2)

		LessOrEqual(t, uint32(1), uint32(2))
		LessOrEqual(t, uint32(2), uint32(2))
	})
	mt.Fail(func(t Testing) {
		LessOrEqual(t, 3, 2)
		LessOrEqual(t, "c", "b")
		LessOrEqual(t, 0.3, 0.2)
		LessOrEqual(t, uint32(3), uint32(2))
	})
}

func TestGreater(t *testing.T) {
	type intA int
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Greater(t, 3, 2)
		Greater(t, intA(3), intA(2))
		Greater(t, "c", "b")
		Greater(t, 0.3, 0.2)
		Greater(t, uint32(3), uint32(2))
	})
	mt.Fail(func(t Testing) {
		Greater(t, 1, 2)
		Greater(t, 2, 2)

		Greater(t, "a", "b")
		Greater(t, "b", "b")

		Greater(t, 0.2, 0.2)
		Greater(t, 0.1, 0.2)

		Greater(t, uint32(2), uint32(2))
		Greater(t, uint32(1), uint32(2))
	})
}

func TestGreaterOrEqual(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		GreaterOrEqual(t, 3, 2)
		GreaterOrEqual(t, 3, 3)

		GreaterOrEqual(t, "c", "b")
		GreaterOrEqual(t, "c", "c")

		GreaterOrEqual(t, 0.3, 0.2)
		GreaterOrEqual(t, 0.3, 0.3)

		GreaterOrEqual(t, uint32(3), uint32(2))
		GreaterOrEqual(t, uint32(3), uint32(3))
	})
	mt.Fail(func(t Testing) {
		GreaterOrEqual(t, 1, 2)
		GreaterOrEqual(t, "a", "b")
		GreaterOrEqual(t, 0.1, 0.2)
		GreaterOrEqual(t, uint32(1), uint32(2))
	})
}

func TestErrorIs(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		ErrorIs(t, io.EOF, io.EOF)
		ErrorIs(t, fmt.Errorf("%w ,ok", io.EOF), io.EOF)
	})
	mt.Fail(func(t Testing) {
		ErrorIs(t, nil, io.EOF)
		ErrorIs(t, io.EOF, fmt.Errorf("%w ,ok", io.EOF))
	})
}

func TestNotErrorIs(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		NotErrorIs(t, nil, io.EOF)
		NotErrorIs(t, io.EOF, fmt.Errorf("%w ,ok", io.EOF))
	})
	mt.Fail(func(t Testing) {
		NotErrorIs(t, io.EOF, io.EOF)
	})
}

func TestLen(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Len(t, "1", 1)
		Len(t, []string{}, 0)
		Len(t, []string{"a"}, 1)

		type ss []string
		Len(t, ss{}, 0)
		Len(t, ss{"a"}, 1)
	})

	mt.Fail(func(t Testing) {
		Len(t, 0, 0)
		Len(t, []string{}, 1)
		Len(t, []string{"a"}, 0)

		type ss []string
		Len(t, ss{}, 1)
		Len(t, ss{"a"}, 2)
	})
}

func TestPanic(t *testing.T) {
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		Panic(t, func() {
			panic("ok")
		})
	})
	mt.Fail(func(t Testing) {
		Panic(t, func() {})
	})
}

func TestHasPrefix(t *testing.T) {
	type str1 string
	type str2 []byte
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		HasPrefix(t, "abc", "a")
		HasPrefix(t, "abc", "ab")
		HasPrefix(t, "abc", "abc")
		HasPrefix[str1](t, str1("abc"), str1("a"))
		HasPrefix[str2](t, str2("abc"), str2("a"))
	})
	mt.Fail(func(t Testing) {
		HasPrefix(t, "abc", "b")
		HasPrefix(t, "abc", "c")
		HasPrefix(t, "abc", "abcd")
		HasPrefix[str1](t, str1("abc"), str1("b"))
		HasPrefix[str2](t, str2("abc"), str2("b"))
	})
}

func TestNotPrefix(t *testing.T) {
	type str1 string
	type str2 []byte
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		NotPrefix(t, "abc", "b")
		NotPrefix(t, "abc", "c")
		NotPrefix(t, "abc", "abcd")
		NotPrefix[str1](t, str1("abc"), str1("b"))
		NotPrefix[str2](t, str2("abc"), str2("b"))
	})
	mt.Fail(func(t Testing) {
		NotPrefix(t, "abc", "a")
		NotPrefix(t, "abc", "ab")
		NotPrefix(t, "abc", "abc")
		NotPrefix[str1](t, str1("abc"), str1("a"))
		NotPrefix[str2](t, str2("abc"), str2("a"))
	})
}

func TestHasSuffix(t *testing.T) {
	type str1 string
	type str2 []byte
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		HasSuffix(t, "abc", "c")
		HasSuffix(t, "abc", "bc")
		HasSuffix(t, "abc", "abc")
		HasSuffix[str1](t, str1("abc"), str1("c"))
		HasSuffix[str2](t, str2("abc"), str2("c"))
	})
	mt.Fail(func(t Testing) {
		HasSuffix(t, "abc", "a")
		HasSuffix(t, "abc", "ab")
		HasSuffix(t, "abc", "abcd")
		HasSuffix[str1](t, str1("abc"), str1("b"))
		HasSuffix[str2](t, str2("abc"), str2("b"))
	})
}

func TestNotSuffix(t *testing.T) {
	type str1 string
	type str2 []byte
	mt := newMyTesting(t)
	mt.Success(func(t Testing) {
		NotSuffix(t, "abc", "a")
		NotSuffix(t, "abc", "ab")
		NotSuffix(t, "abc", "abcd")
		NotSuffix[str1](t, str1("abc"), str1("b"))
		NotSuffix[str2](t, str2("abc"), str2("b"))
	})
	mt.Fail(func(t Testing) {
		NotSuffix(t, "abc", "c")
		NotSuffix(t, "abc", "bc")
		NotSuffix(t, "abc", "abc")
		NotSuffix[str1](t, str1("abc"), str1("c"))
		NotSuffix[str2](t, str2("abc"), str2("c"))
	})
}
