package db

import (
	"context"
	"io"
	"math/rand/v2"
	"time"

	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/store/xkv"
)

type SetModel struct {
	ID         int64    `db:"id,auto_inc,pk"`
	KeyHash    [32]byte `db:"k,unique_index=k_m[1]"`
	MemberHash [32]byte `db:"m,unique_index=k_m[2]"`

	KeyRaw    string `db:"k_raw"`
	MemberRaw string `db:"m_raw"`
	Created   int64  `db:"c"`
}

var _ xkv.Set[string] = (*Set)(nil)

type Set struct {
	Table string
	Meta  *Meta
}

func (s *Set) GetTable() string {
	if s.Table == "" {
		return "xkv_set"
	}
	return s.Table
}

func (s *Set) deleteWithKey(ctx context.Context, tx xdb.DBCore) error {
	orm := xor.New[SetModel](tx)
	orm.Table(s.GetTable())
	_, err := orm.Delete(ctx, xor.Where("k=?", s.Meta.KeyHash[:]))
	return err
}

func (s *Set) SAdd(ctx context.Context, members ...string) (num int64, err error) {
	if len(members) == 0 {
		return 0, nil
	}
	now := time.Now().UnixNano()
	items := make([]SetModel, 0, len(members))
	for _, member := range members {
		item := SetModel{
			KeyHash:    s.Meta.KeyHash,
			KeyRaw:     s.Meta.KeyRaw,
			MemberHash: KeyHash(member),
			MemberRaw:  member,
			Created:    now,
		}
		items = append(items, item)
	}
	err = s.Meta.WithWriteTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		num, err = orm.Upsert(ctx, []string{"k", "m"}, nil, items...)
		return err
	})
	return num, err
}

func (s *Set) SRem(ctx context.Context, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	hashMembers := make([]any, 0, len(members))
	for _, member := range members {
		hashMembers = append(hashMembers, keyHashBytes(member))
	}
	return s.Meta.WithTx(ctx, func(ctx context.Context, tx xdb.DBCore) error {
		_, err := s.Meta.load(ctx, tx)
		if err != nil {
			return err
		}
		cond := &xdb.Condition{}
		cond.And("k=?", s.Meta.KeyHash[:])
		cond.AndInFmt("m in(%s)", hashMembers)

		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		_, err = orm.Delete(ctx, xor.WhereByCond(cond))
		if err != nil {
			return err
		}
		return s.checkExists(ctx, orm)
	})
}

// checkExists 检查 key 是否还存在，若不存在，则删除 meta
func (s *Set) checkExists(ctx context.Context, orm *xor.Model[SetModel]) error {
	orm = orm.New()
	orm.Table(s.GetTable())
	_, found, err := orm.First(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.Columns("c"))
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return s.Meta.delete(ctx, orm.DB())
}

func (s *Set) SRange(ctx context.Context, fn func(member string) bool) error {
	return s.Meta.WithReadTx(ctx, func(as context.Context, tx xdb.DBCore, hasMeta bool) error {
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		for item, err1 := range orm.ListIter(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.OrderBy("id asc"), xor.Columns("m_raw")) {
			if err1 != nil {
				return err1
			}
			if !fn(item.MemberRaw) {
				return io.EOF
			}
		}
		return nil
	})
}

func (s *Set) SMembers(ctx context.Context) ([]string, error) {
	var result []string
	err := s.SRange(ctx, func(member string) bool {
		result = append(result, member)
		return true
	})
	return result, err
}

func (s *Set) SCard(ctx context.Context) (num int64, err error) {
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		var err1 error
		num, err1 = orm.Count(ctx, "*", xor.Where("k=?", s.Meta.KeyHash[:]))
		return err1
	})
	return num, err
}

func (s *Set) SIsMember(ctx context.Context, member string) (ok bool, err error) {
	memberHash := KeyHash(member)
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		_, found, err1 := orm.First(ctx, xor.Where("k=? and m=?", s.Meta.KeyHash[:], memberHash[:]), xor.Columns("c"))
		if err1 == nil {
			ok = found
		}
		return err1
	})
	return ok, err
}

func (s *Set) SMIsMember(ctx context.Context, members []string) (oks []bool, err error) {
	oks = make([]bool, len(members))
	if len(members) == 0 {
		return oks, nil
	}
	hashMembers := make([]any, 0, len(members))
	for _, member := range members {
		hashMembers = append(hashMembers, keyHashBytes(member))
	}
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())

		cond := &xdb.Condition{}
		cond.And("k=?", s.Meta.KeyHash[:])
		cond.AndInFmt("m in (%s)", hashMembers)

		items, err1 := orm.List(ctx, xor.WhereByCond(cond), xor.Columns("m_raw"))
		if err1 == nil {
			mp := make(map[string]bool, len(items))
			for _, item := range items {
				mp[item.MemberRaw] = true
			}
			for i, member := range members {
				oks[i] = mp[member]
			}
		}
		return err1
	})
	return oks, err
}

func (s *Set) SPop(ctx context.Context) (v string, found bool, err error) {
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		total, err1 := orm.Count(ctx, "*", xor.Where("k=?", s.Meta.KeyHash[:]))
		if err1 != nil {
			return err1
		}
		if total < 1 {
			return nil
		}
		rows, err2 := orm.List(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.LimitOffset(1, rand.IntN(int(total))), xor.Columns("m", "m_raw"))
		if err2 != nil || len(rows) == 0 {
			return err2
		}
		first := rows[0]
		_, err3 := orm.Delete(ctx, xor.Where("k=? and m=?", s.Meta.KeyHash[:], first.MemberHash[:]))
		if err3 != nil {
			return err3
		}
		found = true
		v = first.MemberRaw
		return nil
	})
	return v, found, err
}

func (s *Set) SPopN(ctx context.Context, count int) (result []string, err error) {
	if count <= 0 {
		return nil, nil
	}
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		total, err1 := orm.Count(ctx, "*", xor.Where("k=?", s.Meta.KeyHash[:]))
		if err1 != nil {
			return err1
		}
		if total < 1 {
			return nil
		}
		rows, err2 := orm.List(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.Limit(count), xor.OrderByRand(), xor.Columns("m", "m_raw"))
		if err2 != nil || len(rows) == 0 {
			return err2
		}
		members := make([]string, 0, len(rows))
		hashs := make([]any, 0, len(rows))
		for _, row := range rows {
			members = append(members, row.MemberRaw)
			hashs = append(hashs, row.MemberHash[:])
		}
		cond := &xdb.Condition{}
		cond.And("k=?", s.Meta.KeyHash[:])
		cond.AndInFmt("m in (%s)", hashs)

		_, err4 := orm.Delete(ctx, xor.WhereByCond(cond))
		if err4 == nil {
			result = members
		}
		return err4
	})
	return result, err
}

func (s *Set) SRandMember(ctx context.Context) (v string, found bool, err error) {
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		total, err1 := orm.Count(ctx, "*", xor.Where("k=?", s.Meta.KeyHash[:]))
		if err1 != nil {
			return err1
		}
		if total < 1 {
			return nil
		}
		rows, err2 := orm.List(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.LimitOffset(1, rand.IntN(int(total))), xor.Columns("m_raw"))
		if err2 != nil || len(rows) == 0 {
			return err2
		}
		found = true
		v = rows[0].MemberRaw
		return nil
	})
	return v, found, err
}

func (s *Set) SRandMemberN(ctx context.Context, count int) (result []string, err error) {
	if count <= 0 {
		return nil, nil
	}
	err = s.Meta.WithReadTx(ctx, func(ctx context.Context, tx xdb.DBCore, hasMeta bool) error {
		if !hasMeta {
			return nil
		}
		orm := xor.New[SetModel](tx)
		orm.Table(s.GetTable())
		total, err1 := orm.Count(ctx, "*", xor.Where("k=?", s.Meta.KeyHash[:]))
		if err1 != nil {
			return err1
		}
		if total < 1 {
			return nil
		}
		rows, err2 := orm.List(ctx, xor.Where("k=?", s.Meta.KeyHash[:]), xor.OrderByRand(), xor.Limit(count), xor.Columns("m_raw"))
		if err2 != nil || len(rows) == 0 {
			return err2
		}
		result = make([]string, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.MemberRaw)
		}
		return nil
	})
	return result, err
}
