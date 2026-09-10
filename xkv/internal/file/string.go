package file

import (
	"context"
	"strconv"
)

type String struct {
	Base *Base
}

func (s *String) Set(ctx context.Context, value string) error {
	return s.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		err := s.saveValue(value)
		if meta == nil && err != nil {
			// 只有之前就不存在的情况下，才需要删除
			// 此时，只有 meta 写成功
			_ = s.Base.deleteKey()
		}
		return err
	})
}

func (s *String) saveValue(value string) error {
	return s.Base.writeMemberFile("value", value)
}

func (s *String) readValue() (string, bool, error) {
	return s.Base.readMemberFile("value")
}

func (s *String) delete() error {
	return s.Base.deleteKey()
}

func (s *String) SetNX(ctx context.Context, value string) (ok bool, err error) {
	err = s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta != nil {
			return nil
		}
		meta = s.Base.metaOrNew(nil)
		if err1 := s.Base.saveMeta(meta); err1 != nil {
			return err1
		}
		if err2 := s.saveValue(value); err2 != nil {
			_ = s.Base.deleteKey()
			return err2
		}
		ok = true
		return nil
	})
	return ok, err
}

func (s *String) Get(ctx context.Context) (val string, found bool, err error) {
	err = s.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		v, ok, err1 := s.readValue()
		if err1 == nil {
			val = v
			found = ok
		}
		return err1
	})
	return val, found, err
}

func (s *String) GetDel(ctx context.Context) (value string, found bool, err error) {
	err = s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		v, ok, err1 := s.readValue()
		if err1 != nil {
			return err1
		}
		if err2 := s.delete(); err2 != nil {
			return err2
		}
		value = v
		found = ok
		return nil
	})
	return value, found, err
}

func (s *String) GetSet(ctx context.Context, txt string) (value string, found bool, err error) {
	err = s.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		has := meta != nil
		old, ok, err1 := s.readValue()
		if err1 != nil {
			return err1
		}
		meta = s.Base.metaOrNew(meta)
		if err2 := s.Base.saveMeta(meta); err2 != nil {
			return err2
		}
		err3 := s.saveValue(txt)
		if err3 != nil {
			if !has {
				_ = s.delete()
			}
			return err3
		}
		value = old
		found = ok
		return nil
	})
	return value, found, err
}

func (s *String) Incr(ctx context.Context) (num int64, err error) {
	return s.IncrBy(ctx, 1)
}

func (s *String) IncrBy(ctx context.Context, incr int64) (num int64, err error) {
	err = s.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		str, hasOld, err1 := s.readValue()
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
		err2 := s.saveValue(strconv.FormatInt(num, 10))
		if err2 != nil && !hasOld {
			_ = s.delete()
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (s *String) IncrByFloat(ctx context.Context, incr float64) (num float64, err error) {
	err = s.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		str, hasOld, err1 := s.readValue()
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
		err2 := s.saveValue(strconv.FormatFloat(num, 'g', -1, 64))
		if err2 != nil && !hasOld {
			_ = s.delete()
		}
		return err2
	})
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (s *String) Decr(ctx context.Context) (num int64, err error) {
	return s.IncrBy(ctx, -1)
}
