package zhttp

import (
	"net/http"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestHeaderDifference(t *testing.T) {
	got := HeaderDifference(http.Header{
		"a": []string{"a", "b"},
		"b": []string{"c", "b"},
	}, nil)
	xt.Equal(t, got, http.Header{"a": []string{"a", "b"},
		"b": []string{"c", "b"}})

	got = HeaderDifference(http.Header{
		"a": []string{"a", "b"},
		"b": []string{"c", "b"},
	}, http.Header{"a": []string{"a"}})

	xt.Equal(t, got, http.Header{"a": []string{"b"},
		"b": []string{"c", "b"}})
}
