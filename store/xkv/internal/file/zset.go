package file

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"unsafe"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xkv/internal"
)

type ZSet struct {
	Compact func()
	Base
}

func (f ZSet) ZAdd(ctx context.Context, score float64, member string) error {
	if err := f.SaveMeta(internal.DataTypeZSet); err != nil {
		return err
	}
	m := fileZSetMember{
		Member: unsafe.Slice(unsafe.StringData(member), len(member)),
		Score:  score,
	}
	bf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return f.WriteKVDataFile(f.Md5(member), string(bf))
}

func (f ZSet) ZScore(ctx context.Context, member string) (float64, bool, error) {
	str, found, err := f.CheckReadKVDataFile(f.Md5(member), internal.DataTypeZSet, false)
	if err != nil || !found {
		return 0, false, err
	}
	m := &fileZSetMember{}
	bf := unsafe.Slice(unsafe.StringData(str), len(str))
	err = json.Unmarshal(bf, m)
	return m.Score, err == nil, err
}

func (f ZSet) ZRange(ctx context.Context, fn func(member string, score float64) bool) error {
	var list []*fileZSetMember
	err := f.RangeKVFiles(ctx, internal.DataTypeZSet, func(path string, d fs.DirEntry) error {
		bf, err := os.ReadFile(filepath.Join(f.Dir, d.Name()))
		if err != nil {
			return err
		}
		m := &fileZSetMember{}
		err = json.Unmarshal(bf, m)
		if err == nil {
			list = append(list, m)
		}
		return err
	})
	if err != nil {
		return err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score < list[j].Score
	})
	for _, m := range list {
		if !fn(m.MemberString(), m.Score) {
			return nil
		}
	}
	return err
}

func (f ZSet) ZRem(ctx context.Context, members ...string) error {
	var errs []error
	for _, member := range members {
		if err := f.DeleteKVDataFile(f.Md5(member)); err != nil {
			errs = append(errs, err)
		}
	}
	go safely.RunVoid(f.Compact)
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

type fileZSetMember struct {
	Member []byte  `json:"m"`
	Score  float64 `json:"s"`
}

func (fm fileZSetMember) MemberString() string {
	return unsafe.String(unsafe.SliceData(fm.Member), len(fm.Member))
}
