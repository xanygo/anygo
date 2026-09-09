package xi18n

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestBundle_preferred(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		b := NewBundle(LangZh, LangEn)
		xt.True(t, b.preferred([]Language{LangZh, LangEn}))
		xt.False(t, b.preferred([]Language{LangEn, LangZh}))
		xt.False(t, b.preferred([]Language{Language("ja"), LangEn, LangZh}))
		xt.True(t, b.preferred([]Language{Language("ja"), LangZh, LangEn}))
		xt.True(t, b.preferred([]Language{Language("ja"), LangZhCN, LangEn}))
	})
}
