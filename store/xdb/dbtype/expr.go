package dbtype

// Expr 包含文本 SQL 内容和 参数的表达式
type Expr struct {
	SQL  string
	Args []any
}
