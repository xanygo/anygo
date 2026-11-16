package xvld_test

import (
	"testing"

	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xvld"
)

func TestCheckHTTPURL(t *testing.T) {
	xt.NoError(t, xvld.CheckHTTPURL("http://example.com"))
	xt.NoError(t, xvld.CheckHTTPURL("https://example.com"))
	xt.Error(t, xvld.CheckHTTPURL("ftp://example.com"))
}

func TestCheckStringIn(t *testing.T) {
	xt.NoError(t, xvld.CheckStringIn("hello", "hello"))
	xt.NoError(t, xvld.CheckStringIn("hello", "abc", "hello"))

	xt.Error(t, xvld.CheckStringIn("hello", "abc", "world"))
}

func TestCheckMapHasKeys(t *testing.T) {
	mp := map[string]any{
		"k1": "hello",
		"k2": "hello",
	}
	xt.NoError(t, xvld.CheckMapHasKeys(mp, "k1"))
	xt.NoError(t, xvld.CheckMapHasKeys(mp, "k1"))
	xt.NoError(t, xvld.CheckMapHasKeys(mp, "k1", "k2"))

	xt.Error(t, xvld.CheckMapHasKeys(mp, "k1", "k2", "k3"))
	xt.Error(t, xvld.CheckMapHasKeys(mp, "k0", "k3"))
	xt.Error(t, xvld.CheckMapHasKeys(mp))
}
