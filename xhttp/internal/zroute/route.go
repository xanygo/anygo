//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-11

package zroute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/xlog"
)

const MethodAny = "ANY"

type PatternType int8

const (
	PatternExact  PatternType = iota // 完全匹配
	PatternWord                      // 单词通配，一个变量匹配一个目录，可以有前缀和后缀,/user/{name}/{age}, /user/hello-{id}.html
	PatternRegexp                    // 正则匹配，/user/{category}/{id:[0-9]+}, /user/{id:[0-9]+}.html,
)

var patternReg = regexp.MustCompile(`^(([A-Za-z]+(,[A-Za-z]+)*\s+)?)(/\S*)(\s+meta\|(\S*))?$`)

// splitPattern 解析 pattern 中的 Method、Path、Meta 三部分
func splitPattern(pattern string) ([]string, string, string) {
	arr := patternReg.FindStringSubmatch(pattern)
	if len(arr) == 0 {
		return nil, "", ""
	}

	methods := strings.TrimSpace(arr[1])
	if methods == "" {
		return []string{MethodAny}, arr[4], arr[6]
	}
	return strings.Split(methods, ","), arr[4], arr[6]
}

func ParserPattern(prefix string, pattern string) ([]*Route, error) {
	methods, path, metaStr := splitPattern(pattern)
	if path == "" {
		return nil, fmt.Errorf("invalid pattern %q", pattern)
	}

	path = CleanPath(prefix + path)
	rs := make([]*Route, 0, len(methods))

	meta, err := parserMeta(metaStr)
	if err != nil {
		return nil, fmt.Errorf("invalid meta in pattern %q, err: %w", pattern, err)
	}

	for _, method := range methods {
		rt := &Route{
			Method:      CleanMethod(method),
			Pattern:     path,
			PatternType: getPatternType(path),
			Meta:        meta,
		}

		// PatternWord 和 PatternRegexp 类型的通过 PathPrefix PathSuffix 加速匹配
		// PatternExact 的直接整体匹配即可，并不需要 PathPrefix PathSuffix

		if idx := strings.IndexAny(path, "{*"); idx >= 0 {
			rt.PathPrefix = path[:idx]
		}

		if idx := strings.LastIndexAny(path, "}*"); idx > 0 {
			rt.PathSuffix = path[idx+1:]
		}

		var err error
		switch rt.PatternType {
		case PatternWord:
			rt.wordNodes, err = parserWordNodes(path)
			for _, node := range rt.wordNodes {
				if node.Name != "" {
					rt.paramNum++
				}
			}
		case PatternRegexp:
			var str string
			str, err = parserRegexpPattern(path)
			rt.paramNum = strings.Count(str, "（?P<")
			if err == nil {
				rt.regexp, err = regexp.Compile("^" + str + "$")
			}
		default:
		}
		if err != nil {
			return nil, err
		}
		rs = append(rs, rt)
	}
	return rs, nil
}

func getPatternType(pattern string) PatternType {
	if strings.ContainsAny(pattern, ":*") {
		return PatternRegexp
	} else if strings.ContainsAny(pattern, "{") {
		return PatternWord
	} else {
		return PatternExact
	}
}

type Route struct {
	Method     string
	PathPrefix string // 地址前缀
	PathSuffix string // 地址后缀

	PatternType PatternType
	Pattern     string // 路由地址
	Handler     http.Handler

	Meta Meta // 其他元信息

	Info any // 其他信息，在注册的时候额外补充的，目前是 xhttp.RouteInfo

	// wordNodes PatternWord 类型的 Pattern 的节点
	// 个数 = Pattern 中 / 的个数
	wordNodes []*wordNode

	// regexp PatternRegexp 类型的的 Pattern
	regexp *regexp.Regexp

	// 路由变量个数
	paramNum int
}

// String 调试，打印使用
func (sr *Route) String() string {
	data := map[string]any{
		"Method":     sr.Method,
		"PathPrefix": sr.PathPrefix,
		"PathSuffix": sr.PathSuffix,
		"Pattern":    sr.Pattern,
		"Meta":       sr.Meta,
	}
	bf, _ := json.Marshal(data)
	return string(bf)
}

func (sr *Route) UniqKey() string {
	return sr.Method + "|" + sr.Pattern
}

func (sr *Route) LogFields() []xlog.Attr {
	fields := []xlog.Attr{
		xlog.String("Method", sr.Method),
		xlog.String("Pattern", sr.Pattern),
		xlog.Any("PatternType", sr.PatternType),
		xlog.String("PathPrefix", sr.PathPrefix),
		xlog.String("PathSuffix", sr.PathSuffix),
		xlog.Int("ParamNum", sr.paramNum),
		xlog.Any("Meta", sr.Meta),
	}
	switch sr.PatternType {
	case PatternWord:
		fields = append(fields, xlog.Any("PathNodes", sr.wordNodes))
	case PatternRegexp:
		fields = append(fields, xlog.String("PathRegexp", sr.regexp.String()))
	default:
	}
	return fields
}

var _ routeMatcher = (*Route)(nil)

type routeMatcher interface {
	Match(req *http.Request) (map[string]string, bool)
}

func (sr *Route) Match(req *http.Request) (map[string]string, bool) {
	ret, ok1 := sr.match(req)
	if !ok1 {
		return ret, false
	}
	if rr, ok2 := sr.Handler.(routeMatcher); ok2 {
		return rr.Match(req)
	}
	return ret, true
}

func (sr *Route) match(req *http.Request) (map[string]string, bool) {
	if sr.Method != MethodAny && sr.Method != req.Method {
		return nil, false
	}
	if sr.PatternType == PatternExact {
		return nil, sr.Pattern == req.URL.Path
	}
	if !strings.HasPrefix(req.URL.Path, sr.PathPrefix) || !strings.HasSuffix(req.URL.Path, sr.PathSuffix) {
		return nil, false
	}

	if sr.PatternType == PatternWord {
		arr := strings.Split(req.URL.Path, "/")
		if len(arr) != len(sr.wordNodes) {
			return nil, false
		}
		result := make(map[string]string, sr.paramNum)
		for i, node := range sr.wordNodes {
			val, ok := node.Match(arr[i])
			if !ok {
				return nil, false
			}
			if node.Name != "" {
				result[node.Name] = val
			}
		}
		return result, true
	}

	ms := sr.regexp.FindStringSubmatch(req.URL.Path)
	if len(ms) == 0 {
		return nil, false
	}
	names := sr.regexp.SubexpNames()
	result := make(map[string]string, sr.paramNum)
	for i, name := range names {
		if name != "" {
			result[name] = ms[i]
		}
	}
	return result, true
}

func (sr *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sr.Handler.ServeHTTP(w, req)
}

func parserWordNodes(pattern string) ([]*wordNode, error) {
	arr := strings.Split(pattern, "/")
	nodes := make([]*wordNode, len(arr))
	for i, str := range arr {
		node, err := parserWordNode(str)
		if err != nil {
			return nil, err
		}
		nodes[i] = node
	}
	return nodes, nil
}

func parserWordNode(str string) (*wordNode, error) {
	cnt1 := strings.Count(str, "{")
	cnt2 := strings.Count(str, "}")
	if cnt1 == 0 && cnt2 == 0 {
		return &wordNode{
			Prefix: str,
		}, nil
	}

	if cnt1 != 1 || cnt2 != 1 {
		return nil, fmt.Errorf("invalid path %q", str)
	}
	prefix, after, ok1 := strings.Cut(str, "{")
	if !ok1 {
		return nil, fmt.Errorf("not found ‘{’ in %q", str)
	}
	name, suffix, ok2 := strings.Cut(after, "}")
	if !ok2 {
		return nil, fmt.Errorf("not found '}' in %q", after)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("invalid path %q, should like '{name}'", str)
	}
	node := &wordNode{
		Prefix: prefix,
		Suffix: suffix,
		Name:   name,
	}
	return node, nil
}

type wordNode struct {
	Prefix string
	Suffix string
	Name   string // 变量名,若为空，则不是变量
}

func (n *wordNode) Match(str string) (string, bool) {
	if !strings.HasPrefix(str, n.Prefix) || !strings.HasSuffix(str, n.Suffix) {
		return "", false
	}
	if n.Name == "" {
		return "", true
	}
	result := str[len(n.Prefix) : len(str)-len(n.Suffix)]
	return result, true
}

const uuidReg = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

// 解析正则路由地址：
//
//		正则表达式：
//		/user/{category}/{id:[0-9]+}, /user/{id:[0-9]+}.html, /user/hello-{id:[0-9]+}-{age:[0-9]+}.html
//
//		 /{id:\d{2}-\d{3}-\d{4}}.{ext:\d{3}}
//
//		*通配符（简化正则）(* 可以匹配包含 / 的所有字符):
//		/user/*,  /user/*/detail, /user/*/detail/*, /user/*/detail/*.html
//	 /user/{s1:*},  /user/{s1:*}/detail,  /user/{s1:*}/detail/{s2:*}
func parserRegexpPattern(pattern string) (string, error) {
	cnt1 := strings.Count(pattern, "{")
	cnt2 := strings.Count(pattern, "}")
	if cnt1 != cnt2 {
		return "", fmt.Errorf("invalid path %q, has %d '{' and %d '}'", pattern, cnt1, cnt2)
	}
	names := make(map[string]bool)
	var regPatternNew string // 归一化后的正则地址
	for {
		leftIndex, rightIndex, found := xstr.BytePairIndex(pattern, '{', '}')
		if !found && (rightIndex != -1 || leftIndex != -1) {
			return "", fmt.Errorf("invalid path %q", pattern)
		}
		starIndex := strings.IndexByte(pattern, '*')

		// 先处理 /* 这种直接使用 * 的路由地址，如 /*/{id}
		if starIndex != -1 && (starIndex < leftIndex || !found) {
			name := fmt.Sprintf("p%d", len(names))
			if names[name] {
				return "", fmt.Errorf("dup name %q in path %q", name, pattern)
			}
			names[name] = true
			regPatternNew += regexp.QuoteMeta(pattern[:starIndex]) + fmt.Sprintf("(?P<%s>%s)", name, ".*")
			pattern = pattern[starIndex+1:]
			continue
		}

		if !found {
			regPatternNew += regexp.QuoteMeta(pattern)
			break
		}

		// 找到变量参数，的到如 category、id:[0-9]+、id:\d{2}-\d{3}-\d{4}
		varTxt := pattern[leftIndex+1 : rightIndex]

		name, reg, ok := strings.Cut(varTxt, ":")
		if names[name] {
			return "", fmt.Errorf("dup name %q in path %q", name, pattern)
		}
		regPatternNew += regexp.QuoteMeta(pattern[:leftIndex])
		names[name] = true
		if ok { // {id:[0-9]+} 、{id:*}
			if txt, ok1 := regexpAlias[reg]; ok1 {
				reg = txt
			} else {
				switch reg {
				case "*":
					reg = ".*"
				case "UUID":
					reg = uuidReg
				case "Base62":
					reg = `[0-9a-zA-Z]+`
				case "Base36":
					reg = `[0-9a-z]+`
				case "Base58":
					reg = `[1-9A-HJ-NP-Za-km-z]+`
				case "Base64URL":
					reg = `[0-9a-zA-Z\-_]+`
				case "UINT":
					reg = `0|[1-9][0-9]*`
				case "INT":
					reg = `[-]?(0|[1-9][0-9]*)`
				}
			}
			regPatternNew += fmt.Sprintf("(?P<%s>%s)", name, reg)
		} else { // {id}
			regPatternNew += fmt.Sprintf("(?P<%s>%s)", name, "[^/]+")
		}
		pattern = pattern[rightIndex+1:]
	}
	return regPatternNew, nil
}

var regexpAlias = map[string]string{}

func RegisterRegexpAlias(name string, reg string) {
	if name == "" || reg == "" {
		panic(fmt.Sprintf("invalid param: RegisterRegexpAlias(%q,%q)", name, reg))
	}
	_ = regexp.MustCompile(reg)
	regexpAlias[name] = reg
}
