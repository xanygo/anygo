package xcmp

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBound_ParserMin(t *testing.T) {
	t.Run("int64-min-1", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMin("1")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Value: 1})
	})
	t.Run("int64-min-2", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMin("abc")
		xt.Error(t, err)
		xt.Equal(t, b, &Bound[int64]{})
	})
	t.Run("int64-min-3", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMin("-inf")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Inf: true})
	})
	t.Run("int64-min-4", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMin("(1")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Value: 1, Exclude: true})
	})
	t.Run("int64-max-1", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMax("1")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Value: 1})
	})
	t.Run("int64-max-2", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMax("abc")
		xt.Error(t, err)
		xt.Equal(t, b, &Bound[int64]{})
	})
	t.Run("int64-max-3", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMax("+inf")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Inf: true})
	})
	t.Run("int64-max-4", func(t *testing.T) {
		b := &Bound[int64]{}
		err := b.ParserMax("(1")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[int64]{Value: 1, Exclude: true})
	})

	t.Run("float64-min-1", func(t *testing.T) {
		b := &Bound[float64]{}
		err := b.ParserMin("1")
		xt.NoError(t, err)
		xt.Equal(t, b, &Bound[float64]{Value: 1})
	})

	t.Run("float64-min-2", func(t *testing.T) {
		b := &Bound[float64]{}
		err := b.ParserMin("abc")
		xt.Error(t, err)
		xt.Equal(t, b, &Bound[float64]{})
	})
}

func TestBound_MatchMin(t *testing.T) {
	b1 := Bound[int64]{
		Value: 2,
	}
	xt.True(t, b1.MatchMin(3))
	xt.True(t, b1.MatchMin(2))
	xt.False(t, b1.MatchMin(1))

	xt.True(t, b1.MatchMax(1))
	xt.True(t, b1.MatchMax(2))
	xt.False(t, b1.MatchMax(3))
}
