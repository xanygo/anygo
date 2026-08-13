package model

import (
	"context"
	"testing"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
)

var _ xdb.HasTable = MPK{}

type MPK struct {
	Class string `db:"c,pk,size:255"`
	Name  string `db:"name,pk,size:255"` // mssql 需要size 才能正常创建联合主键
	Note  string `db:"note"`
}

func (m MPK) TableName() string {
	return "ut_mpk"
}

func withMPK(ctx context.Context, t *testing.T, client *xdb.Client) {
	sc := xdb.MustNewSchemaAPI(client)
	err := sc.DropTableIfExists(ctx, MPK{}.TableName())
	xt.NoError(t, err)

	err = xdb.Migrate(client, MPK{})
	xt.NoError(t, err)

	orm := xdb.NewMode[MPK](client)
	value := MPK{
		Class: "a",
		Name:  "hello",
		Note:  "Hello World",
	}
	err = orm.Insert(ctx, value)
	xt.NoError(t, err)

	old, ok, err := orm.FindByPK(ctx, value)
	xt.NoError(t, err)
	xt.True(t, ok)
	xt.Equal(t, old.Note, value.Note)
}
