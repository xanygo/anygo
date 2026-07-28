package file

import (
	"context"
	"strconv"

	"github.com/xanygo/anygo/store/xkv/internal"
)

type String struct {
	Base
}

func (f *String) Set(ctx context.Context, value string) error {
	if err := f.SaveMeta(internal.DataTypeString); err != nil {
		return err
	}
	return f.WriteKVDataFile("value", value)
}

func (f *String) Get(ctx context.Context) (string, bool, error) {
	return f.CheckReadKVDataFile("value", internal.DataTypeString, false)
}

func (f *String) Incr(ctx context.Context) (num int64, err error) {
	return f.IncrBy(ctx, 1)
}

func (f *String) IncrBy(ctx context.Context, incr int64) (num int64, err error) {
	value, _, err := f.Get(ctx)
	if err != nil {
		return 0, err
	}
	if value == "" {
		num = 0
	} else {
		num, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	num = num + incr
	err = f.Set(ctx, strconv.FormatInt(num, 10))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (f *String) Decr(ctx context.Context) (num int64, err error) {
	return f.IncrBy(ctx, -1)
}
