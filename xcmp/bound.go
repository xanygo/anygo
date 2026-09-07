package xcmp

import (
	"errors"
	"fmt"
	"strconv"
)

// Bound 表示一个范围边界条件。
// 支持普通数值边界、开区间边界以及正负无穷边界。
//
// 支持 Redis 风格:
//
//	"1"     => 包含 1 的边界
//	"(1"    => 排除 1 的边界
//	"-inf"  => 负无穷边界
//	"+inf"  => 正无穷边界
type Bound[T float64 | int64] struct {
	// Value 表示边界数值。
	//
	// 当 NegInf 或 PosInf 为 true 时，该字段无效。
	Value T

	// Exclude 表示是否排除该边界值。
	//
	// false:
	//   表示闭区间边界，例如 [1
	//
	// true:
	//   表示开区间边界，例如 (1
	Exclude bool

	// Inf 表示该边界为负无穷或者正无穷
	//
	// 设置为 true 时，Value 字段无效。
	Inf bool
}

var errEmptyBound = errors.New("empty bound")

// ParserMin 解析下边界值，如 "1","(1","-inf"
func (b *Bound[T]) ParserMin(str string) error {
	*b = Bound[T]{}
	if str == "" {
		return errEmptyBound
	}
	if str == "+inf" {
		return fmt.Errorf("canot parser %q with ParserMin", str)
	}
	if str == "-inf" {
		b.Inf = true
		return nil
	}

	if str[0] == '(' {
		b.Exclude = true
		str = str[1:]
	}
	return b.parseValue(str)
}

// ParserMax 解析上边界值，，如 "1","(1","+inf"
func (b *Bound[T]) ParserMax(str string) error {
	*b = Bound[T]{}
	if str == "" {
		return errEmptyBound
	}
	if str == "-inf" {
		return fmt.Errorf("canot parser %q with ParserMax", str)
	}
	if str == "+inf" {
		b.Inf = true
		return nil
	}
	if str[0] == '(' {
		b.Exclude = true
		str = str[1:]
	}
	return b.parseValue(str)
}

func (b *Bound[T]) parseValue(str string) error {
	switch any(b.Value).(type) {
	case float64:
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		b.Value = T(v)

	case int64:
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		b.Value = T(v)

	default:
		return errors.New("unsupported bound type")
	}

	return nil
}

// MatchMin 作为下边界(min)时，判断传入的值是否在下边界之上
// 如 min=5,传入值 6，返回 true
func (b *Bound[T]) MatchMin(v T) bool {
	if b.Inf {
		return true
	}
	return v > b.Value || (!b.Exclude && v == b.Value)
}

// MatchMax 作为上边界(max)时，判断传入的值是否在上边界之下
// 如 max=10, 传入值 6，返回 true
func (b *Bound[T]) MatchMax(v T) bool {
	if b.Inf {
		return true
	}
	return v < b.Value || (!b.Exclude && v == b.Value)
}
