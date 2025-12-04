//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-20

package model

import (
	"context"
	"log"
	"time"

	"github.com/xanygo/anygo/cli/xcolor"
	"github.com/xanygo/anygo/store/xdb"
)

var _ xdb.HasTable = User{}

type User struct {
	ID           uint64    `db:"id,pk,auto_inc"`
	Email        string    // 不添加 db 标签
	Password     string    `db:"password,not-null"`
	Status       Status    `db:"status,not-null"`
	RegisterTime time.Time `db:"register_time,codec:date_time,default:fn|CURRENT_TIMESTAMP"`
	Idx          *int64    `db:"idx,not-null"`
	Scores       []int     `db:"scores,codec:auto_json,native:int[]"`
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
	return "user"
}

type Status uint

func WithUser(ctx context.Context, client *xdb.Client) {
	xdb.Migrate(client, User{})

	orm := xdb.NewMode[User](client)
	u := User{
		Password:     "demo",
		RegisterTime: time.Now(),
		Scores:       []int{1, 2, 3},
		a:            123,
	}
	id, err := orm.InsertReturningID(ctx, u)
	log.Println("insert", id, errorText(err))

	items, err := orm.List(ctx, "")
	log.Println("list=", items, errorText(err))

	ret, err := orm.Update(ctx, u, "id=?", 3)
	log.Println("Update:", ret, errorText(err))

	u2 := User{
		ID:       4,
		Password: "hello",
	}
	ret, err = orm.UpdateByPK(ctx, u2)
	log.Println("UpdateByPK:", ret, errorText(err))

	cnt, err := orm.Count(ctx, "id", "")
	log.Println("Count:", cnt, errorText(err))

	u3 := User{
		ID:           5,
		Password:     "hello",
		RegisterTime: time.Now(),
	}
	cnt, err = orm.Upsert(ctx, []string{"id"}, []string{"register_time"}, u3)
	log.Println("Upsert:", cnt, errorText(err))
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return xcolor.RedString(err.Error())
}
