package xcachex

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xmap"
	"github.com/xanygo/anygo/ds/xslice"
	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/internal/zreflect"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/store/xcache"
	"github.com/xanygo/anygo/store/xdb"
	"github.com/xanygo/anygo/store/xdb/xor"
	"github.com/xanygo/anygo/xerror"
	"github.com/xanygo/anygo/xlog"
)

var _ xcache.StringCache = (*Database)(nil)
var _ xcache.MCache[string, string] = (*Database)(nil)
var _ xcache.HasStats = (*Database)(nil)

type dbModel struct {
	ID      int64  `db:"id,auto_inc,pk"`
	TypeID  uint32 `db:"t,unique_index=t_k[1]"` // key 和 value 实际类型签名
	Key     string `db:"k,unique_index=t_k[2]"`
	Value   string `db:"v"`
	Created int64  `db:"c"`
	Updated int64  `db:"u"`
	Expires int64  `db:"e,index"` // 赋值为： time.Now().UnixMicro()
}

type Database struct {
	DB        xdb.DBCore // 必填
	TypeID    uint32     // 必填，key 和 value 实际类型签名,可以使用 GenTypeID 获取
	Table     string     // 可选，默认 xcache
	KeyPrefix string     // 可选, key 的前缀

	// Capacity Dir 近似的最大缓存个数，>0 时有效
	// 每次 GC 时，若数量超限，会按照缓存的创建时间排序，删除创建时间更靠前的
	Capacity int64

	// GC 触发过期缓存清理的间隔时间，可选，默认为 60 秒
	// 若值为 -1 则禁止自动清理
	GC time.Duration

	BGTimeout time.Duration // 可选，后台清理数据每次操作的超时时间，默认 30 秒

	compactTime xsync.Interval // 存储上一次清理的时间
	gcRunning   atomic.Bool

	cntRead   atomic.Uint64
	cntWrite  atomic.Uint64
	cntDelete atomic.Uint64
	cntHit    atomic.Uint64
}

func (d *Database) Init(param map[string]any) error {
	if d.KeyPrefix == "" {
		d.KeyPrefix, _ = xmap.GetString(param, "KeyPrefix")
	}

	if d.Table == "" {
		d.Table, _ = xmap.GetString(param, "Table")
	}

	if d.Capacity == 0 {
		d.Capacity, _ = xmap.GetInt64(param, "Capacity")
	}

	if d.GC == 0 {
		gc, ok := xmap.GetString(param, "GC")
		if ok {
			dur, err := time.ParseDuration(gc)
			if err != nil {
				return err
			}
			d.GC = dur
		}
	}

	if d.BGTimeout == 0 {
		str, ok := xmap.GetString(param, "BGTimeout")
		if ok {
			dur, err := time.ParseDuration(str)
			if err != nil {
				return err
			}
			d.BGTimeout = dur
		}
	}

	if d.DB == nil {
		service, ok := xmap.GetString(param, "Service")
		if !ok || service == "" {
			return fmt.Errorf("no Service in %v", param)
		}
		db, err := xdb.NewClientWithService(service)
		if err != nil {
			return err
		}
		d.DB = db
	}
	return nil
}

var stringTypeID = zreflect.TypeID2[string, string]()

func (d *Database) getTypeID() uint32 {
	if d.TypeID > 0 {
		return d.TypeID
	}
	return stringTypeID
}

func (d *Database) GenTypeID[K comparable, V any]() uint32 {
	return zreflect.TypeID2[K, V]()
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
	num, err := d.keysCountCtx(ctx)
	if err != nil {
		return -1
	}
	return num
}

func (d *Database) keysCountCtx(ctx context.Context) (int64, error) {
	num, err := d.orm().Count(ctx, "*", xor.Where("t=?", d.getTypeID()))
	if err != nil {
		xlog.Warn(ctx, "DatabaseCache keysCountCtx error",
			xlog.Err("error", err),
			xlog.String("table", d.getTable()),
		)
	}
	return num, err
}

func (d *Database) getBGTimeout() time.Duration {
	if d.BGTimeout != 0 {
		return d.BGTimeout
	}
	return 30 * time.Second
}

func (d *Database) orm() *xor.Model[*dbModel] {
	orm := xor.New[*dbModel](d.DB)
	orm.Table(d.getTable())
	return orm
}

func (d *Database) Has(ctx context.Context, key string) (bool, error) {
	defer d.autoCompact()

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
	defer d.autoCompact()
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
	defer d.autoCompact()

	now := time.Now()
	cond := &xdb.Condition{}
	cond.And("t=?", d.getTypeID())
	cond.And("k=?", key)
	cond.And("e>?", now.UnixMicro()) // 只有在有效期内，才允许续期

	orm := d.orm()
	ne := now.Add(life)
	_, err := orm.ModifyFirst(ctx, func(value *dbModel) (*dbModel, error) {
		value.Expires = ne.UnixMicro()
		return value, nil
	}, xor.WhereByCond(cond))
	return err
}

func (d *Database) get(ctx context.Context, key string) (*dbModel, error) {
	orm := d.orm()
	value, found, err := orm.First(ctx, xor.Columns("id", "e", "v"), xor.Where("t=? and k=?", d.getTypeID(), key))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, xerror.NotFound
	}
	now := time.Now().UnixMicro()
	if value.Expires < now {
		go safely.RunCtxVoid(ctx, func(ctx context.Context) {
			d.deleteExpired(ctx, true, 0, value.ID)
		})
		return nil, xerror.NotFound
	}
	return value, nil
}

func (d *Database) Get(ctx context.Context, key string) (value string, err error) {
	defer d.autoCompact()

	d.cntRead.Add(1)
	item, err := d.get(ctx, key)
	if err != nil {
		return value, err
	}
	d.cntHit.Add(1)
	return item.Value, nil
}

func (d *Database) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	defer d.autoCompact()

	d.cntWrite.Add(1)
	now := time.Now().UnixMicro()
	expires := time.Now().Add(ttl).UnixMicro()
	item := &dbModel{
		TypeID:  d.getTypeID(),
		Key:     d.fullKey(key),
		Value:   value,
		Created: now,
		Updated: now,
		Expires: expires,
	}
	orm := d.orm()
	_, err := orm.Upsert(ctx, []string{"t", "k"}, []string{"v", "u", "e"}, item)
	return err
}

func (d *Database) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	defer d.autoCompact()
	d.cntDelete.Add(uint64(len(keys)))
	cond := &xdb.Condition{}
	cond.And("t=?", d.getTypeID())
	cond.AndInFmt("k in (%s)", xslice.ToAnys(keys))
	orm := d.orm()
	_, err := orm.Delete(ctx, xor.WhereByCond(cond))
	return err
}

func (d *Database) MSet(ctx context.Context, values map[string]string, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}
	defer d.autoCompact()
	d.cntWrite.Add(uint64(len(values)))

	now := time.Now().UnixMicro()
	expires := time.Now().Add(ttl).UnixMicro()
	orm := d.orm()
	items := make([]*dbModel, 0, len(values))
	for k, v := range values {
		item := &dbModel{
			TypeID:  d.getTypeID(),
			Key:     d.fullKey(k),
			Value:   v,
			Created: now,
			Updated: now,
			Expires: expires,
		}
		items = append(items, item)
	}
	_, err := orm.Upsert(ctx, []string{"t", "k"}, []string{"v", "u", "e"}, items...)
	return err
}

func (d *Database) MGet(ctx context.Context, keys ...string) (result map[string]string, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	defer d.autoCompact()
	d.cntRead.Add(uint64(len(keys)))

	cond := &xdb.Condition{}
	cond.And("t=?", d.getTypeID())
	cond.AndInFmt("k in (%s)", xslice.ToAnys(keys))

	orm := d.orm()
	items, err := orm.List(ctx, xor.Columns("id", "k", "v", "e"), xor.WhereByCond(cond), xor.Limit(len(keys)))
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMicro()
	result = make(map[string]string, len(items))
	var expires []int64
	for _, item := range items {
		if item.Expires > now {
			result[item.Key] = item.Value
		} else {
			expires = append(expires, item.ID)
		}
	}
	d.cntHit.Add(uint64(len(result)))
	if len(expires) > 0 {
		go safely.RunCtxVoid(ctx, func(ctx context.Context) {
			d.deleteExpired(ctx, true, 0, expires...)
		})
	}
	return result, nil
}

func (d *Database) deleteExpired(ctx context.Context, onlyExpire bool, batchNum int, ids ...int64) (total int64, err error) {
	if len(ids) == 0 {
		return 0, nil
	}
	ctx = context.WithoutCancel(ctx)
	ctx, cancel := context.WithTimeout(ctx, d.getBGTimeout())
	defer cancel()
	if batchNum <= 0 {
		batchNum = min(len(ids), 1000)
	}

	var errs []error
	for items := range slices.Chunk(ids, batchNum) {
		cond := &xdb.Condition{}
		cond.And("t=?", d.getTypeID())
		cond.AndInFmt("id in (%s)", xslice.ToAnys(items))

		if onlyExpire {
			cond.And("e <= ?", time.Now().UnixMicro())
		}

		num, err := d.orm().Delete(ctx, xor.WhereByCond(cond))
		if err != nil {
			errs = append(errs, err)
		}
		total += num
	}
	return total, errors.Join(errs...)
}

// ClearExpired 清理过期数据的方法，需要主动调用
//
// limit: 本次调用最多删除的总条数。若为 0，则清理所有过期数据。
// batchNum: 会采用多次批量删除的方式，每次最多删除 batchNum 条数据
func (d *Database) ClearExpired(ctx context.Context, limit int, batchNum int) (int64, error) {
	if limit <= 0 {
		limit = math.MaxInt
	}
	start := time.Now()
	deleted, err := d.doClear(ctx, int64(limit), true, batchNum)
	xlog.Info(ctx, "xcache.Database.ClearExpired",
		xlog.Cost(start),
		xlog.Err("error", err),
		xlog.Int64("deleted", deleted),
	)
	if err != nil {
		xlog.Error(ctx, "xcache.Database.ClearExpired",
			xlog.Cost(start),
			xlog.Err("error", err),
			xlog.Int64("deleted", deleted),
		)
	}
	return deleted, err
}

// ClearWithCapacity 按照容量清楚数据，若超过指定容量，则删除先创建的数据
func (d *Database) ClearWithCapacity(ctx context.Context, capacity int64, batchNum int) (int64, error) {
	num, err := d.keysCountCtx(ctx)
	if err != nil {
		return 0, err
	}
	needDelete := num - capacity
	if needDelete <= 0 {
		return 0, nil
	}
	start := time.Now()
	deleted, err := d.doClear(ctx, needDelete, false, batchNum)
	xlog.Info(ctx, "xcache.Database.ClearWithCapacity",
		xlog.Cost(start),
		xlog.Err("error", err),
		xlog.Int64("deleted", deleted),
	)
	if err != nil {
		xlog.Error(ctx, "xcache.Database.ClearWithCapacity",
			xlog.Cost(start),
			xlog.Err("error", err),
			xlog.Int64("deleted", deleted),
		)
	}
	return deleted, err
}

func (d *Database) doClear(ctx context.Context, needDelete int64, onlyExpire bool, batchNum int) (int64, error) {
	if batchNum <= 0 {
		batchNum = 1000
	}

	var deleted int64
	var lastID int64
	orm := d.orm()

	now := time.Now().UnixMicro()

	var col xor.Option
	if onlyExpire {
		col = xor.Columns("id", "e")
	} else {
		xor.Columns("id")
	}

	for needDelete > 0 {
		select {
		case <-ctx.Done():
			return deleted, context.Cause(ctx)
		default:
			//  继续
		}

		cond := &xdb.Condition{}
		cond.And("t=?", d.getTypeID())
		cond.And("id>?", lastID)

		if onlyExpire {
			cond.And("e<=?", now)
		}

		items, err := orm.List(ctx, col,
			xor.WhereByCond(cond),
			xor.OrderBy("id asc"),
			xor.Limit(min(5000, max(needDelete, int64(batchNum)))),
		)
		if err != nil || len(items) == 0 {
			return deleted, err
		}

		expires := make([]int64, 0, len(items))
		for _, item := range items {
			if onlyExpire {
				if item.Expires <= now {
					expires = append(expires, item.ID)
				}
			} else {
				expires = append(expires, item.ID)
			}
			lastID = item.ID
		}
		num, err := d.deleteExpired(ctx, true, batchNum, expires...)
		if err != nil {
			return deleted, err
		}
		deleted += num
		needDelete -= int64(len(items))
	}
	return deleted, nil
}

func (d *Database) getGC() time.Duration {
	if d.GC > time.Second {
		return d.GC
	}
	return 60 * time.Second
}

func (d *Database) autoCompact() {
	if d.GC < 0 {
		return
	}
	if !d.compactTime.Allow(d.getGC()) {
		return
	}
	go safely.Run(d.doCompact)
}

func (d *Database) doCompact() {
	if !d.gcRunning.CompareAndSwap(false, true) {
		return
	}
	defer d.gcRunning.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), d.getBGTimeout())
	defer cancel()

	_, err := d.ClearExpired(ctx, 0, 1000)
	if err != nil {
		return
	}

	if d.Capacity > 0 {
		d.ClearWithCapacity(ctx, d.Capacity, 1000)
	}
}

// Migrate 在测试环境下使用，创建表结构
func (d *Database) Migrate(ctx context.Context) error {
	obj := dbModel{}
	return xor.MigrateWithTable(ctx, d.DB, obj, d.getTable())
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
