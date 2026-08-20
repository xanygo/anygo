// Package xmime 提供额外的 mime 类型支持，如 alpine linux 上的 mime 就是精简的，在使用 go http.Server 对外提供一些静态资源的时候，
// 应用可能不能输出正确的 Content-Type，导致资源不能正常展示。
package xmime
