package zerror

import (
	"errors"
	"fmt"
)

var (
	ErrBreak   = errors.New("break")                 // 中断当前逻辑
	ErrSkipOne = fmt.Errorf("%w skip one", ErrBreak) // 跳过当前一条数据
	ErrSkipAll = fmt.Errorf("%w skip all", ErrBreak) // 跳过所有数据
)

var ErrInvalidType = errors.New("invalid type") // 错误的数据类型
