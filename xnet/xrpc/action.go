//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-16

package xrpc

import (
	"fmt"
	"time"
)

func NewAction(name string, tryTotal int) Action {
	return Action{
		Name:     name,
		TryIndex: 0,
		TryTotal: tryTotal,
		Start:    time.Now(),
	}
}

type Action struct {
	Name     string
	TryIndex int
	TryTotal int
	Start    time.Time
	End      time.Time
}

func (t Action) TryString() string {
	return fmt.Sprintf("%d/%d", t.TryIndex, t.TryTotal)
}

func (t Action) IsEnd() bool {
	return t.TryIndex+1 >= t.TryTotal
}

func (t Action) Cost() time.Duration {
	return t.End.Sub(t.Start)
}
