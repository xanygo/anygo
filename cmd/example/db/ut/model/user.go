//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package model

import (
	"context"
	"testing"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xt"
)

var _ xor.HasTable = User{}

type User struct {
	ID           uint64    `db:"id,pk,auto_inc"`
	Email        string    // 不添加 db 标签
	Username     string    `db:"username,unique_index,size=200"`
	Password     string    `db:"password,not-null"`
	Status       Status    `db:"status,not-null"`
	RegisterTime time.Time `db:"register_time,codec=date_time,default=fn|CURRENT_TIMESTAMP"`
	Idx          *int64    `db:"idx,not-null"`
	Scores       []int     `db:"scores,codec=auto_json"`
	Enable       bool      `db:"enable,not-null"`
	Version      int64     `db:"version,auto=Incr"`
	a            int
	UserEmb1
	JS1     *UserJS1  `db:"js1,codec=json"`
	Created time.Time `db:"created,auto=Created"`
	Updated time.Time `db:"updated,auto=Updated"`
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

	err = xor.Migrate(ctx, client, User{})
	xt.NoError(t, err)

	orm := xor.New[User](client)
	num := int64(1)
	u := User{
		Password:     "demo",
		Username:     "user1",
		Idx:          &num,
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

	items, err := orm.List(ctx, xor.WhereAll())
	xt.NoError(t, err)
	xt.NotEmpty(t, items)

	u.Status = 2
	ret, err := orm.Update(ctx, u, xor.Where("id=?", id))
	xt.NoError(t, err)
	xt.Equal(t, ret, 1)

	u2 := User{
		ID:       uint64(id),
		Password: "hello",
	}
	ret, err = orm.UpdateByPK(ctx, u2)
	xt.NoError(t, err)
	xt.Equal(t, ret, 1)

	cnt, err := orm.Count(ctx, "id", xor.WhereAll())
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

	t.Run("ModifyFirstByPK", func(t *testing.T) {
		num, err := orm.ModifyFirstByPK(ctx, u2, func(nv User) (User, error) {
			return nv, xerror.SkipOne
		})
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		num, err = orm.ModifyFirstByPK(ctx, u2, func(nv User) (User, error) {
			nv.Username = "user-3000"
			return nv, nil
		})
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
	})

	t.Run("where-bool", func(t *testing.T) {
		u1 := User{
			Username: "enable-true",
			Enable:   true,
		}
		err1 := orm.Insert(ctx, u1)
		xt.NoError(t, err1)
		list, err1 := orm.List(ctx, xor.Where("enable=?", true))
		xt.NoError(t, err1)
		xt.NotEmpty(t, list)
	})
}
