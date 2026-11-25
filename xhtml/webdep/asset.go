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
	AssetJS AssetType = "js"

	// AssetCSS 样式表
	AssetCSS AssetType = "css"

	// HTML 链接声明（通过 rel 指定具体关系）
	AssetLink AssetType = "link"
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

	// HTML 标签的附加属性。
	Attrs map[string]string

	// Depends 指定当前资源依赖的其他资源名称。
	// 渲染时会先输出依赖资源，再输出当前资源。
	// 支持多个依赖，依赖资源会自动去重。
	Depends []string
}

func (a Asset) String() string {
	b := &strings.Builder{}
	a.writeBuilder(b)
	return b.String()
}

func (a Asset) Clone() Asset {
	a.Depends = slices.Clone(a.Depends)
	return a
}

func (a Asset) writeBuilder(b *strings.Builder) {
	switch a.Type {
	case AssetCSS:
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(a.URL))
		b.WriteString(`"`)
		a.writeAttrs(b)
		b.WriteString(">")

	case AssetJS:
		b.WriteString(`<script src="`)
		b.WriteString(html.EscapeString(a.URL))
		b.WriteString(`"`)
		a.writeAttrs(b)
		b.WriteString("></script>")

	case AssetLink:
		b.WriteString(`<link href="`)
		b.WriteString(html.EscapeString(a.URL))
		b.WriteString(`" `)
		a.writeAttrs(b)
		b.WriteString(">")

	default:
		fmt.Fprintf(b, "unsupported AssetType %q", a.Type)
	}
}

// writeSRI 输出 SRI 相关属性
func (a Asset) writeAttrs(b *strings.Builder) {
	if len(a.Attrs) == 0 {
		return
	}
	for k, v := range a.Attrs {
		b.WriteString(" ")
		b.WriteString(k)
		if v != "" {
			b.WriteString(`=`)
			fmt.Fprintf(b, "%q", v)
		}
	}

	if _, has1 := a.Attrs["integrity"]; has1 {
		if _, has2 := a.Attrs["crossorigin"]; !has2 {
			b.WriteString(` crossorigin="anonymous"`)
		}
	}
}
