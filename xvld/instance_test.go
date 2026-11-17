package xvld_test

import (
	"testing"

	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xvld"
)

type TestUser20 struct {
	Scores []int `xv:"dive,min=2"`
}

func TestStruct(t *testing.T) {
	u1 := TestUser20{
		Scores: []int{1},
	}
	xt.Error(t, xvld.Struct(u1))
}

func TestVar(t *testing.T) {
	xt.Error(t, xvld.Var("", xvld.Rule{Name: "required"}))
	xt.NoError(t, xvld.Var("a", xvld.Rule{Name: "required"}))

	xt.Error(t, xvld.Var("a", xvld.Rule{Name: "required"}, xvld.NewRule("min-len", "2")))
	xt.NoError(t, xvld.Var("he", xvld.Rule{Name: "required"}, xvld.NewRule("min-len", "2")))

	xt.Error(t, xvld.Var(1, xvld.Rule{Name: "required"}, xvld.NewRule("min", "2")))
	xt.NoError(t, xvld.Var("2", xvld.Rule{Name: "required"}, xvld.NewRule("min", "2")))

	xt.Error(t, xvld.Var([]string{"a"}, xvld.Rule{Name: "required"}, xvld.NewRule("regexp", `^\d+$`)))
	xt.NoError(t, xvld.Var([]string{"123"}, xvld.Rule{Name: "required"}, xvld.NewRule("regexp", `^\d+$`)))
}
