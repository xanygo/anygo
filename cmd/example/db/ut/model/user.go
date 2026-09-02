//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package model

import (
	"context"
	"fmt"
	"strings"
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

	t.Run("InsertReturningID", func(t *testing.T) {
		u := User{
			Password:     "demo",
			Username:     "user1",
			Idx:          new(int64(1)),
			RegisterTime: time.Now(),
			Scores:       []int{1, 2, 3},
			a:            123,
		}
		id, err := orm.InsertReturningID(ctx, u)
		xt.NoError(t, err)
		xt.True(t, id == 1 || id == 0) // 目前 mssql 不能返回id
	})

	t.Run("list", func(t *testing.T) {
		items, err := orm.List(ctx, xor.WhereAll())
		xt.NoError(t, err)
		xt.NotEmpty(t, items)

		u := items[0]
		u.Status = 2
		ret, err := orm.Update(ctx, u, xor.Where("id=?", u.ID))
		xt.NoError(t, err)
		xt.Equal(t, ret, 1)

		u.Password = "hello"
		ret, err = orm.UpdateByPK(ctx, u)
		xt.NoError(t, err)
		xt.Equal(t, ret, 1)
	})

	t.Run("count", func(t *testing.T) {
		cnt, err := orm.Count(ctx, "id", xor.WhereAll())
		xt.NoError(t, err)
		xt.Equal(t, cnt, 1)
	})

	t.Run("Upsert", func(t *testing.T) {
		u3 := User{
			Username:     "user2",
			Password:     "hello",
			RegisterTime: time.Now(),
		}
		cnt, err := orm.Upsert(ctx, []string{"username"}, []string{"register_time"}, u3)
		xt.NoError(t, err)
		xt.Equal(t, cnt, 1)
	})

	t.Run("ModifyFirstByPK", func(t *testing.T) {
		first, err := orm.GetFirst(ctx, xor.WhereByPK(User{ID: 1}))
		xt.NoError(t, err)

		num, err := orm.ModifyFirstByPK(ctx, first, func(nv User) (User, error) {
			return nv, xerror.SkipOne
		})
		xt.NoError(t, err)
		xt.Equal(t, num, 0)

		num, err = orm.ModifyFirstByPK(ctx, first, func(nv User) (User, error) {
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

	t.Run("UpdateMap", func(t *testing.T) {
		u1 := User{
			Username: "UpdateMap-2026",
			Enable:   true,
		}
		err1 := orm.Insert(ctx, u1)
		xt.NoError(t, err1)

		data := map[string]any{"idx": xdb.Expr("{idx}+1")}
		num, err := orm.UpdateMap(ctx, data, xor.Where("username=?", u1.Username))
		xt.NoError(t, err)
		xt.Equal(t, num, 1)

		data = map[string]any{"idx": xdb.Expr("{idx}+?", 2)}
		num, err = orm.UpdateMap(ctx, data, xor.Where("username=?", u1.Username))
		xt.NoError(t, err)
		xt.Equal(t, num, 1)
	})

	t.Run("Select", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			u1 := User{
				Username: fmt.Sprintf("Select-%d", i),
				Enable:   true,
			}
			err1 := orm.Insert(ctx, u1)
			xt.NoError(t, err1)
		}
		list, err := orm.New().Select[xdb.Map](ctx, xor.WhereAll())
		xt.NoError(t, err)
		xt.NotEmpty(t, list)
		var found bool
		for _, item := range list {
			if name, ok := item["username"]; ok {
				if str, ok2 := name.(string); ok2 && strings.HasPrefix(str, "Select-") {
					found = true
				}
			}
		}
		xt.True(t, found)
	})
}
