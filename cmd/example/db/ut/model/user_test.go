//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package model

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xt"
)

var _ xdb.HasTable = User{}

type User struct {
	ID           uint64    `db:"id,pk,auto_inc"`
	Email        string    // 不添加 db 标签
	Username     string    `db:"username,unique_index,size:200"`
	Password     string    `db:"password,not-null"`
	Status       Status    `db:"status,not-null"`
	RegisterTime time.Time `db:"register_time,codec:date_time,default:fn|CURRENT_TIMESTAMP"`
	Idx          *int64    `db:"idx,not-null"`
	Scores       []int     `db:"scores,codec:auto_json"`
	Enable       bool      `db:"enable,not-null"`
	a            int
	UserEmb1
	JS1 *UserJS1 `db:"js1,not-null,codec:json"`
}

type UserEmb1 struct {
	E1 int `db:"e1,not-null"`
}

type UserJS1 struct {
	ID   int
	Name string
}

func (u User) TableName() string {
	return "ut_user"
}

type Status uint

func withUser(ctx context.Context, t *testing.T, client *xdb.Client) {
	sc := xdb.MustNewSchemaAPI(client)
	err := sc.DropTableIfExists(ctx, User{}.TableName())
	xt.NoError(t, err)

	err = xdb.Migrate(client, User{})
	xt.NoError(t, err)

	orm := xdb.NewMode[User](client)
	u := User{
		Password:     "demo",
		Username:     "user1",
		RegisterTime: time.Now(),
		Scores:       []int{1, 2, 3},
		a:            123,
	}
	id, err := orm.InsertReturningID(ctx, u)
	xt.NoError(t, err)
	xt.True(t, id == 1 || id == 0) // 目前 mssql 不能返回id
	if id == 0 {
		id = 1
	}

	items, err := orm.List(ctx, "")
	xt.NoError(t, err)
	xt.NotEmpty(t, items)

	u.Status = 2
	ret, err := orm.Update(ctx, u, "id=?", id)
	xt.NoError(t, err)
	xt.Equal(t, ret, 1)

	u2 := User{
		ID:       uint64(id),
		Password: "hello",
	}
	ret, err = orm.UpdateByPK(ctx, u2)
	xt.NoError(t, err)
	xt.Equal(t, ret, 1)

	cnt, err := orm.Count(ctx, "id", "")
	xt.NoError(t, err)
	xt.Equal(t, cnt, 1)

	u3 := User{
		Username:     "user2",
		Password:     "hello",
		RegisterTime: time.Now(),
	}
	cnt, err = orm.Upsert(ctx, []string{"username"}, []string{"register_time"}, u3)
	xt.NoError(t, err)
	xt.Equal(t, cnt, 1)
}
