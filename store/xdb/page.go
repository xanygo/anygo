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
	Total int // 总结果条数
	Page  int // 当前页数,总是 > 0, 若是第一页，则值为 1
	Size  int // 每一页的结果数
}

// TotalPages 总页数，返回结果总是 >= 1
func (p Pagination) TotalPages() int {
	if p.Total <= p.Size {
		return 1
	}
	page := math.Ceil(float64(p.Total) / float64(p.Size))
	return int(page)
}

// NearPages 输入当前页码，得到有效的前后 num 页的起止页码
func (p Pagination) NearPages(num int) (start int, end int) {
	start = p.Page - num
	end = p.Page + num
	for start < 1 {
		start++
		end++
	}
	totalPage := p.TotalPages()
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
		Page:  page,
		Total: int(total),
		Size:  size,
	}
	items := make([]Record[T], len(datas))
	for idx, value := range datas {
		items[idx] = Record[T]{
			Value: value,
			Order: idx,
			Index: (page-1)*size + idx,
			Ext:   map[string]any{},
		}
	}
	return pageInfo, items, nil
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
