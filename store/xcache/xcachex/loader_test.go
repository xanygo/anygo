//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-9-5

package xcachex

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xanygo/anygo/xt"
)

func TestConfigFileLoad(t *testing.T) {
	cf := &ConfigFile{
		Items: []map[string]any{
			{
				"Name": "cache1",
				"Type": "File",
				"Dir":  filepath.Join(t.TempDir(), "cache1"),
			},
			{
				"Name":     "cache2",
				"Type":     "MemoryLRU",
				"Capacity": 12345,
			},
			{
				"Name":     "cache3",
				"Type":     "MemoryFIFO",
				"Capacity": 12345,
			},
			{
				"Name":     "cache4",
				"Type":     "MemoryLIFO",
				"Capacity": 12345,
			},
			{
				"Name": "cache5",
				"Type": "Nop",
			},
			{
				"Name": "cache6",
				"Type": "Chains", // 必填，缓存类型。链式多级缓存
				"Chains": []map[string]any{ // 必填。应包含 >=1 个有效值
					{
						"Ref":          "cache2", // 必填，引用的数据库名称，在此配置中已经定义好的
						"Life":         "1800s",  // 必填，缓存有效期
						"WriteTimeout": "3s",     // 可选，异步写超时时间
					},
					{
						"Ref":          "cache4", // 必填，引用的数据库名称，在此配置中已经定义好的
						"Life":         "3600s",  // 必填，但是最后一个对象，此值不用
						"WriteTimeout": "3s",     // 可选
					},
				},
			},
			{
				"Name": "cache7",
				"Type": "Wrap",   // 必填，缓存类型。链式多级缓存
				"Ref":  "cache1", // 必填，引用的数据库名称，在此配置中已经定义好的
				"Life": "1800s",  // 可选，强制设置的缓存有效期
				"KeyTransform": map[string]any{ // 可选，对缓存的 key 做变换处理
					"string": map[string]any{ // 可选，对于 key 的类型是 string 的调用，可以添加前缀和后缀
						"Prefix": "prefix_", // 可选，给 key 添加前缀
						"Suffix": "_suffix", // 可选，给 key 添加后缀
					},
					"Default": map[string]any{ // 可选，对于没有找到的情况。
						"Refuse": true, // 可选，拒绝。让 Cache 调用报错
						"Panic":  true, // 可选，拒绝。让 Cache 调用 panic，在 Refuse 前判断
					},
				},
			},
		},
	}

	t.Run("case 1", func(t *testing.T) {
		for i := 1; i <= 7; i++ {
			name := fmt.Sprintf("cache%d", i)
			cc, err := cf.Load[string, string](name)
			xt.NoError(t, err)
			xt.NotNil(t, cc)
		}
	})
}
