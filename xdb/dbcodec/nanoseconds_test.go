package dbcodec_test

import (
	"testing"
	"time"

	"github.com/xanygo/anygo/xdb/dbcodec"
	"github.com/xanygo/anygo/xt"
)

func TestNanoseconds(t *testing.T) {
	var dc dbcodec.Nanoseconds
	t.Run("case 1", func(t *testing.T) {
		var tm time.Time
		got, err := dc.Encode(tm)
		xt.NoError(t, err)
		xt.Equal(t, got, 0)

		err = dc.Decode("0", &tm)
		xt.NoError(t, err)
		t.Logf("tm=%s IsZero=%v", tm.String(), tm.IsZero())
		xt.True(t, tm.IsZero())
	})
}
