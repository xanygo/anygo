//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-10-20

package xhtml

// Table1 一个简单的表格
type Table1 struct {
	WithAttrs
	head []Element
	rows [][]Element
	foot []Element
}

// SetHeader 设置表头
func (t *Table1) SetHeader(cells ...Element) {
	t.head = cells
}

// AddRow 添加一行内容
func (t *Table1) AddRow(cells ...Element) {
	t.rows = append(t.rows, cells)
}

// AddRows 添加多行内容
func (t *Table1) AddRows(rows ...[]Element) {
	t.rows = append(t.rows, rows...)
}

// SetFooter 设置表格的页脚
func (t *Table1) SetFooter(cells ...Element) {
	t.foot = cells
}

// HTML 实现 Element 接口
func (t *Table1) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<table")
	bw.WriteWithSep(" ", t.Attrs)
	bw.Write(">\n")
	bw.Write("<thead>\n<tr>")
	for i := 0; i < len(t.head); i++ {
		bw.Write(t.head[i])
	}
	bw.Write("</tr>\n</thead>\n<tbody>\n")
	for i := 0; i < len(t.rows); i++ {
		row := t.rows[i]
		bw.Write("<tr>")
		for j := 0; j < len(row); j++ {
			bw.Write(row[j])
		}
		bw.Write("</tr>\n")
	}
	bw.Write("</tbody>\n")
	if len(t.foot) > 0 {
		bw.Write("<tfoot>\n<tr>")
		for i := 0; i < len(t.foot); i++ {
			bw.Write(t.foot[i])
		}
		bw.Write("</tr>\n</tfoot>\n")
	}
	bw.Write("</table>\n")
	return bw.HTML()
}

// NewTd 创建一个新的 td
func NewTd(val ...Element) *Any {
	return &Any{
		Tag:  "td",
		Body: val,
	}
}

// NewTh 创建一个新的 th
func NewTh(val ...Element) *Any {
	return &Any{
		Tag:  "th",
		Body: val,
	}
}
