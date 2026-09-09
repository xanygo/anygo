package zerror

import (
	"errors"
	"fmt"
)

var ErrBreak = errors.New("break")                   // 中断当前逻辑
var ErrSkipOne = fmt.Errorf("%w skip one", ErrBreak) // 跳过当前数据
var ErrSkipAll = fmt.Errorf("%w skip all", ErrBreak) // 跳过所有数据
