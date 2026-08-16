package zreflect

import (
	"reflect"
)

// CallStringMethod 获取类型 T 或者 *T 上指定无参 string 方法的返回值。
//
// 例如：
//
//	type User struct{}
//
//	func (User) TableName() string {
//		return "user"
//	}
func CallStringMethod(v any, method string) string {
	if v == nil {
		return ""
	}

	rv := reflect.ValueOf(v)
	return callStringMethodRecursive(rv, method)
}

func callStringMethodRecursive(v reflect.Value, method string) string {
	if !v.IsValid() {
		return ""
	}

	// 如果当前是 nil pointer，先构造一个对应的零值对象。
	//
	// *User(nil)
	//     ↓
	// *User(&User{})
	//
	// **User(nil)
	//     ↓
	// **User(new(*User))
	//
	// 后续继续向下查找。
	if v.Kind() == reflect.Pointer && v.IsNil() {
		v = reflect.New(v.Type().Elem())
	}

	if result, ok := callStringMethod(v, method); ok {
		return result
	}

	// 当前类型没有方法时，如果是非 pointer，再尝试 *T。
	//
	// 解决：
	//
	//	type User struct{}
	//	func (*User) TableName() string
	//	var u User
	//
	// 此时 User 没有 TableName，但 *User 有。
	if v.Kind() != reflect.Pointer {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)

		result, _ := callStringMethod(ptr, method)
		return result
	}

	// 当前是 pointer，继续向下解引用。
	//
	// 例如：
	//
	//	**User
	//	   ↓
	//	*User
	//	   ↓
	//	User
	if !v.IsNil() {
		return callStringMethodRecursive(v.Elem(), method)
	}

	// 理论上这里不会到达，因为前面已经将 nil pointer,转换成了非 nil 的零值 pointer。
	return ""
}

func callStringMethod(v reflect.Value, method string) (string, bool) {
	m := v.MethodByName(method)
	if !m.IsValid() {
		return "", false
	}

	mt := m.Type()

	if mt.NumIn() != 0 || mt.NumOut() != 1 || mt.Out(0).Kind() != reflect.String {
		return "", false
	}
	return m.Call(nil)[0].String(), true
}
