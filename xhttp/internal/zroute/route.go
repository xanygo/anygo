//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-11

package zroute

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/xanygo/anygo/xlog"
)

const MethodAny = "ANY"

type PatternType int8

const (
	PatternExact PatternType = iota
	PatternWord
	PatternRegexp
)

var patternReg = regexp.MustCompile(`^(([A-Za-z]+(,[A-Za-z]+)*\s+)?)(/\S*)$`)

// splitPattern 解析 pattern 中的 Method 和 Path值
func splitPattern(pattern string) ([]string, string) {
	arr := patternReg.FindStringSubmatch(pattern)
	if len(arr) == 0 {
		return nil, ""
	}
	methods := strings.TrimSpace(arr[1])
	if methods == "" {
		return []string{MethodAny}, arr[4]
	}
	return strings.Split(methods, ","), arr[4]
}

func ParserPattern(prefix string, pattern string) ([]*Route, error) {
	methods, path := splitPattern(pattern)
	if path == "" {
		return nil, fmt.Errorf("invalid pattern %q", pattern)
	}
	path = CleanPath(prefix + path)
	rs := make([]*Route, 0, len(methods))
	for _, method := range methods {
		rt := &Route{
			Method:      CleanMethod(method),
			Pattern:     path,
			PatternType: getPatternType(path),
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

	Info any // 其他信息，在注册的时候额外补充的，目前是 xhttp.RouteInfo

	// wordNodes PatternWord 类型的 Pattern 的节点
	// 个数 = Pattern 中 / 的个数
	wordNodes []*wordNode

	// regexp PatternRegexp 类型的的 Pattern
	regexp *regexp.Regexp

	// 路由变量个数
	paramNum int
}

func (sr *Route) LogFields() []xlog.Attr {
	fields := []xlog.Attr{
		xlog.String("Method", sr.Method),
		xlog.String("Pattern", sr.Pattern),
		xlog.Any("PatternType", sr.PatternType),
		xlog.String("PathPrefix", sr.PathPrefix),
		xlog.String("PathSuffix", sr.PathSuffix),
		xlog.Int("ParamNum", sr.paramNum),
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

func (sr *Route) Match(req *http.Request) (map[string]string, bool) {
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

// 解析正则路由地址：
//
//	正则表达式：
//	/user/{category}/{id:[0-9]+}, /user/{id:[0-9]+}.html, /user/hello-{id:[0-9]+}-{age:[0-9]+}.html
//
//	*通配符（简化正则）(* 可以匹配包含 / 的所有字符):
//	   /user/*,  /user/*/detail, /user/*/detail/*, /user/*/detail/*.html
//	   /user/{s1:*},  /user/{s1:*}/detail,  /user/{s1:*}/detail/{s2:*}
func parserRegexpPattern(pattern string) (string, error) {
	cnt1 := strings.Count(pattern, "{")
	cnt2 := strings.Count(pattern, "}")
	if cnt1 != cnt2 {
		return "", fmt.Errorf("invalid path %q, has %d '{' and %d '}'", pattern, cnt1, cnt2)
	}
	names := make(map[string]bool)
	var regPatternNew string // 归一化后的正则地址
	for {
		leftIndex := strings.IndexByte(pattern, '{')
		rightIndex := strings.IndexByte(pattern, '}')
		if leftIndex > rightIndex {
			return "", fmt.Errorf("invalid path %q", pattern)
		}
		starIndex := strings.IndexByte(pattern, '*')

		// 先处理 /* 这种之间使用 * 的路由地址
		if starIndex != -1 && (starIndex < leftIndex || leftIndex == -1) {
			name := fmt.Sprintf("p%d", len(names))
			if names[name] {
				return "", fmt.Errorf("dup name %q in path %q", name, pattern)
			}
			names[name] = true
			regPatternNew += regexp.QuoteMeta(pattern[:starIndex]) + fmt.Sprintf("(?P<%s>%s)", name, ".*")
			pattern = pattern[starIndex+1:]
			continue
		}

		if rightIndex != -1 {
			str := pattern[leftIndex+1 : rightIndex]
			name, reg, ok := strings.Cut(str, ":")
			if names[name] {
				return "", fmt.Errorf("dup name %q in path %q", name, pattern)
			}
			regPatternNew += pattern[:leftIndex]
			names[name] = true
			if ok { // {id:[0-9]+} 、{id:*}
				if reg == "*" {
					reg = ".*"
				}
				regPatternNew += fmt.Sprintf("(?P<%s>%s)", name, reg)
			} else { // {id}
				regPatternNew += fmt.Sprintf("(?P<%s>%s)", name, ".*")
			}
			pattern = pattern[rightIndex+1:]
			continue
		}
		regPatternNew += regexp.QuoteMeta(pattern)
		break
	}
	return regPatternNew, nil
}
