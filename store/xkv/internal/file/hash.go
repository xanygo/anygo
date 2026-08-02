package file

import (
	"context"
	"encoding/json"
	"io/fs"
	"strconv"
	"unsafe"

	"github.com/xanygo/anygo/safely"
)

type hashKV struct {
	Field string `json:"f"`
	Value []byte `json:"v"` // 序列化采用的 json，为了让二进制支持的更好，需要使用 []byte 而不是 string
}

func (f hashKV) String() string {
	bf, _ := json.Marshal(f)
	return string(bf)
}

func (f hashKV) ValueString() string {
	return unsafe.String(unsafe.SliceData(f.Value), len(f.Value))
}

type Hash struct {
	Compact func()
	Base    *Base
}

func (h *Hash) saveField(field, value string) error {
	kv := hashKV{
		Field: field,
		Value: unsafe.Slice(unsafe.StringData(value), len(value)),
	}
	return h.Base.writeMemberFile(h.Base.md5(field), kv.String())
}

func (h *Hash) deleteField(field string) error {
	return h.Base.deleteMemberFile(h.Base.md5(field))
}

func (h *Hash) readFieldValue(field string) (string, bool, error) {
	str, found, err := h.Base.readMemberFile(h.Base.md5(field))
	if err != nil || !found {
		return "", false, err
	}
	kv := &hashKV{}
	err = json.Unmarshal([]byte(str), kv)
	return kv.ValueString(), err == nil, err
}

func (h *Hash) readFieldFile(fp string) (*hashKV, error) {
	bf, err := h.Base.readMemberFileByPath(fp)
	if err != nil {
		return nil, err
	}
	if len(bf) == 0 {
		return nil, nil
	}
	kv := &hashKV{}
	err = json.Unmarshal(bf, kv)
	if err != nil {
		return nil, err
	}
	return kv, err
}

func (h *Hash) HSet(ctx context.Context, field, value string) error {
	return h.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		return h.saveField(field, value)
	})
}

func (h *Hash) HMSet(ctx context.Context, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	return h.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		for k, v := range values {
			if err := h.saveField(k, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *Hash) HGet(ctx context.Context, field string) (value string, found bool, err error) {
	err = h.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		value, found, err = h.readFieldValue(field)
		return err
	})
	return value, found, err
}

func (h *Hash) HMGet(ctx context.Context, fields ...string) (result map[string]string, err error) {
	if len(fields) == 0 {
		return nil, nil
	}
	err = h.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		for _, field := range fields {
			value, found, err1 := h.readFieldValue(field)
			if err1 != nil {
				return err1
			}
			if found {
				if result == nil {
					result = make(map[string]string, len(fields))
				}
				result[field] = value
			}
		}

		return err
	})
	return result, err
}

func (h *Hash) HDel(ctx context.Context, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	defer func() {
		go safely.RunVoid(h.Compact)
	}()
	return h.Base.lock(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		for _, field := range fields {
			if err := h.deleteField(field); err != nil {
				return err
			}
		}
		h.Base.deleteKeyWhenNoMember(ctx)
		return nil
	})
}

func (h *Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	return h.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return h.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			kv, err1 := h.readFieldFile(path)
			if err1 != nil || kv == nil {
				return err1
			}
			if !fn(kv.Field, kv.ValueString()) {
				return fs.SkipAll
			}
			return nil
		})
	})
}

func (h *Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := h.HRange(ctx, func(field string, value string) bool {
		result[field] = value
		return true
	})
	return result, err
}

func (h *Hash) HExists(ctx context.Context, field string) (ok bool, err error) {
	err = h.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		_, found, err1 := h.readFieldValue(field)
		if err1 != nil || !found {
			return err1
		}
		ok = true
		return nil
	})
	return ok, err
}

func (h *Hash) HIncrBy(ctx context.Context, field string, increment int64) (num int64, err error) {
	err = h.Base.lockWrite(ctx, func(ctx context.Context, meta *Meta) error {
		old, found, err1 := h.readFieldValue(field)
		if err1 != nil {
			return err1
		}
		if old == "" {
			num = 0
		} else {
			num, err1 = strconv.ParseInt(old, 10, 64)
			if err1 != nil {
				return err1
			}
		}
		num += increment
		err1 = h.saveField(field, strconv.FormatInt(num, 10))
		if !found && err1 != nil {
			h.Base.deleteKeyWhenNoMember(ctx)
		}
		return err1
	})
	return num, err
}

func (h *Hash) HLen(ctx context.Context) (num int64, err error) {
	err = h.Base.lockRead(ctx, func(ctx context.Context, meta *Meta) error {
		if meta == nil {
			return nil
		}
		return h.Base.rangeMemberFiles(ctx, func(path string, d fs.DirEntry) error {
			num++
			return nil
		})
	})
	return num, err
}
