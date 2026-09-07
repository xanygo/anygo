//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-04-15

package xslice

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

type filterCond struct {
	Tag string
	Not bool
}

type compiledFilter struct {
	Groups [][]filterCond
	HasAny bool
	HasAll bool
}

func parseFilter(filter string) (compiledFilter, error) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return compiledFilter{
			HasAll: true,
		}, nil
	}

	result := compiledFilter{}

	orParts := strings.Split(filter, ",")
	for index, part := range orParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch part {
		case "[ANY]":
			if index != len(orParts)-1 {
				return compiledFilter{}, fmt.Errorf("invalid filter %q: [any] must be the last condition", filter)
			}
			result.HasAny = true
			continue
		case "[ALL]":
			if index != len(orParts)-1 {
				return compiledFilter{}, fmt.Errorf("invalid filter %q: [any] must be the last condition", filter)
			}
			result.HasAll = true
			continue
		default:
		}

		andParts := strings.Split(part, "&")
		var group []filterCond

		for _, tag := range andParts {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}

			cond := filterCond{}

			if strings.HasPrefix(tag, "!") {
				cond.Not = true
				cond.Tag = strings.TrimSpace(tag[1:])
			} else {
				cond.Tag = tag
			}

			if cond.Tag != "" {
				group = append(group, cond)
			}
		}

		if len(group) > 0 {
			result.Groups = append(result.Groups, group)
		}
	}

	return result, nil
}

// BuildTagFirst  构建一个基于 tag 表达式的切片过滤函数，过滤并返回一个第一个满足条件的结果
// filter 的参数详见 BuildTagFilter
func BuildTagFirst[T any](filter string, getTags func(T) []string) (func([]T) T, error) {
	bf, err := BuildTagFilter[T](filter, getTags, 1)
	if err != nil {
		return nil, err
	}
	return func(items []T) (v T) {
		values := bf(items)
		if len(values) == 1 {
			return values[0]
		}
		return v
	}, nil
}

// BuildTagFilter 构建一个基于 tag 表达式的切片过滤函数（支持多结果返回）。
//
// 该函数会解析 filter 表达式，并返回一个可复用的过滤函数。
// 过滤函数用于从切片中筛选满足条件的元素，并返回一个新的切片。
//
// 支持的表达式语法：
//
//  1. OR（或关系）： 使用 "," 分隔，例如："tag1,tag2" 表示：匹配 tag1 或 tag2 的元素
//
//  2. AND（与关系）： 使用 "&" 连接，例如："tag1 & tag2" 表示：必须同时包含 tag1 和 tag2
//
//  3. NOT（非关系）： 使用 "!" 前缀，例如： "!tag1"，表示：不包含 tag1
//
//  4. [ANY]（兜底规则）: 特殊关键字 "[ANY]" 必须作为最后一个 OR 条件使用，表示：当所有条件均不匹配时，从原始切片中返回任意一个元素
//
//  5. [ALL]（兜底规则）: 特殊关键字 "[ALL]" 必须作为最后一个 OR 条件使用，表示：当所有条件均不匹配时，返回原始切片
//
// 示例：
//
//  1. "tag1 & tag2, tag3, [ANY]"  => 语义：①优先匹配 tag1 & tag2 ② 否则匹配 tag3 ③如果都不匹配，则返回任意一个元素
//  2. "tag1 , tag2, [ALL]"  => 语义：①优先匹配 tag1 ② 否则匹配 tag2 ③如果都不匹配，则返回原始切片
//
// 参数说明：
//
//	filter: tag 过滤表达式字符串
//	getTags:  用于从类型 T 中提取 tag 列表的函数，示例：func(v T) []string
//	limit:返回结果数量上限。 limit > 0：最多返回 limit 个匹配结果； limit <= 0：返回所有匹配结果
//
// 示例：
//
//	fn, _ := BuildTagFilter[[]Item](
//	    "tag1 & tag2, tag3, any",
//	    func(v Item) []string { return v.Tags },
//	    10,
//	)
//
//	result := fn(items)
func BuildTagFilter[T any](filter string, getTags func(T) []string, limit int) (func([]T) []T, error) {
	cf, err := parseFilter(filter)
	if err != nil {
		return nil, err
	}

	return func(items []T) []T {
		var result []T

		for _, item := range items {
			tags := getTags(item)

			tagSet := make(map[string]struct{}, len(tags))
			for _, t := range tags {
				tagSet[t] = struct{}{}
			}

			matchItem := false

			for _, group := range cf.Groups {
				match := true

				for _, cond := range group {
					_, has := tagSet[cond.Tag]

					if cond.Not {
						if has {
							match = false
							break
						}
					} else {
						if !has {
							match = false
							break
						}
					}
				}

				if match {
					matchItem = true
					break
				}
			}

			if matchItem {
				result = append(result, item)
				if limit > 0 && len(result) >= limit {
					return result
				}
			}
		}

		if len(result) == 0 && cf.HasAny && len(items) > 0 {
			n := rand.IntN(len(items))
			return []T{items[n]}
		} else if len(result) == 0 && cf.HasAll {
			if limit > 0 && len(items) > limit {
				return result[:limit]
			}
			return items
		}
		return result
	}, nil
}
