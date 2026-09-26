package xattr

import (
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestFileConfig_getRootDir(t *testing.T) {
	c1 := FileConfig{
		SelfPath: "xxx/conf/app.yml",
	}
	xt.Equal(t, c1.getRootDir(), "xxx")

	c2 := FileConfig{
		SelfPath: "xxx/conf/product/app.yml",
	}
	xt.Equal(t, c2.getRootDir(), "xxx")

	c3 := FileConfig{
		SelfPath: "xxx/conf_product/app.yml",
	}
	xt.Equal(t, c3.getRootDir(), "xxx")
}
