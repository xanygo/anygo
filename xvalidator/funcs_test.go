package xvalidator_test

import (
	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xvalidator"
	"testing"
)

func TestCheckHTTPURL(t *testing.T) {
	xt.NoError(t, xvalidator.CheckHTTPURL("http://example.com"))
	xt.NoError(t, xvalidator.CheckHTTPURL("https://example.com"))
	xt.Error(t, xvalidator.CheckHTTPURL("ftp://example.com"))
}

func TestCheckStringIn(t *testing.T) {
	xt.NoError(t, xvalidator.CheckStringIn("hello", "hello"))
	xt.NoError(t, xvalidator.CheckStringIn("hello", "abc", "hello"))

	xt.Error(t, xvalidator.CheckStringIn("hello", "abc", "world"))
}

func TestCheckMapHasKeys(t *testing.T) {
	mp := map[string]any{
		"k1": "hello",
		"k2": "hello",
	}
	xt.NoError(t, xvalidator.CheckMapHasKeys(mp, "k1"))
	xt.NoError(t, xvalidator.CheckMapHasKeys(mp, "k1"))
	xt.NoError(t, xvalidator.CheckMapHasKeys(mp, "k1", "k2"))

	xt.Error(t, xvalidator.CheckMapHasKeys(mp, "k1", "k2", "k3"))
	xt.Error(t, xvalidator.CheckMapHasKeys(mp, "k0", "k3"))
	xt.Error(t, xvalidator.CheckMapHasKeys(mp))
}
