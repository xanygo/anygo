package xcookiejar

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xdb"
	"github.com/xanygo/anygo/xdb/xor"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xsync"
)

var _ Storage = (*Database)(nil)

const DefaultMaxPerKey = 128

// Database 使用数据库存储 Cookie 信息的实现
type Database struct {
	Client xdb.DBCore // 必填，数据库对象
	Table  string     // 可选，表名，默认值为 xcookiejar

	// MaxPerKey 可选，一个 jar 最多允许多少个 Cookie，默认值为 128，-1 为不限制
	MaxPerKey int

	compactTime xsync.Interval // 存储上一次清理的时间
}

func (d *Database) getTable() string {
	if d.Table != "" {
		return d.Table
	}
	return "xcookiejar"
}

func (d *Database) getLimit() int {
	if d.MaxPerKey == 0 {
		return DefaultMaxPerKey
	}
	// -1 为不限制
	return d.MaxPerKey
}

// DeleteEntry implements [Storage].
func (d *Database) orm() *xor.Model[dbModel] {
	return xor.New[dbModel](d.Client).Table(d.getTable())
}

// Migrate 在测试环境下使用，创建表结构
func (d *Database) Migrate(ctx context.Context) error {
	obj := dbModel{}
	return xor.MigrateWithTable(ctx, d.Client, obj, d.getTable())
}

var dbSelectField = []string{"id", "domain", "path", "name", "expires", "entry", "created"}

// Get implements [Storage].
func (d *Database) Get(ctx context.Context, key string) ([]Entry, error) {
	defer d.autoCompact()

	orm := d.orm()
	items, err := orm.List(ctx, xor.StringColumns(dbSelectField...), xor.Where("kh=?", hash(key)))
	if err != nil || len(items) == 0 {
		return nil, err
	}

	result := make([]Entry, 0, len(items))
	selectIDs := make([]any, 0, len(items))
	var expireIDs []any
	for _, item := range items {
		if item.isExpired() {
			expireIDs = append(expireIDs, item.ID)
		} else {
			entry := item.Entry
			entry.Domain = item.Domain
			entry.Path = item.Path
			entry.Name = item.Name
			entry.SeqNum = item.ID
			entry.Creation = item.Created

			result = append(result, entry)
			selectIDs = append(selectIDs, item.ID)
		}
	}

	if len(expireIDs) > 0 {
		cond := &xdb.Condition{}
		cond.AndInFmt("id in (%s)", expireIDs)
		orm.Delete(ctx, xor.WhereByCond(cond))
	}

	if len(selectIDs) > 0 {
		cond := &xdb.Condition{}
		cond.AndInFmt("id in (%s)", selectIDs)
		orm.UpdateMap(ctx, map[string]any{"lastaccess": time.Now()}, xor.WhereByCond(cond))
	}

	return result, nil
}

// Set implements [Storage].
func (d *Database) Set(ctx context.Context, key string, items []Entry) error {
	defer d.autoCompact()

	if len(items) == 0 {
		return nil
	}
	now := time.Now()
	values := make([]dbModel, 0, len(items))
	keyHash := hash(key)
	for _, entry := range items {
		item := dbModel{
			KeyHash:    keyHash,
			EntryHash:  hash(entry.ID()),
			Domain:     entry.Domain,
			Path:       entry.Path,
			Name:       entry.Name,
			Expires:    entry.Expires,
			Created:    now,
			Updated:    now,
			LastAccess: now,
		}
		entry.Domain = ""
		entry.Path = ""
		entry.Name = ""
		entry.Expires = time.Time{}
		entry.Creation = time.Time{}

		item.Entry = entry

		if !entry.Persistent && entry.Expires.IsZero() {
			item.Expires = now.AddDate(100, 0, 0)
		}
		values = append(values, item)
	}
	_, err := d.orm().Upsert(ctx, []string{"kh", "eh"}, []string{"name", "expires", "entry", "updated", "lastaccess"}, values...)
	if err != nil {
		return err
	}
	return d.checkLimit(ctx, keyHash)
}

// checkLimit 检查并删除超过数量限制的
func (d *Database) checkLimit(ctx context.Context, keyHash [32]byte) error {
	limit := d.getLimit()
	if limit < 0 {
		return nil
	}
	orm := d.orm()
	num, err := orm.Count(ctx, "*", xor.Where("kh = ?", keyHash))
	if num <= int64(limit) || err != nil {
		return err
	}
	items, err := orm.List(ctx,
		xor.Columns("id"),
		xor.Where("kh = ?", keyHash),
		xor.OrderBy("id asc"),
		xor.Limit(num-int64(limit)),
	)
	if len(items) == 0 || err != nil {
		return err
	}
	ids := make([]any, len(items))
	for index, item := range items {
		ids[index] = item.ID
	}
	cond := &xdb.Condition{}
	cond.AndInFmt("id in (%s)", ids)
	_, err = orm.Delete(ctx, xor.WhereByCond(cond))
	return err
}

// DeleteEntry implements [Storage].
func (d *Database) DeleteEntry(ctx context.Context, key string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	ehs := make([]any, len(ids))
	for index, id := range ids {
		ehs[index] = hash(id)
	}

	cond := &xdb.Condition{}
	cond.And("kh", hash(key))
	cond.AndInFmt("eh in (%s)", ehs)
	_, err := d.orm().Delete(ctx, xor.WhereByCond(cond))
	return err
}

// DeleteKey implements [Storage].
func (d *Database) DeleteKey(ctx context.Context, key string) error {
	_, err := d.orm().Delete(ctx, xor.Where("kh=?", hash(key)))
	return err
}

func (d *Database) autoCompact() {
	if !d.compactTime.Allow(10 * time.Minute) {
		return
	}
	go safely.Run(d.compact)
}

func (d *Database) compact() {
	orm := d.orm()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	now := time.Now()
	ctx = xlog.WithContext(ctx)
	xlog.AddMetaAttr(ctx, xlog.Time("compactStart", now))
	var lastID int64
	for {
		items, err := orm.List(ctx,
			xor.Columns("id"),
			xor.Where("id>? and expires <= ?", lastID, now),
			xor.OrderBy("id asc"),
			xor.Limit(1000),
		)
		if len(items) == 0 {
			return
		}
		if err != nil {
			return
		}
		var ids []any
		for _, item := range items {
			ids = append(ids, item.ID)
			if item.ID > lastID {
				lastID = item.ID
			}
		}
		cond := &xdb.Condition{}
		cond.AndInFmt("id in (%s)", ids)
		cond.And("expires <= ?", time.Now())
		_, err = orm.Delete(ctx, xor.WhereByCond(cond))
		if err != nil {
			return
		}
	}
}

type dbModel struct {
	ID         int64     `db:"id,auto_inc,pk"`
	KeyHash    [32]byte  `db:"kh,unique_index=kh_eh[1]"`
	EntryHash  [32]byte  `db:"eh,unique_index=kh_eh[2]"`
	Domain     string    `db:"domain"`
	Path       string    `db:"path"`
	Name       string    `db:"name"`
	Expires    time.Time `db:"expires"`
	Entry      Entry     `db:"entry"`
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
	LastAccess time.Time `db:"lastaccess"` // 最后一次访问
}

func (dm dbModel) XDBTag() string {
	return "db"
}

func (dm dbModel) isExpired() bool {
	return dm.Expires.Before(time.Now())
}

func hash(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}
