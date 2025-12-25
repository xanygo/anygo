//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-24

package xlimiter

import "time"

func NewInterval(dur time.Duration) *Interval {
	return &Interval{
		tm:  time.NewTimer(dur),
		gap: dur,
	}
}

type Interval struct {
	tm  *time.Timer
	gap time.Duration
}

func (g *Interval) Allow() bool {
	select {
	case <-g.tm.C:
		g.Reset()
		return true
	default:
		return false
	}
}

func (g *Interval) Reset() {
	g.tm.Reset(g.gap)
}

func (g *Interval) Stop() {
	g.tm.Stop()
}
