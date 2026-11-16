package xstruct

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestParserTag(t *testing.T) {
	tag1 := ParserTag("xdb", "csv1,codec=csv,null")
	xt.Equal(t, tag1, Tag{
		name: "xdb",
		keys: []string{"csv1", "codec", "null"},
		values: map[string]string{
			"csv1":  "",
			"codec": "csv",
			"null":  "",
		},
	})
	xt.Equal(t, tag1.Name(), "xdb")
	xt.Equal(t, tag1.Values(), map[string]string{"csv1": "", "codec": "csv", "null": ""})
	xt.True(t, tag1.Has("codec"))
	xt.True(t, tag1.Has("null"))
	xt.Equal(t, tag1.Value("codec"), "csv")

	tag2 := tag1.Shift()
	xt.Equal(t, tag2.Name(), "csv1")
	xt.Equal(t, tag2.Value("codec"), "csv")
}
