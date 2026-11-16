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
