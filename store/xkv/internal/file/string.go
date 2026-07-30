package file

import (
	"context"
	"strconv"
)

type String struct {
	Base *Base
}

func (f *String) Set(ctx context.Context, value string) error {
	return f.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		err := f.saveValue(value)
		if meta == nil && err != nil {
			// 只有之前就不存在的情况下，才需要删除
			// 此时，只有 meta 写成功
			_ = f.Base.deleteKey()
		}
		return err
	})
}

func (f *String) saveValue(value string) error {
	return f.Base.writeMemberFile("value", value)
}

func (f *String) readValue() (string, bool, error) {
	return f.Base.readMemberFile("value")
}

func (f *String) delete() error {
	return f.Base.deleteKey()
}

func (f *String) SetNX(ctx context.Context, value string) (ok bool, err error) {
	err = f.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta != nil {
			return nil
		}
		meta = f.Base.metaOrNew(nil)
		if err1 := f.Base.saveMeta(meta); err1 != nil {
			return err1
		}
		if err2 := f.saveValue(value); err2 != nil {
			_ = f.Base.deleteKey()
			return err2
		}
		ok = true
		return nil
	})
	return ok, err
}

func (f *String) Get(ctx context.Context) (val string, found bool, err error) {
	err = f.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		v, ok, err1 := f.readValue()
		if err1 == nil {
			val = v
			found = ok
		}
		return err1
	})
	return val, found, err
}

func (f *String) GetDel(ctx context.Context) (value string, found bool, err error) {
	err = f.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		v, ok, err1 := f.readValue()
		if err1 != nil {
			return err1
		}
		if err2 := f.delete(); err2 != nil {
			return err2
		}
		value = v
		found = ok
		return nil
	})
	return value, found, err
}

func (f *String) GetSet(ctx context.Context, txt string) (value string, found bool, err error) {
	err = f.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		has := meta != nil
		old, ok, err1 := f.readValue()
		if err1 != nil {
			return err1
		}
		meta = f.Base.metaOrNew(meta)
		if err2 := f.Base.saveMeta(meta); err2 != nil {
			return err2
		}
		err3 := f.saveValue(txt)
		if err3 != nil {
			if !has {
				_ = f.delete()
			}
			return err3
		}
		value = old
		found = ok
		return nil
	})
	return value, found, err
}

func (f *String) Incr(ctx context.Context) (num int64, err error) {
	return f.IncrBy(ctx, 1)
}

func (f *String) IncrBy(ctx context.Context, incr int64) (num int64, err error) {
	err = f.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		str, hasOld, err1 := f.readValue()
		if err1 != nil {
			return err1
		}
		if str == "" {
			num = 0
		} else {
			num, err1 = strconv.ParseInt(str, 10, 64)
			if err1 != nil {
				return err1
			}
		}
		num = num + incr
		err2 := f.saveValue(strconv.FormatInt(num, 10))
		if err2 != nil && !hasOld {
			_ = f.delete()
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (f *String) IncrByFloat(ctx context.Context, incr float64) (num float64, err error) {
	err = f.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		str, hasOld, err1 := f.readValue()
		if err1 != nil {
			return err1
		}
		if str == "" {
			num = 0
		} else {
			num, err1 = strconv.ParseFloat(str, 64)
			if err1 != nil {
				return err1
			}
		}
		num = num + incr
		err2 := f.saveValue(strconv.FormatFloat(num, 'g', -1, 64))
		if err2 != nil && !hasOld {
			_ = f.delete()
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (f *String) Decr(ctx context.Context) (num int64, err error) {
	return f.IncrBy(ctx, -1)
}
