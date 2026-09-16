package xcookiejar

import (
	"net/http/cookiejar"

	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xkv"
	"github.com/xanygo/anygo/xkv/xkvx"
)

func New(o *cookiejar.Options) (*Jar, error) {
	db1 := xkvx.NewMemoryAny[Entry](xcodec.JSON)
	db2 := xkvx.NewMemory()
	store := &KV{
		EntryStore: func(key string) xkv.Hash[Entry] {
			return db1.Hash(key)
		},
		MetaStore: func() xkv.ZSet[string] {
			return db2.ZSet("cookiejar-meta")
		},
	}

	jar := &Jar{
		Storage: store,
	}
	if o != nil {
		jar.PSList = o.PublicSuffixList
	}
	return jar, nil
}

type Options = cookiejar.Options
