package webdep

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestAssetString(t *testing.T) {
	t.Run("js-1", func(t *testing.T) {
		a := Asset{
			Name: "htmx",
			Type: AssetJS,
			URL:  "/assets/htmx.min.js",
			Attrs: map[string]string{
				"defer": "",
			},
		}
		got := a.String()
		want := `<script src="/assets/htmx.min.js" defer></script>`
		xt.Equal(t, got, want)
	})

	t.Run("css-1", func(t *testing.T) {
		a := Asset{
			Name: "bootstrap-css",
			Type: AssetCSS,
			URL:  "/assets/bootstrap.min.css",
			Attrs: map[string]string{
				"integrity": "hash1234",
			},
		}
		got := a.String()
		want := `<link rel="stylesheet" href="/assets/bootstrap.min.css" integrity="hash1234" crossorigin="anonymous">`
		xt.Equal(t, got, want)
	})
}
