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
					xt.Equal(t, []int{0}, r1.Values())
				case 1:
					xt.Equal(t, []int{0, 1}, r1.Values())
				case 2:
					xt.Equal(t, []int{0, 1, 2}, r1.Values())
				case 3:
					xt.Equal(t, []int{1, 2, 3}, r1.Values())
				case 4:
					xt.Equal(t, []int{2, 3, 4}, r1.Values())
				case 5:
					xt.Equal(t, []int{3, 4, 5}, r1.Values())
				}
				if i < 2 {
					xt.Equal(t, i+1, r1.Len())
				} else {
					xt.Equal(t, 3, r1.Len())
				}
			}
			// 0,1,2 | 3,4,5 | 6,7,8 | 9
			want2 := []int{7, 8, 9}
			xt.Equal(t, want2, r1.Values())

			r1.Clear()
			xt.Empty(t, r1.Values())
			xt.Equal(t, 0, r1.Len())
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
					xt.Equal(t, []int{0}, r1.Values())
					xt.Equal(t, 0, old)
					xt.False(t, swapped)
				case 1:
					xt.Equal(t, []int{0, 1}, r1.Values())
					xt.Equal(t, 0, old)
					xt.False(t, swapped)
				case 2:
					xt.Equal(t, []int{0, 1, 2}, r1.Values())
					xt.Equal(t, 0, old)
					xt.False(t, swapped)
				case 3:
					xt.Equal(t, []int{1, 2, 3}, r1.Values())
					xt.Equal(t, 0, old)
					xt.True(t, swapped)
				case 4:
					xt.Equal(t, []int{2, 3, 4}, r1.Values())
					xt.Equal(t, 1, old)
					xt.True(t, swapped)
				}

				if i < 2 {
					xt.Equal(t, i+1, r1.Len())
				} else {
					xt.Equal(t, 3, r1.Len())
				}
			}
			// 0,1,2 | 3,4,5 | 6,7,8 | 9
			want2 := []int{7, 8, 9}
			xt.Equal(t, want2, r1.Values())
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
		xt.Equal(t, wants, gots)
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
				xt.Equal(t, []int{0}, r1.Values())
			case 1:
				xt.Equal(t, []int{0, 1}, r1.Values())
			case 2:
				xt.Equal(t, []int{0, 1, 2}, r1.Values())
			case 3:
				xt.Equal(t, []int{1, 2, 3}, r1.Values())
			case 4:
				xt.Equal(t, []int{2, 3, 4}, r1.Values())
			}

			if i < 2 {
				xt.Equal(t, i+1, r1.Len())
			} else {
				xt.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		// want1 := []int{9, 7, 8}
		// xt.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		xt.Equal(t, want2, r1.Values())

		r1.Clear()
		xt.Empty(t, r1.Values())
		xt.Equal(t, 0, r1.Len())
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
				xt.Equal(t, []int{0}, r1.Values())
				xt.Equal(t, 0, old)
				xt.False(t, swapped)
			case 1:
				xt.Equal(t, []int{0, 1}, r1.Values())
				xt.Equal(t, 0, old)
				xt.False(t, swapped)
			case 2:
				xt.Equal(t, []int{0, 1, 2}, r1.Values())
				xt.Equal(t, 0, old)
				xt.False(t, swapped)
			case 3:
				xt.Equal(t, []int{1, 2, 3}, r1.Values())
				xt.Equal(t, 0, old)
				xt.True(t, swapped)
			case 4:
				xt.Equal(t, []int{2, 3, 4}, r1.Values())
				xt.Equal(t, 1, old)
				xt.True(t, swapped)
			}

			if i < 2 {
				xt.Equal(t, i+1, r1.Len())
			} else {
				xt.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		// want1 := []int{9, 7, 8}
		// xt.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		xt.Equal(t, want2, r1.Values())

		r1.Clear()
		xt.Empty(t, r1.Values())
		xt.Equal(t, 0, r1.Len())
	}
	check(t, NewUniqRing[int](3))
	check(t, NewSyncUniqRing[int](3))
}

func TestRingUnique3(t *testing.T) {
	check := func(t *testing.T, r1 testRingType[int]) {
		for range 10 {
			r1.Push(1)
			xt.Equal(t, []int{1}, r1.Values())
		}
		for range 10 {
			old, swapped := r1.PushSwap(1)
			xt.Equal(t, []int{1}, r1.Values())
			xt.Equal(t, 1, old)
			xt.True(t, swapped)
			xt.Equal(t, []int{1}, r1.Values())
		}

		{
			old, swapped := r1.PushSwap(2)
			xt.Equal(t, 0, old)
			xt.Equal(t, 2, r1.Len())
			xt.False(t, swapped)
			xt.Equal(t, []int{1, 2}, r1.Values())
		}

		{
			old, swapped := r1.PushSwap(3)
			xt.Equal(t, 0, old)
			xt.False(t, swapped)
			xt.Equal(t, []int{1, 2, 3}, r1.Values())
		}
		{
			old, swapped := r1.PushSwap(4)
			xt.Equal(t, []int{2, 3, 4}, r1.Values())
			xt.Equal(t, 1, old)
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
		xt.Equal(t, 0, got)
		xt.Equal(t, 0, r1.Len())

		r1.Push(1)
		xt.Equal(t, []int{1}, r1.Values())
		xt.Equal(t, 1, r1.Len())

		r1.Push(2)
		xt.Equal(t, []int{1, 2}, r1.Values())
		xt.Equal(t, 2, r1.Len())

		r1.Push(3)
		xt.Equal(t, []int{1, 2, 3}, r1.Values())
		xt.Equal(t, 3, r1.Len())

		r1.Push(4)
		xt.Equal(t, []int{2, 3, 4}, r1.Values())
		xt.Equal(t, 3, r1.Len())

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, 2, got)
		xt.Equal(t, []int{3, 4}, r1.Values())
		xt.Equal(t, 2, r1.Len())

		r1.Push(5)
		xt.Equal(t, []int{3, 4, 5}, r1.Values())

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, 3, got)
		xt.Equal(t, []int{4, 5}, r1.Values())

		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, 4, got)
		xt.Equal(t, []int{5}, r1.Values())

		r1.Push(6)
		xt.Equal(t, []int{5, 6}, r1.Values())
		got, ok = r1.Pop()
		xt.True(t, ok)
		xt.Equal(t, 5, got)
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
	xt.Equal(t, "123", w.String())
	w.WriteString("4")
	xt.Equal(t, "234", w.String())
	w.Reset()
	xt.Equal(t, "", w.String())
	w.Write([]byte("5"))
	xt.Equal(t, "5", w.String())
}
