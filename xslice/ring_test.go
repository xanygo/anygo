//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-08-25

package xslice

import (
	"iter"
	"testing"

	"github.com/xanygo/anygo/xt"
)

type testRingType[T any] interface {
	Push(values ...T)
	PushSwap(v T) (old T, swapped bool)
	Len() int
	Range(fn func(v T) bool)
	Iter() iter.Seq[T]
	Values() []T
	Clear()
}

var (
	_ testRingType[int] = (*Ring[int])(nil)
	_ testRingType[int] = (*SyncRing[int])(nil)
)

type testRingUniqueType[T comparable] interface {
	Push(values ...T)
	PushSwap(v T) (old T, swapped bool)
	Len() int
	Range(fn func(v T) bool)
	Iter() iter.Seq[T]
	Values() []T
	Clear()
}

var (
	_ testRingUniqueType[int] = (*UniqRing[int])(nil)
	_ testRingUniqueType[int] = (*SyncUniqRing[int])(nil)
)

func TestNewRing(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			xt.NotNil(t, recover())
		}()
		_ = NewRing[int](0)
	})
}

func TestNewRingSync(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			xt.NotNil(t, recover())
		}()
		_ = NewSyncRing[int](0)
	})
}

func TestRing(t *testing.T) {
	t.Run("Set-3-Values", func(t *testing.T) {
		check := func(t *testing.T, r1 testRingType[int]) {
			xt.Nil(t, r1.Values())
			for i := range 10 {
				r1.Push(i)
				switch i {
				case 0:
					xt.Equal(t, r1.Values(), []int{0})
				case 1:
					xt.Equal(t, r1.Values(), []int{0, 1})
				case 2:
					xt.Equal(t, r1.Values(), []int{0, 1, 2})
				case 3:
					xt.Equal(t, r1.Values(), []int{1, 2, 3})
				case 4:
					xt.Equal(t, r1.Values(), []int{2, 3, 4})
				case 5:
					xt.Equal(t, r1.Values(), []int{3, 4, 5})
				}
				if i < 2 {
					xt.Equal(t, r1.Len(), i+1)
				} else {
					xt.Equal(t, r1.Len(), 3)
				}
			}
			// 0,1,2 | 3,4,5 | 6,7,8 | 9
			want2 := []int{7, 8, 9}
			xt.Equal(t, r1.Values(), want2)

			r1.Clear()
			xt.Empty(t, r1.Values())
			xt.Equal(t, r1.Len(), 0)
		}
		t.Run("common", func(t *testing.T) {
			check(t, NewRing[int](3))
		})
		t.Run("sync", func(t *testing.T) {
			check(t, NewSyncRing[int](3))
		})
	})

	t.Run("AddSwap", func(t *testing.T) {
		check := func(t *testing.T, r1 testRingType[int]) {
			for i := range 10 {
				old, swapped := r1.PushSwap(i)

				switch i {
				case 0:
					xt.Equal(t, r1.Values(), []int{0})
					xt.Equal(t, old, 0)
					xt.False(t, swapped)
				case 1:
					xt.Equal(t, r1.Values(), []int{0, 1})
					xt.Equal(t, old, 0)
					xt.False(t, swapped)
				case 2:
					xt.Equal(t, r1.Values(), []int{0, 1, 2})
					xt.Equal(t, old, 0)
					xt.False(t, swapped)
				case 3:
					xt.Equal(t, r1.Values(), []int{1, 2, 3})
					xt.Equal(t, old, 0)
					xt.True(t, swapped)
				case 4:
					xt.Equal(t, r1.Values(), []int{2, 3, 4})
					xt.Equal(t, old, 1)
					xt.True(t, swapped)
				}

				if i < 2 {
					xt.Equal(t, r1.Len(), i+1)
				} else {
					xt.Equal(t, r1.Len(), 3)
				}
			}
			// 0,1,2 | 3,4,5 | 6,7,8 | 9
			want2 := []int{7, 8, 9}
			xt.Equal(t, r1.Values(), want2)
		}
		check(t, NewRing[int](3))
	})

	t.Run("iter", func(t *testing.T) {
		r1 := NewRing[int](3)
		r1.Push(1, 2, 3)
		var gots []int
		for v := range r1.Iter() {
			gots = append(gots, v)
		}
		wants := []int{1, 2, 3}
		xt.Equal(t, gots, wants)
	})
}

func TestNewRingUnique(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			xt.NotNil(t, recover())
		}()
		_ = NewUniqRing[int](0)
	})
}

func TestNewRingUniqueSync(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			xt.NotNil(t, recover())
		}()
		_ = NewSyncUniqRing[int](0)
	})
}

func TestRingUnique1(t *testing.T) {
	check := func(t *testing.T, r1 testRingType[int]) {
		xt.Nil(t, r1.Values())
		for i := range 10 {
			r1.Push(i)

			switch i {
			case 0:
				xt.Equal(t, r1.Values(), []int{0})
			case 1:
				xt.Equal(t, r1.Values(), []int{0, 1})
			case 2:
				xt.Equal(t, r1.Values(), []int{0, 1, 2})
			case 3:
				xt.Equal(t, r1.Values(), []int{1, 2, 3})
			case 4:
				xt.Equal(t, r1.Values(), []int{2, 3, 4})
			}

			if i < 2 {
				xt.Equal(t, r1.Len(), i+1)
			} else {
				xt.Equal(t, r1.Len(), 3)
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		// want1 := []int{9, 7, 8}
		// xt.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		xt.Equal(t, r1.Values(), want2)

		r1.Clear()
		xt.Empty(t, r1.Values())
		xt.Equal(t, r1.Len(), 0)
	}
	check(t, NewUniqRing[int](3))
	check(t, NewSyncUniqRing[int](3))
}

func TestRingUnique2(t *testing.T) {
	check := func(t *testing.T, r1 testRingType[int]) {
		for i := range 10 {
			old, swapped := r1.PushSwap(i)

			switch i {
			case 0:
				xt.Equal(t, r1.Values(), []int{0})
				xt.Equal(t, old, 0)
				xt.False(t, swapped)
			case 1:
				xt.Equal(t, r1.Values(), []int{0, 1})
				xt.Equal(t, old, 0)
				xt.False(t, swapped)
			case 2:
				xt.Equal(t, r1.Values(), []int{0, 1, 2})
				xt.Equal(t, old, 0)
				xt.False(t, swapped)
			case 3:
				xt.Equal(t, r1.Values(), []int{1, 2, 3})
				xt.Equal(t, old, 0)
				xt.True(t, swapped)
			case 4:
				xt.Equal(t, r1.Values(), []int{2, 3, 4})
				xt.Equal(t, old, 1)
				xt.True(t, swapped)
			}

			if i < 2 {
				xt.Equal(t, r1.Len(), i+1)
			} else {
				xt.Equal(t, r1.Len(), 3)
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		// want1 := []int{9, 7, 8}
		// xt.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		xt.Equal(t, r1.Values(), want2)

		r1.Clear()
		xt.Empty(t, r1.Values())
		xt.Equal(t, r1.Len(), 0)
	}
	check(t, NewUniqRing[int](3))
	check(t, NewSyncUniqRing[int](3))
}

func TestRingUnique3(t *testing.T) {
	check := func(t *testing.T, r1 testRingType[int]) {
		for range 10 {
			r1.Push(1)
			xt.Equal(t, r1.Values(), []int{1})
		}
		for range 10 {
			old, swapped := r1.PushSwap(1)
			xt.Equal(t, r1.Values(), []int{1})
			xt.Equal(t, old, 1)
			xt.True(t, swapped)
			xt.Equal(t, r1.Values(), []int{1})
		}

		{
			old, swapped := r1.PushSwap(2)
			xt.Equal(t, old, 0)
			xt.Equal(t, r1.Len(), 2)
			xt.False(t, swapped)
			xt.Equal(t, r1.Values(), []int{1, 2})
		}

		{
			old, swapped := r1.PushSwap(3)
			xt.Equal(t, old, 0)
			xt.False(t, swapped)
			xt.Equal(t, r1.Values(), []int{1, 2, 3})
		}
		{
			old, swapped := r1.PushSwap(4)
			xt.Equal(t, r1.Values(), []int{2, 3, 4})
			xt.Equal(t, old, 1)
			xt.True(t, swapped)
		}
	}

	check(t, NewUniqRing[int](3))
	check(t, NewSyncUniqRing[int](3))
}

func BenchmarkRingUnique(b *testing.B) {
	checkAdd := func(r1 testRingUniqueType[int]) {
		for i := range 100 {
			r1.Push(i, i+1, i+2, i+3)
		}
	}
	b.Run("non-sync-add", func(b *testing.B) {
		r := NewUniqRing[int](100)
		for i := 0; i < b.N; i++ {
			checkAdd(r)
		}
	})
	b.Run("sync-add", func(b *testing.B) {
		r := NewSyncUniqRing[int](100)
		for i := 0; i < b.N; i++ {
			checkAdd(r)
		}
	})
}

func TestRing_Pop(t *testing.T) {
	type ppv interface {
		Push(values ...int)
		Pop() (int, bool)
		Values() []int
		Len() int
	}
	check := func(t *testing.T, r1 ppv) {
		got, ok := r1.Pop()
		xt.False(t, ok)
		xt.Equal(t, got, 0)
		xt.Equal(t, r1.Len(), 0)

		r1.Push(1)
		xt.Equal(t, r1.Values(), []int{1})
		xt.Equal(t, r1.Len(), 1)

		r1.Push(2)
		xt.Equal(t, r1.Values(), []int{1, 2})
		xt.Equal(t, r1.Len(), 2)

		r1.Push(3)
		xt.Equal(t, r1.Values(), []int{1, 2, 3})
		xt.Equal(t, r1.Len(), 3)

		r1.Push(4)
		xt.Equal(t, r1.Values(), []int{2, 3, 4})
		xt.Equal(t, r1.Len(), 3)

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, got, 2)
		xt.Equal(t, r1.Values(), []int{3, 4})
		xt.Equal(t, r1.Len(), 2)

		r1.Push(5)
		xt.Equal(t, r1.Values(), []int{3, 4, 5})

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, got, 3)
		xt.Equal(t, r1.Values(), []int{4, 5})

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, got, 4)
		xt.Equal(t, r1.Values(), []int{5})

		r1.Push(6)
		xt.Equal(t, r1.Values(), []int{5, 6})
		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, got, 5)
	}
	t.Run("ring", func(t *testing.T) {
		r1 := NewRing[int](3)
		check(t, r1)
	})

	t.Run("sync-ring", func(t *testing.T) {
		r1 := NewSyncRing[int](3)
		check(t, r1)
	})
}

func TestSyncRingWriter_WriteString(t *testing.T) {
	w := NewSyncRingWriter(3)
	w.WriteString("1")
	w.WriteString("2")
	w.WriteString("3")
	xt.Equal(t, w.String(), "123")
	w.WriteString("4")
	xt.Equal(t, w.String(), "234")
	w.Reset()
	xt.Equal(t, w.String(), "")
	w.Write([]byte("5"))
	xt.Equal(t, w.String(), "5")
}
