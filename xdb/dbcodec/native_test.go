package dbcodec_test

import (
	"encoding"
	"errors"
	"testing"
	"time"

	"github.com/xanygo/anygo/xdb"
	"github.com/xanygo/anygo/xdb/dbcodec"
	"github.com/xanygo/anygo/xdb/dbschema"
	"github.com/xanygo/anygo/xdb/dbtype"
	"github.com/xanygo/anygo/xdb/xor"
	"github.com/xanygo/anygo/xdb/xtdr"
	"github.com/xanygo/anygo/xt"
	"github.com/xanygo/anygo/xtime"
)

type Model1 struct {
	Dur1 time.Duration  `db:"dur1"`
	Dur2 xtime.Duration `db:"dur2"`
	Dur3 MyInt          `db:"dur3"`
	Dur4 *MyInt         `db:"dur4"`
}

func (m Model1) TableName() string {
	return "model1"
}

var _ encoding.TextMarshaler = (*MyInt)(nil)
var _ encoding.TextUnmarshaler = (*MyInt)(nil)

type MyInt int64

// MyInt 实际是 int64,所以不应该调用 MarshalText 和 UnmarshalText
func (m *MyInt) MarshalText() (text []byte, err error) {
	return nil, errors.New("unexpect call MarshalText")
}

func (m *MyInt) UnmarshalText(text []byte) error {
	return errors.New("unexpect call UnmarshalText")
}

var cols = []dbtype.ColumnSchema{
	{
		Name:    "dur1",
		Kind:    dbtype.KindInt64,
		Codec:   &dbcodec.Native{},
		NotNull: true,
		Group:   []string{},
	},
	{
		Name:    "dur2",
		Kind:    dbtype.KindInt64,
		Codec:   &dbcodec.Native{},
		NotNull: true,
		Group:   []string{},
	},
	{
		Name:    "dur3",
		Kind:    dbtype.KindInt64,
		Codec:   &dbcodec.Native{},
		NotNull: true,
		Group:   []string{},
	},
	{
		Name:    "dur4",
		Kind:    dbtype.KindInt64,
		Codec:   &dbcodec.Native{},
		NotNull: true,
		Group:   []string{},
	},
}

func TestNative(t *testing.T) {
	db := xtdr.MustOpen()
	client := xdb.NewClient("mysql", "test", db)
	defer xtdr.Reset()
	orm := xor.New[Model1](client)

	t.Run("case 1", func(t *testing.T) {
		df, err := client.Dialect()
		xt.NoError(t, err)
		xt.NotNil(t, df)

		sc, err := dbschema.Schema(df, Model1{})
		xt.NoError(t, err)
		xt.NotEmpty(t, sc)
		xt.Equal(t, len(sc.Columns), len(cols))
		for i := 0; i < len(sc.Columns); i++ {
			item := sc.Columns[i]
			item.ReflectType = nil
			t.Run(item.Name, func(t *testing.T) {
				xt.Equal(t, item, cols[i])
			})
		}
	})

	t.Run("case 2", func(t *testing.T) {
		xtdr.ExpectQuery("wc:SELECT*", []string{"dur1", "dur2", "dur3", "dur4"}, [][]any{{1, 200, 3000, 40000}, {2, 2000, 5000, 6789}})
		value, found, err := orm.First(t.Context(), xor.WhereAll())
		xt.NoError(t, err)
		xt.True(t, found)
		xt.NotEmpty(t, value)
		d4 := MyInt(40000)
		want := Model1{
			Dur1: time.Duration(1),
			Dur2: xtime.Duration(200),
			Dur3: MyInt(3000),
			Dur4: &d4,
		}
		xt.Equal(t, value, want)
	})
}
