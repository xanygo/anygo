package xtime

import (
	"time"
)

type DateInt int32

func (d DateInt) Year() int {
	return int(d / 10000)
}

func (d DateInt) Month() int {
	return int(d/100) % 100
}

func (d DateInt) Day() int {
	return int(d % 100)
}

func (d DateInt) Time() time.Time {
	if d == 0 {
		return time.Time{}
	}
	return time.Date(
		d.Year(),
		time.Month(d.Month()),
		d.Day(),
		0, 0, 0, 0,
		time.Local,
	)
}

func DateIntOf(t time.Time) DateInt {
	if t.IsZero() {
		return 0
	}
	return DateInt(t.Year()*10000 + int(t.Month())*100 + t.Day())
}
