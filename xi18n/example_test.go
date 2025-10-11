//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-10

package xi18n_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/xanygo/anygo/xi18n"
)

func ExampleXI() {
	b := &xi18n.Bundle{}
	// 在 index 这个名字空间下，为 zh 和 en 两种语言定义 title 这个资源
	b.MustLocalize(xi18n.LangZh).MustAdd("index", &xi18n.Message{Key: "title", Other: "你好 {0}"})
	b.MustLocalize(xi18n.LangEn).MustAdd("index", &xi18n.Message{Key: "title", Other: "hello {0}"})

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 context 里读取到 Bundle 和语言支持列表，并渲染内容
		txt := xi18n.XI(r.Context(), "index@title", "AnyGo")
		_, _ = w.Write([]byte(txt))
	})

	// 解析 Accept-Language 的中间件
	handler = (&xi18n.HTTPLanguageHandler{
		Bundle: b,
	}).Next(handler)
	ts := httptest.NewServer(handler)

	send := func(accept string) {
		fmt.Println("Accept-Language:", accept)

		reqZh, _ := http.NewRequest("GET", ts.URL, nil)
		reqZh.Header.Set("Accept-Language", accept)
		resp, err := ts.Client().Do(reqZh)
		if err == nil {
			defer resp.Body.Close()
			content, _ := io.ReadAll(resp.Body)
			fmt.Println("resp:", string(content))
		} else {
			fmt.Println("err:", err)
		}
	}

	// 模拟优先 【中文】 的浏览器请求
	send("zh-CN,zh;q=0.9,en;q=0.8") // 输出 resp: 你好 AnyGo

	fmt.Println("----")

	// 模拟优先 【英文】 的浏览器请求
	send("en,en-US;q=0.9,zh;q=0.8") // 输出 resp: hello AnyGo

	// Output:
	// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
	// resp: 你好 AnyGo
	// ----
	// Accept-Language: en,en-US;q=0.9,zh;q=0.8
	// resp: hello AnyGo
}
