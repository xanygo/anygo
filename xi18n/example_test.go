//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-10

package xi18n_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/xanygo/anygo"
	"github.com/xanygo/anygo/xi18n"
)

func ExampleRA() {
	b := &xi18n.Bundle{}
	// 在 index 这个名字空间下，为 zh 和 en 两种语言定义 title 这个资源
	b.MustLocalize(xi18n.LangZh).MustAdd("index", &xi18n.Message{Key: "title", Other: "你好 {0}"})
	b.MustLocalize(xi18n.LangEn).MustAdd("index", &xi18n.Message{Key: "title", Other: "hello {0}"})

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 context 里读取到 Bundle 和语言支持列表，并渲染内容
		txt := anygo.Must1(xi18n.RA(r.Context(), "index/title", "AnyGo"))
		_, _ = w.Write([]byte(txt))
	})

	// 解析 Accept-Language 的中间件，将 Bundle 和 Language 信息绑定到 context 里去
	handler = (&xi18n.HTTPHandler{
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

func ExampleMessage_Render() {
	m := &xi18n.Message{
		Key:  "demo",
		Zero: "zero books", // 当传入数字 0 时使用。可选
		One:  "one book",   // 当传入数字 1 时使用。可选
		// Two:   "two books",  // 当传入数字 2 时使用。可选
		Few:   "few books", // 当传入数字 (2-10) 时使用。可选
		Other: "{0} books", // 当传入其他数字时使用。兜底，必填字段
	}

	// 复数功能
	fmt.Println(m.Render(0))
	fmt.Println(m.Render(1))
	fmt.Println(m.Render(2))
	fmt.Println(m.Render(3))
	fmt.Println(m.Render(20))

	// Output:
	// zero books <nil>
	// one book <nil>
	// 2 books <nil>
	// few books <nil>
	// 20 books <nil>
}
