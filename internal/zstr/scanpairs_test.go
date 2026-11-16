package zstr

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestScanPairs(t *testing.T) {
	t.Run("text1", func(t *testing.T) {
		const text = `required,regexp(#)=^[a-zA-Z]{3,16}$#,min= 3 , max=20`
		var count int
		data := map[string]string{}
		err := ScanPairs(text, ',', func(key, value string) bool {
			count++
			data[key] = value
			switch key {
			case "required":
				xt.Equal(t, value, "")
			case "regexp":
				xt.Equal(t, value, "^[a-zA-Z]{3,16}$")
			case "min":
				xt.Equal(t, value, " 3 ")
			case " max":
				xt.Equal(t, value, "20")
			}
			return true
		})
		xt.Equal(t, count, 4, "data=%#v", data)
		xt.NoError(t, err)
	})
	t.Run("text2", func(t *testing.T) {
		const text = `required,min=3,max=20,`
		var count int
		ScanPairs(text, ',', func(key, value string) bool {
			count++
			switch key {
			case "required":
				xt.Equal(t, value, "")
			case "min":
				xt.Equal(t, value, "3")
			case "max":
				xt.Equal(t, value, "20")
			}
			return true
		})
		xt.Equal(t, count, 3)
	})
	t.Run("text3", func(t *testing.T) {
		const text = `required,min=3,max=20, ,`
		var count int
		ScanPairs(text, ',', func(key, value string) bool {
			count++
			switch key {
			case "required":
				xt.Equal(t, value, "")
			case "min":
				xt.Equal(t, value, "3")
			case "max":
				xt.Equal(t, value, "20")
			case " ":
				xt.Equal(t, value, "")
			}
			return true
		})
		xt.Equal(t, count, 4)
	})

	t.Run("text4", func(t *testing.T) {
		const text = `regexp(#)=\d{3,5}#`
		var count int
		ScanPairs(text, ',', func(key, value string) bool {
			count++
			switch key {
			case "regexp":
				xt.Equal(t, value, `\d{3,5}`)
			}
			return true
		})
		xt.Equal(t, count, 1)
	})

}
