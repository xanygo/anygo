//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-20

package xdb

import (
	"context"
	"math"
)

// Pagination 分页信息
type Pagination struct {
	TotalRecords int // 总结果条数
	PageIndex    int // 当前页数,总是 > 0, 若是第一页，则值为 1
	PageCount    int // 总页数,可选，当 >0 时有效。否则依据 TotalRecords 计数而来
	PageSize     int // 当前页最大结果数
}

// GetPageCount 总页数，返回结果总是 >= 1
func (p Pagination) GetPageCount() int {
	if p.PageCount > 0 {
		return p.PageCount
	}
	if p.TotalRecords <= p.PageSize {
		return 1
	}
	page := math.Ceil(float64(p.TotalRecords) / float64(p.PageSize))
	return int(page)
}

// NearPages 输入当前页码，得到有效的前后 num 页的起止页码
func (p Pagination) NearPages(num int) (start int, end int) {
	start = p.PageIndex - num
	end = p.PageIndex + num
	for start < 1 {
		start++
		end++
	}
	totalPage := p.GetPageCount()
	for end > totalPage {
		end--
		start--
	}
	if start < 1 {
		start = 1
	}
	if end > totalPage {
		end = totalPage
	}
	return start, end
}

// PageList 分页查询
//
// page: 当前页码，总是 >=1
// size: 查询结果数，总是 >=1
func PageList[T any](ctx context.Context, b Builder, page int, size int, search CountSearch[T]) (Pagination, []Record[T], error) {
	if b == nil {
		b = EmptyBuilder()
	}
	total, datas, err := search(ctx, b, page, size)
	if err != nil {
		return Pagination{}, nil, err
	}
	pageInfo := Pagination{
		PageIndex:    page,
		TotalRecords: int(total),
		PageSize:     size,
	}
	items := toPageRecord[T](datas, page, size)
	return pageInfo, items, nil
}

func toPageRecord[T any](datas []T, page int, size int) []Record[T] {
	items := make([]Record[T], len(datas))
	for idx, value := range datas {
		items[idx] = Record[T]{
			Value: value,
			Order: idx,
			Index: (page-1)*size + idx,
			Ext:   map[string]any{},
		}
	}
	return items
}

// CountSearch 统计结果数以及分页的结果集
type CountSearch[T any] func(ctx context.Context, b Builder, page int, size int) (int64, []T, error)

type Record[T any] struct {
	Value T
	Order int // 当前页面索引
	Index int // 在所有页面的索引
	Ext   map[string]any
}

func (r Record[T]) HumanIndex() int {
	return r.Index + 1
}
