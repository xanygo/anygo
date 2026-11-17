package internal

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestRuleString(t *testing.T) {
	a := Rule{Name: "required"}
	xt.Equal(t, a.String(), "required")
	b := Rule{Name: "min", Param: "10"}
	xt.Equal(t, b.String(), `min="10"`)
}
