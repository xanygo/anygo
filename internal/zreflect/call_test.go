package zreflect_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/xt"
)

func TestCallStringMethod(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		xt.Empty(t, zreflect.CallStringMethod("123", "Name"))
	})

	t.Run("struct", func(t *testing.T) {
		var u1 User1
		xt.Equal(t, zreflect.CallStringMethod(u1, "TableName"), "user1")

		var u2 User2
		xt.Equal(t, zreflect.CallStringMethod(u2, "TableName"), "user2")

		var u3 user3
		xt.Equal(t, zreflect.CallStringMethod(u3, "TableName"), "user3")
	})

	t.Run("ptr", func(t *testing.T) {
		u1 := &User1{}
		xt.Equal(t, zreflect.CallStringMethod(u1, "TableName"), "user1")

		u2 := &User2{}
		xt.Equal(t, zreflect.CallStringMethod(u2, "TableName"), "user2")

		u3 := &user3{}
		xt.Equal(t, zreflect.CallStringMethod(u3, "TableName"), "user3")
	})

	t.Run("nil ptr", func(t *testing.T) {
		var u1 *User1
		xt.Equal(t, zreflect.CallStringMethod(u1, "TableName"), "user1")

		var u2 *User2
		xt.Equal(t, zreflect.CallStringMethod(u2, "TableName"), "user2")

		var u3 *user3
		xt.Equal(t, zreflect.CallStringMethod(u3, "TableName"), "user3")
	})

	t.Run("nil nil ptr", func(t *testing.T) {
		var u1 **User1
		xt.Equal(t, zreflect.CallStringMethod(u1, "TableName"), "user1")

		var u2 **User2
		xt.Equal(t, zreflect.CallStringMethod(u2, "TableName"), "user2")

		var u3 **user3
		xt.Equal(t, zreflect.CallStringMethod(u3, "TableName"), "user3")
	})
}

type User1 struct{}

func (User1) TableName() string {
	return "user1"
}

type User2 struct{}

func (*User2) TableName() string {
	return "user2"
}

type user3 struct{}

func (user3) TableName() string {
	return "user3"
}
