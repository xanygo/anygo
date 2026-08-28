package xcachex

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/xerror"
)

var _ xcache.StringCache = (*Database)(nil)
var _ xcache.MCache[string, string] = (*Database)(nil)
var _ xcache.HasStats = (*Database)(nil)

type dbModel struct {
	Key     string `db:"k,pk"`
	Value   string `db:"v"`
	Created int64  `db:"c"`
	Updated int64  `db:"u"`
	Expires int64  `db:"e,index"` // 赋值为： time.Now().UnixMicro()
}

type Database struct {
	DB        xdb.DBCore    // 必填
	Table     string        // 可选，默认 xcache
	KeyPrefix string        // 可选
	BGTimeout time.Duration // 可选，后台清理数据每次操作的超时时间，默认 1 秒

	cntRead   atomic.Uint64
	cntWrite  atomic.Uint64
	cntDelete atomic.Uint64
	cntHit    atomic.Uint64
}

func (d *Database) fullKey(key string) string {
	return d.KeyPrefix + key
}

func (d *Database) getTable() string {
	if d.Table != "" {
		return d.Table
	}
	return "xcache"
}

func (d *Database) keysCount() int64 {
	ctx, cancel := context.WithTimeout(context.Background(), d.getBGTimeout())
	defer cancel()
	orm := d.orm()
	num, err := orm.Count(ctx, "*", "")
	if err != nil {
		return -1
	}
	return num
}

func (d *Database) getBGTimeout() time.Duration {
	if d.BGTimeout != 0 {
		return d.BGTimeout
	}
	return time.Second
}

func (d *Database) orm() *xdb.Model[*dbModel] {
	orm := xdb.NewMode[*dbModel](d.DB)
	orm.Table(d.getTable())
	return orm
}

func (d *Database) Has(ctx context.Context, key string) (bool, error) {
	item, err := d.get(ctx, key)
	if err != nil {
		if errors.Is(err, xerror.NotFound) {
			return false, nil
		}
		return false, err
	}
	return item != nil, nil
}

func (d *Database) TTL(ctx context.Context, key string) (time.Duration, error) {
	item, err := d.get(ctx, key)
	if err != nil {
		if errors.Is(err, xerror.NotFound) {
			return 0, nil
		}
		return 0, err
	}
	ttl := time.Duration(item.Expires-time.Now().UnixMicro()) * time.Microsecond
	return max(ttl, 0), nil
}

func (d *Database) Expire(ctx context.Context, key string, life time.Duration) error {
	now := time.Now()
	cond := xdb.Condition{}
	cond.And("k=?", key)
	cond.And("e>?", now.UnixMicro())

	where, args := cond.MustBuild()
	orm := d.orm()
	ne := now.Add(life)
	_, err := orm.ModifyFirst(ctx, func(value *dbModel) (*dbModel, error) {
		value.Expires = ne.UnixMicro()
		return value, nil
	}, where, args...)
	return err
}

func (d *Database) get(ctx context.Context, key string) (*dbModel, error) {
	orm := d.orm()
	orm.SetSelectFields("e", "v")
	value, found, err := orm.First(ctx, "k=?", key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, xerror.NotFound
	}
	now := time.Now().UnixMicro()
	if value.Expires < now {
		go safely.RunCtxVoid(ctx, func(ctx context.Context) {
			ctx = context.WithoutCancel(ctx)
			ctx, cancel := context.WithTimeout(ctx, d.getBGTimeout())
			defer cancel()
			_, _ = orm.Delete(ctx, "k=? and e <= ?", key, now)
		})
		return nil, xerror.NotFound
	}
	return value, nil
}

func (d *Database) Get(ctx context.Context, key string) (value string, err error) {
	d.cntRead.Add(1)
	item, err := d.get(ctx, key)
	if err != nil {
		return value, err
	}
	d.cntHit.Add(1)
	return item.Value, nil
}

func (d *Database) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	d.cntWrite.Add(1)
	now := time.Now().UnixMicro()
	expires := time.Now().Add(ttl).UnixMicro()
	item := &dbModel{
		Key:     d.fullKey(key),
		Value:   value,
		Created: now,
		Updated: now,
		Expires: expires,
	}
	orm := d.orm()
	_, err := orm.Upsert(ctx, []string{"k"}, []string{"v", "u", "e"}, item)
	return err
}

func (d *Database) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	d.cntDelete.Add(uint64(len(keys)))
	cond := xdb.Condition{}
	cond.AndInFmt("k in (%s)", xslice.ToAnys(keys))
	where, args, err := cond.Build()
	if err != nil {
		return err
	}
	orm := d.orm()
	_, err = orm.Delete(ctx, where, args...)
	return err
}

func (d *Database) MSet(ctx context.Context, values map[string]string, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}
	d.cntWrite.Add(uint64(len(values)))

	now := time.Now().UnixMicro()
	expires := time.Now().Add(ttl).UnixMicro()
	orm := d.orm()
	items := make([]*dbModel, 0, len(values))
	for k, v := range values {
		item := &dbModel{
			Key:     d.fullKey(k),
			Value:   v,
			Created: now,
			Updated: now,
			Expires: expires,
		}
		items = append(items, item)
	}
	_, err := orm.Upsert(ctx, []string{"k"}, []string{"v", "u", "e"}, items...)
	return err
}

func (d *Database) MGet(ctx context.Context, keys ...string) (result map[string]string, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	d.cntRead.Add(uint64(len(keys)))

	cond := xdb.Condition{}
	cond.AndInFmt("k in (%s)", xslice.ToAnys(keys))
	where, args, err := cond.Build()
	if err != nil {
		return nil, err
	}
	orm := d.orm()
	orm.SetSelectFields("k", "v", "e").Limit(len(keys))
	items, err := orm.List(ctx, where, args...)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMicro()
	result = make(map[string]string, len(items))
	var expires []string
	for _, item := range items {
		if item.Expires > now {
			result[item.Key] = item.Value
		} else {
			expires = append(expires, item.Key)
		}
	}
	d.cntHit.Add(uint64(len(result)))
	if len(expires) > 0 {
		go safely.RunCtxVoid(ctx, func(ctx context.Context) {
			d.deleteExpired(ctx, expires...)
		})
	}
	return result, nil
}

func (d *Database) deleteExpired(ctx context.Context, keys ...string) (int64, error) {
	ctx = context.WithoutCancel(ctx)
	ctx, cancel := context.WithTimeout(ctx, d.getBGTimeout())
	defer cancel()
	cond := xdb.Condition{}
	cond.AndInFmt("k in (%s)", xslice.ToAnys(keys))
	cond.And("e <= ?", time.Now().UnixMicro())
	where, args, err := cond.Build()
	if err != nil {
		return 0, err
	}
	orm := d.orm()
	return orm.Delete(ctx, where, args...)
}

// ClearExpired 清理过期数据的方法，需要主动调用
//
// limit: 本次调用最多删除的总条数
// batchNum: 会采用多次批量删除的方式，每次最多删除 batchNum 条数据
func (d *Database) ClearExpired(ctx context.Context, limit int, batchNum int) (int64, error) {
	var deleted int64
	orm := d.orm()
	orm.SetSelectFields("k", "e")
	needDelete := limit
	for needDelete > 0 {
		select {
		case <-ctx.Done():
			return deleted, context.Cause(ctx)
		default:
			//  继续
		}
		now := time.Now().UnixMicro()
		orm.Limit(max(needDelete, batchNum)) // 一次最多查询并且删除 batchNum 条

		ctx1, cancel1 := context.WithTimeout(ctx, d.getBGTimeout())
		items, err := orm.List(ctx1, "e <= ?", now)
		cancel1()

		if err != nil || len(items) == 0 {
			return deleted, err
		}
		keys := make([]string, 0, len(items))
		for _, item := range items {
			if item.Expires <= now {
				keys = append(keys, item.Key)
			}
		}
		_, err = d.deleteExpired(ctx, keys...)
		if err != nil {
			return deleted, err
		}
		needDelete -= len(items)
	}
	return deleted, nil
}

// Migrate 在测试环境下使用，创建表结构
func (d *Database) Migrate(ctx context.Context) error {
	obj := dbModel{}
	return xdb.MigrateWithTable(ctx, d.DB, obj, d.getTable())
}

func (d *Database) Stats() xcache.Stats {
	return xcache.Stats{
		Keys:   d.keysCount(),
		Read:   d.cntRead.Load(),
		Write:  d.cntWrite.Load(),
		Delete: d.cntDelete.Load(),
		Hit:    d.cntHit.Load(),
	}
}
