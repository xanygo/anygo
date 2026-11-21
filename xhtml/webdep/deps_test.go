package webdep

import (
	"html/template"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestDepsGet(t *testing.T) {
	reg := NewRegistry()
	a := Asset{
		Name:  "htmx",
		Type:  AssetJS,
		URL:   "/assets/htmx.min.js",
		Defer: true,
	}
	reg.Register(a)
	b := Asset{
		Name:      "my-js",
		Type:      AssetJS,
		URL:       "/assets/myjs.min.js",
		DependsOn: []string{"htmx"},
	}
	reg.Register(b)
	xt.Equal(t, reg.Names(), []string{"htmx", "my-js"})
	xt.Equal(t, reg.Len(), 2)

	t.Run("deps-1", func(t *testing.T) {
		deps := NewDeps(reg)
		xt.False(t, deps.Needs("htmx"))
		h1, err1 := deps.Render()
		xt.NoError(t, err1)
		xt.Empty(t, h1)

		deps.Use("htmx")
		xt.True(t, deps.Needs("htmx"))
		h2, err2 := deps.Render()
		xt.NoError(t, err2)
		xt.NotEmpty(t, h2)
		want2 := `<script src="/assets/htmx.min.js" defer></script>` + "\n"
		xt.Equal(t, h2, template.HTML(want2))

		deps.Use("my-js")
		h3, err3 := deps.Render()
		xt.NoError(t, err3)
		want3 := "<script src=\"/assets/htmx.min.js\" defer></script>\n<script src=\"/assets/myjs.min.js\"></script>\n"
		xt.Equal(t, h3, template.HTML(want3))
	})
	t.Run("deps-2", func(t *testing.T) {
		deps := NewDeps(reg)
		deps.Use("my-js")
		h3, err3 := deps.Render()
		xt.NoError(t, err3)
		want3 := "<script src=\"/assets/htmx.min.js\" defer></script>\n<script src=\"/assets/myjs.min.js\"></script>\n"
		xt.Equal(t, h3, template.HTML(want3))
	})

}
