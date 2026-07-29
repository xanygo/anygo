package file

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
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

func (h *Hash) HSet(ctx context.Context, field, value string) error {
	if err := h.Base.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	kv := hashKV{
		Field: field,
		Value: unsafe.Slice(unsafe.StringData(value), len(value)),
	}
	return h.Base.WriteKVDataFile(h.Base.Md5(field), kv.String())
}

func (h *Hash) HMSet(ctx context.Context, values map[string]string) error {
	if err := h.Base.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	var errs []error
	for k, v := range values {
		kv := hashKV{
			Field: k,
			Value: []byte(v),
		}
		if err := h.Base.WriteKVDataFile(h.Base.Md5(k), kv.String()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (h *Hash) HGet(ctx context.Context, field string) (string, bool, error) {
	str, found, err := h.Base.CheckReadKVDataFile(h.Base.Md5(field), internal.DataTypeHash, false)
	if err != nil || !found {
		return "", false, err
	}
	kv := &hashKV{}
	err = json.Unmarshal([]byte(str), kv)
	if err != nil {
		return "", false, err
	}
	return kv.ValueString(), true, nil
}

func (h *Hash) HDel(ctx context.Context, fields ...string) error {
	var errs []error
	for _, field := range fields {
		if err := h.Base.DeleteKVDataFile(h.Base.Md5(field)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(h.Compact)
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (h *Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	err := h.Base.RangeKVFiles(ctx, internal.DataTypeHash, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(h.Base.Dir, d.Name()))
		if err != nil {
			return err
		}
		kv := &hashKV{}
		err = json.Unmarshal(bf, kv)
		if err != nil {
			return err
		}
		if !fn(kv.Field, kv.ValueString()) {
			return fs.SkipAll
		}
		return nil
	})
	return err
}

func (h *Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := h.HRange(ctx, func(field string, value string) bool {
		result[field] = value
		return true
	})
	return result, err
}

func (h *Hash) HExists(ctx context.Context, field string) (bool, error) {
	_, found, err := h.Base.CheckReadKVDataFile(h.Base.Md5(field), internal.DataTypeHash, false)
	if err != nil || !found {
		return false, err
	}
	return true, nil
}

func (h *Hash) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	old, found, err := h.HGet(ctx, field)
	if err != nil {
		return 0, err
	}
	var num int64
	if !found {
		num = increment
	} else {
		oldNum, err := strconv.ParseInt(old, 10, 64)
		if err != nil {
			return 0, err
		}
		num = oldNum + increment
	}
	err = h.HSet(ctx, field, strconv.FormatInt(num, 10))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (h *Hash) HLen(ctx context.Context) (num int64, err error) {
	err = h.Base.RangeKVFiles(ctx, internal.DataTypeHash, func(path string, d fs.DirEntry) error {
		num++
		return nil
	})
	return num, err
}
