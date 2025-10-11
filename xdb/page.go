//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-11-20

package xdb

import (
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
