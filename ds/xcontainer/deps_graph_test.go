package xcontainer

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestDepGraph(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		deps := &DepGraph[string]{}
		xt.NoError(t, deps.Add("a", "b"))
		xt.NoError(t, deps.Add("a", "b"))
		xt.Error(t, deps.Add("b", "a"))

		xt.True(t, deps.Has("a", "b"))
		xt.False(t, deps.Has("b", "a"))

		xt.False(t, deps.Remove("a", "not-found"))

		xt.True(t, deps.Remove("a", "b"))
		xt.False(t, deps.Has("a", "b"))

		deps.RemoveNode("c")

		xt.False(t, deps.Has("a", "b"))
		xt.NoError(t, deps.Add("a", "b"))
		xt.True(t, deps.Has("a", "b"))

		deps.RemoveNode("a")
		xt.False(t, deps.Has("a", "b"))
	})
}
