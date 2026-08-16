package xstruct

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestParserTag(t *testing.T) {
	tag1 := ParserTag("csv1,codec=csv,null")
	xt.Equal(t, tag1, Tag{
		name: "csv1",
		values: map[string]string{
			"codec": "csv",
			"null":  "",
		},
	})
	xt.Equal(t, tag1.Name(), "csv1")
	xt.Equal(t, tag1.Values(), map[string]string{"codec": "csv", "null": ""})
	xt.True(t, tag1.Has("codec"))
	xt.True(t, tag1.Has("null"))
	xt.Equal(t, tag1.Value("codec"), "csv")
}
