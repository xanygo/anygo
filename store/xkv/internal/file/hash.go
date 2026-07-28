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

type Hash struct {
	Compact func()
	Base
}

type fileHashKV struct {
	Field string `json:"f"`
	Value []byte `json:"v"` // 序列化采用的 json，为了让二进制支持的更好，需要使用 []byte 而不是 string
}

func (f fileHashKV) String() string {
	bf, _ := json.Marshal(f)
	return string(bf)
}

func (f fileHashKV) ValueString() string {
	return unsafe.String(unsafe.SliceData(f.Value), len(f.Value))
}

func (f Hash) HSet(ctx context.Context, field, value string) error {
	if err := f.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	kv := fileHashKV{
		Field: field,
		Value: unsafe.Slice(unsafe.StringData(value), len(value)),
	}
	return f.WriteKVDataFile(f.Md5(field), kv.String())
}

func (f Hash) HMSet(ctx context.Context, values map[string]string) error {
	if err := f.SaveMeta(internal.DataTypeHash); err != nil {
		return err
	}
	var errs []error
	for k, v := range values {
		kv := fileHashKV{
			Field: k,
			Value: []byte(v),
		}
		if err := f.WriteKVDataFile(f.Md5(k), kv.String()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (f Hash) HGet(ctx context.Context, field string) (string, bool, error) {
	str, found, err := f.CheckReadKVDataFile(f.Md5(field), internal.DataTypeHash, false)
	if err != nil || !found {
		return "", false, err
	}
	kv := &fileHashKV{}
	err = json.Unmarshal([]byte(str), kv)
	if err != nil {
		return "", false, err
	}
	return kv.ValueString(), true, nil
}

func (f Hash) HDel(ctx context.Context, fields ...string) error {
	var errs []error
	for _, field := range fields {
		if err := f.DeleteKVDataFile(f.Md5(field)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(f.Compact)
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (f Hash) HRange(ctx context.Context, fn func(field string, value string) bool) error {
	err := f.RangeKVFiles(ctx, internal.DataTypeHash, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err != nil {
			return err
		}
		kv := &fileHashKV{}
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

func (f Hash) HGetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	err := f.HRange(ctx, func(field string, value string) bool {
		result[field] = value
		return true
	})
	return result, err
}

func (f Hash) HExists(ctx context.Context, field string) (bool, error) {
	_, found, err := f.CheckReadKVDataFile(f.Md5(field), internal.DataTypeHash, false)
	if err != nil || !found {
		return false, err
	}
	return true, nil
}

func (f Hash) HIncrBy(ctx context.Context, field string, increment int64) (int64, error) {
	old, found, err := f.HGet(ctx, field)
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
	err = f.HSet(ctx, field, strconv.FormatInt(num, 10))
	if err != nil {
		return 0, err
	}
	return num, nil
}
