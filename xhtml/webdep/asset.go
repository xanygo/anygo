package webdep

import (
	"fmt"
	"html"
	"slices"
	"strings"
)

// AssetType 表示 Web 资源的类型。
type AssetType string

const (
	AssetJS  AssetType = "js"
	AssetCSS AssetType = "css"
)

const (
	CrossOriginAnonymous      = "anonymous"
	CrossOriginUseCredentials = "use-credentials"
)

// Asset 描述一个 Web 静态资源及其加载属性。
type Asset struct {
	// Name 是资源的唯一名称，用于注册、查找和声明依赖。
	// 同一注册表(Registry)中，资源名称应保持唯一。
	Name string

	// Type 指定资源类型，支持 AssetJS 和 AssetCSS。
	Type AssetType

	// Tags 资源的标签
	Tags []string

	// URL 是资源的访问地址，可以是 CDN 地址、相对路径或绝对路径。
	URL string

	// Integrity 指定资源的 Subresource Integrity（SRI）校验值。
	// 通常使用 sha256、sha384 或 sha512 算法生成。
	// 非空时，Render 会输出 integrity 属性。
	Integrity string

	// CrossOrigin 指定跨域资源的请求方式。
	// 支持 anonymous 和 use-credentials。
	// 当 Integrity 非空且 CrossOrigin 为空时，默认使用 anonymous。
	CrossOrigin string

	// Media 指定 CSS 资源适用的媒体类型或媒体查询条件。
	// 例如 all、screen、print 或 (max-width: 768px)。
	// 仅适用于 AssetCSS。
	Media string

	// Defer 指定 JavaScript 资源是否使用 defer 属性。
	// 设置为 true 时，脚本将在 HTML 解析完成后执行。
	// 仅适用于 AssetJS。
	Defer bool

	// Async 指定 JavaScript 资源是否使用 async 属性。
	// 设置为 true 时，脚本下载完成后立即执行，不保证执行顺序。
	// 仅适用于 AssetJS。
	Async bool

	// DependsOn 指定当前资源依赖的其他资源名称。
	// 渲染时会先输出依赖资源，再输出当前资源。
	// 支持多个依赖，依赖资源会自动去重。
	DependsOn []string
}

func (a Asset) String() string {
	b := &strings.Builder{}
	a.writeBuilder(b)
	return b.String()
}

func (a Asset) Clone() Asset {
	a.DependsOn = slices.Clone(a.DependsOn)
	return a
}

func (a Asset) writeBuilder(b *strings.Builder) {
	switch a.Type {
	case AssetCSS:
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(a.URL))
		b.WriteString(`"`)

		if a.Media != "" {
			b.WriteString(` media="`)
			b.WriteString(html.EscapeString(a.Media))
			b.WriteString(`"`)
		}

		a.writeSRI(b)
		b.WriteString(">")

	case AssetJS:
		b.WriteString(`<script src="`)
		b.WriteString(html.EscapeString(a.URL))
		b.WriteString(`"`)

		a.writeSRI(b)

		if a.Defer {
			b.WriteString(" defer")
		}
		if a.Async {
			b.WriteString(" async")
		}

		b.WriteString("></script>")
	default:
		fmt.Fprintf(b, "unsupported AssetType %q", a.Type)
	}
}

// writeSRI 输出 SRI 相关属性
func (a Asset) writeSRI(b *strings.Builder) {
	if a.Integrity != "" {
		b.WriteString(` integrity="`)
		b.WriteString(html.EscapeString(a.Integrity))
		b.WriteString(`"`)
	}

	crossOrigin := a.CrossOrigin
	if crossOrigin == "" && a.Integrity != "" {
		crossOrigin = CrossOriginAnonymous
	}

	if crossOrigin != "" {
		b.WriteString(` crossorigin="`)
		b.WriteString(html.EscapeString(crossOrigin))
		b.WriteString(`"`)
	}
}
