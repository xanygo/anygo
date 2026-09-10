package zloader_test

import (
	"testing"

	"github.com/xanygo/anygo/internal/zloader"
	"github.com/xanygo/anygo/xenc"
	"github.com/xanygo/anygo/xenc/xcodec"
	"github.com/xanygo/anygo/xt"
)

func TestParserCodec(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		got, err := zloader.ParserCodec(nil, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.Nil(t, got)
	})

	t.Run("case 2", func(t *testing.T) {
		got, err := zloader.ParserCodec(nil, zloader.FieldCodec, xcodec.JSON)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})

	t.Run("case 3", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.Nil(t, got)
	})
	t.Run("case 4", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, xcodec.JSON)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})
	t.Run("case 5-0", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type": "JSON",
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})

	t.Run("case 5-1", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type": "not-found",
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.Error(t, err)
		xt.Nil(t, got)
	})

	t.Run("case 6", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type":   "JSON",
				"Cipher": map[string]any{},
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})
	t.Run("case 7-1", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type": "JSON",
				"Cipher": map[string]any{
					"Type": "Base64",
				},
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})
	t.Run("case 7-2", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type": "JSON",
				"Cipher": map[string]any{
					"Type": "not-found",
					"Key":  "hello",
				},
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.Error(t, err)
		xt.Nil(t, got)
	})

	t.Run("case 8-0", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type":   "JSONV2",
				"Cipher": []map[string]any{},
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})

	checkCodec := func(t *testing.T, codec xenc.Codec) {
		a := map[string]string{"k1": "v1"}
		bf, err := codec.Marshal(a)
		xt.NoError(t, err)
		xt.NotEmpty(t, bf)
		t.Logf("Marshal output=%q", bf)

		b := map[string]string{}
		err = codec.Unmarshal(bf, &b)
		xt.NoError(t, err)
		xt.Equal(t, b, a)
	}

	t.Run("case 8-1", func(t *testing.T) {
		item := map[string]any{
			"Codec": map[string]any{
				"Type": "JSON",
				"Cipher": []map[string]any{
					{
						"Type": "AesOFB",
						"Key":  "hello",
					},
				},
			},
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
		xt.NoError(t, err)
		xt.NotNil(t, got)

		checkCodec(t, got)
	})

	t.Run("case 8-2", func(t *testing.T) {
		names := []string{"Base64", "Base62", "Base58", "Base36"}
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				item := map[string]any{
					"Codec": map[string]any{
						"Type": "JSON",
						"Cipher": []map[string]any{
							{
								"Type": "AesOFB",
								"Key":  "hello",
							},
							{
								"Type": name,
							},
						},
					},
				}
				got, err := zloader.ParserCodec(item, zloader.FieldCodec, nil)
				xt.NoError(t, err)
				xt.NotNil(t, got)

				checkCodec(t, got)
			})
		}
	})

	t.Run("case 9-0", func(t *testing.T) {
		item := map[string]any{
			"Codec": "",
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, xcodec.JSON)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})
	t.Run("case 9-1", func(t *testing.T) {
		item := map[string]any{
			"Codec": "JSON",
		}
		got, err := zloader.ParserCodec(item, zloader.FieldCodec, xcodec.JSON)
		xt.NoError(t, err)
		xt.NotNil(t, got)
	})
}
