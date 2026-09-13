package dbcodec_test

import (
	"testing"
	"uuid"

	"github.com/xanygo/anygo/xdb/dbcodec"
	"github.com/xanygo/anygo/xt"
)

func TestUUID(t *testing.T) {
	dc := dbcodec.UUID{}
	t.Run("case 1", func(t *testing.T) {
		u1 := uuid.NewV4()
		got, err := dc.Encode(u1)
		xt.NoError(t, err)
		xt.Len(t, got, 16)

		var u2 uuid.UUID
		xt.Empty(t, u2)

		err = dc.Decode(string(got.([]byte)), &u2)
		xt.NoError(t, err)
		xt.NotEmpty(t, u2)
	})

	t.Run("case 2", func(t *testing.T) {
		u1 := uuid.NewV7()
		got, err := dc.Encode(u1)
		xt.NoError(t, err)
		xt.Len(t, got, 16)

		var u2 uuid.UUID
		xt.Empty(t, u2)

		err = dc.Decode(string(got.([]byte)), &u2)
		xt.NoError(t, err)
		xt.NotEmpty(t, u2)
	})

	t.Run("case 3", func(t *testing.T) {
		var u2 uuid.UUID
		xt.Empty(t, u2)

		err := dc.Decode(string(u2[:]), &u2)
		xt.NoError(t, err)
		xt.Empty(t, u2)
	})
	t.Run("case 4", func(t *testing.T) {
		var u2 uuid.UUID
		xt.Empty(t, u2)

		err := dc.Decode("00000000-0000-0000-0000-000000000000", &u2)
		xt.NoError(t, err)
		xt.Empty(t, u2)
	})
}
