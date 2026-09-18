//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/xanygo/anygo/xdb"
	"github.com/xanygo/anygo/xdb/dbschema"
	"github.com/xanygo/anygo/xdb/xor"
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

	Time3 time.Time `db:"time3,codec=timespan"`

	// default=fn|Now 同时支持
	// date_time，timespan，milliseconds，microseconds 这些类型
	Time4 time.Time `db:"time4,kind=date_time,default=fn|Now"`

	Time5 time.Time `db:"time5,kind=milliseconds,default=fn|Now"`
	Time6 time.Time `db:"time6,kind=microseconds,default=fn|Now"`

	Time7 time.Time `db:"time7,codec=timespan,default=fn|Now"`
	Time8 time.Time `db:"time8,codec=milliseconds,default=fn|Now"`
	Time9 time.Time `db:"time9,codec=microseconds,default=fn|Now"`

	UUID uuid.UUID `db:"uuid"`
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

func checkSchema(t *testing.T, client *xdb.Client) {
	fy, err := client.Dialect()
	xt.NoError(t, err)
	sc, err := dbschema.Schema(fy, User{})
	xt.NoError(t, err)
	xt.NotNil(t, sc)

	t.Run("register_time", func(t *testing.T) {
		rt, err := sc.ColumnByName("register_time")
		xt.NoError(t, err)
		xt.NotEmpty(t, rt)

		xt.Equal(t, rt.Codec.Name(), "date_time")
	})

	t.Run("created", func(t *testing.T) {
		rt, err := sc.ColumnByName("created")
		xt.NoError(t, err)
		xt.NotEmpty(t, rt)

		xt.Equal(t, rt.Codec.Name(), "milliseconds")
	})
}

type Status uint

func withUser(ctx context.Context, t *testing.T, client *xdb.Client) {
	t.Run("schema", func(t *testing.T) {
		checkSchema(t, client)
	})
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

	t.Run("upsert-updatecols-empty", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			const un = "upsert-empty-nil"
			_, err = orm.Delete(ctx, xor.Where("username=?", un))
			xt.NoError(t, err)

			for i := 0; i < 3; i++ {
				t.Logf("loop=%d", i)
				u3 := User{
					Username:     un,
					Password:     "hello",
					RegisterTime: time.Now(),
					Version:      int64(i),
				}
				cnt, err := orm.Upsert(ctx, []string{"username"}, nil, u3)
				xt.NoError(t, err)
				if i == 0 {
					xt.Equal(t, cnt, 1)
				} else {
					// mysql upsert 冲突更新后，影响条数是2
					if orm.DB().Driver() == "mysql" {
						xt.Equal(t, cnt, 2)
					} else {
						xt.Equal(t, cnt, 1)
					}
				}

			}
		})
		t.Run("zero-len", func(t *testing.T) {
			const un = "upsert-empty-zero-len"
			_, err = orm.Delete(ctx, xor.Where("username=?", un))
			xt.NoError(t, err)
			for i := 0; i < 3; i++ {
				t.Logf("loop=%d", i)
				u3 := User{
					Username:     un,
					Password:     "hello",
					RegisterTime: time.Now(),
					Version:      int64(i),
				}
				cnt, err := orm.Upsert(ctx, []string{"username"}, []string{}, u3)
				xt.NoError(t, err)
				if i == 0 {
					xt.Equal(t, cnt, 1)
				} else {
					xt.Equal(t, cnt, 0)
				}

			}
		})
	})

	t.Run("ModifyFirstByPK", func(t *testing.T) {
		first, err := orm.GetFirst(ctx, xor.WhereByPK(User{ID: 1}))
		xt.NoError(t, err)

		num, err := orm.ModifyFirstByPK(ctx, first, func(nv User) (User, error) {
			return nv, xerror.ErrSkipOne
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

	t.Run("timespan", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			u1 := User{
				Username:     fmt.Sprintf("timespan-%d", i),
				Time3:        time.Now(),
				RegisterTime: time.Now(),
			}
			err1 := orm.Insert(ctx, u1)
			xt.NoError(t, err1)
		}
		t.Run("register_time", func(t *testing.T) {
			// register_time 实际存储的类型是 DataTime
			list, err := orm.New().List(ctx, xor.Where("register_time>?", sql.Named("?register_time", time.Now().Add(-time.Minute))))
			xt.NoError(t, err)
			xt.NotEmpty(t, list)
		})

		t.Run("time3", func(t *testing.T) {
			// time3 实际存储的类型是 millseconds
			list, err := orm.New().List(ctx, xor.Where("time3>?", sql.Named("?time3", time.Now().Add(-time.Minute))))
			xt.NoError(t, err)
			xt.NotEmpty(t, list)
		})
	})

	t.Run("timespan-default", func(t *testing.T) {
		u := User{}
		name := "user-timespan-default"
		str := fmt.Sprintf("insert into %s (%s)values( '%s' )", orm.Quote(u.TableName()), orm.Quote("username"), name)
		ret, err := xdb.Exec(ctx, client, str)
		xt.NoError(t, err)
		xt.NotEmpty(t, ret)
	})

}
