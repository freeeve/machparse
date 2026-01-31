package ast

import (
	"reflect"
	"sync"
)

// isNil checks if a Node interface contains nil.
func isNil(n Node) bool {
	if n == nil {
		return true
	}
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// Node pools for reducing allocations during parsing.
// Use Get* functions to obtain nodes and Release* to return them.

// Slice pools for common slice types
var (
	selectExprSlicePool = sync.Pool{
		New: func() any {
			s := make([]SelectExpr, 0, 8)
			return &s
		},
	}
	exprSlicePool = sync.Pool{
		New: func() any {
			s := make([]Expr, 0, 4)
			return &s
		},
	}
	orderBySlicePool = sync.Pool{
		New: func() any {
			s := make([]*OrderByExpr, 0, 4)
			return &s
		},
	}
)

// GetSelectExprSlice returns a []SelectExpr from the pool.
func GetSelectExprSlice() *[]SelectExpr {
	return selectExprSlicePool.Get().(*[]SelectExpr)
}

// ReleaseSelectExprSlice returns a []SelectExpr to the pool.
func ReleaseSelectExprSlice(s *[]SelectExpr) {
	*s = (*s)[:0]
	selectExprSlicePool.Put(s)
}

// GetExprSlice returns a []Expr from the pool.
func GetExprSlice() *[]Expr {
	return exprSlicePool.Get().(*[]Expr)
}

// ReleaseExprSlice returns a []Expr to the pool.
func ReleaseExprSlice(s *[]Expr) {
	*s = (*s)[:0]
	exprSlicePool.Put(s)
}

// GetOrderBySlice returns a []*OrderByExpr from the pool.
func GetOrderBySlice() *[]*OrderByExpr {
	return orderBySlicePool.Get().(*[]*OrderByExpr)
}

// ReleaseOrderBySlice returns a []*OrderByExpr to the pool.
func ReleaseOrderBySlice(s *[]*OrderByExpr) {
	*s = (*s)[:0]
	orderBySlicePool.Put(s)
}

// Node pools for reducing allocations during parsing.
var (
	colNamePool = sync.Pool{
		New: func() any { return &ColName{} },
	}
	literalPool = sync.Pool{
		New: func() any { return &Literal{} },
	}
	binaryExprPool = sync.Pool{
		New: func() any { return &BinaryExpr{} },
	}
	funcExprPool = sync.Pool{
		New: func() any { return &FuncExpr{} },
	}
	aliasedExprPool = sync.Pool{
		New: func() any { return &AliasedExpr{} },
	}
	selectStmtPool = sync.Pool{
		New: func() any { return &SelectStmt{} },
	}
	tableNamePool = sync.Pool{
		New: func() any { return &TableName{} },
	}
	orderByExprPool = sync.Pool{
		New: func() any { return &OrderByExpr{} },
	}
	aliasedTableExprPool = sync.Pool{
		New: func() any { return &AliasedTableExpr{} },
	}
	joinExprPool = sync.Pool{
		New: func() any { return &JoinExpr{} },
	}
	unaryExprPool = sync.Pool{
		New: func() any { return &UnaryExpr{} },
	}
)

// GetColName returns a ColName from the pool.
func GetColName() *ColName {
	return colNamePool.Get().(*ColName)
}

// ReleaseColName returns a ColName to the pool.
func ReleaseColName(c *ColName) {
	*c = ColName{} // reset
	colNamePool.Put(c)
}

// GetLiteral returns a Literal from the pool.
func GetLiteral() *Literal {
	return literalPool.Get().(*Literal)
}

// ReleaseLiteral returns a Literal to the pool.
func ReleaseLiteral(l *Literal) {
	*l = Literal{} // reset
	literalPool.Put(l)
}

// GetBinaryExpr returns a BinaryExpr from the pool.
func GetBinaryExpr() *BinaryExpr {
	return binaryExprPool.Get().(*BinaryExpr)
}

// ReleaseBinaryExpr returns a BinaryExpr to the pool.
func ReleaseBinaryExpr(b *BinaryExpr) {
	*b = BinaryExpr{} // reset
	binaryExprPool.Put(b)
}

// GetFuncExpr returns a FuncExpr from the pool.
func GetFuncExpr() *FuncExpr {
	return funcExprPool.Get().(*FuncExpr)
}

// ReleaseFuncExpr returns a FuncExpr to the pool.
func ReleaseFuncExpr(f *FuncExpr) {
	*f = FuncExpr{} // reset
	funcExprPool.Put(f)
}

// GetAliasedExpr returns an AliasedExpr from the pool.
func GetAliasedExpr() *AliasedExpr {
	return aliasedExprPool.Get().(*AliasedExpr)
}

// ReleaseAliasedExpr returns an AliasedExpr to the pool.
func ReleaseAliasedExpr(a *AliasedExpr) {
	*a = AliasedExpr{} // reset
	aliasedExprPool.Put(a)
}

// GetSelectStmt returns a SelectStmt from the pool.
func GetSelectStmt() *SelectStmt {
	return selectStmtPool.Get().(*SelectStmt)
}

// ReleaseSelectStmt returns a SelectStmt to the pool.
func ReleaseSelectStmt(s *SelectStmt) {
	*s = SelectStmt{} // reset
	selectStmtPool.Put(s)
}

// GetTableName returns a TableName from the pool.
func GetTableName() *TableName {
	return tableNamePool.Get().(*TableName)
}

// ReleaseTableName returns a TableName to the pool.
func ReleaseTableName(t *TableName) {
	*t = TableName{} // reset
	tableNamePool.Put(t)
}

// GetOrderByExpr returns an OrderByExpr from the pool.
func GetOrderByExpr() *OrderByExpr {
	return orderByExprPool.Get().(*OrderByExpr)
}

// ReleaseOrderByExpr returns an OrderByExpr to the pool.
func ReleaseOrderByExpr(o *OrderByExpr) {
	*o = OrderByExpr{} // reset
	orderByExprPool.Put(o)
}

// GetAliasedTableExpr returns an AliasedTableExpr from the pool.
func GetAliasedTableExpr() *AliasedTableExpr {
	return aliasedTableExprPool.Get().(*AliasedTableExpr)
}

// ReleaseAliasedTableExpr returns an AliasedTableExpr to the pool.
func ReleaseAliasedTableExpr(a *AliasedTableExpr) {
	*a = AliasedTableExpr{} // reset
	aliasedTableExprPool.Put(a)
}

// GetJoinExpr returns a JoinExpr from the pool.
func GetJoinExpr() *JoinExpr {
	return joinExprPool.Get().(*JoinExpr)
}

// ReleaseJoinExpr returns a JoinExpr to the pool.
func ReleaseJoinExpr(j *JoinExpr) {
	*j = JoinExpr{} // reset
	joinExprPool.Put(j)
}

// GetUnaryExpr returns a UnaryExpr from the pool.
func GetUnaryExpr() *UnaryExpr {
	return unaryExprPool.Get().(*UnaryExpr)
}

// ReleaseUnaryExpr returns a UnaryExpr to the pool.
func ReleaseUnaryExpr(u *UnaryExpr) {
	*u = UnaryExpr{} // reset
	unaryExprPool.Put(u)
}

// ReleaseAST recursively releases all pooled nodes in an AST.
// Call this when done with a parsed statement to return nodes to pools.
func ReleaseAST(node Node) {
	if isNil(node) {
		return
	}

	switch n := node.(type) {
	case *SelectStmt:
		for _, col := range n.Columns {
			ReleaseAST(col)
		}
		// Release Columns slice to pool
		if cap(n.Columns) > 0 {
			cols := n.Columns[:0]
			ReleaseSelectExprSlice(&cols)
		}
		ReleaseAST(n.From)
		ReleaseAST(n.Where)
		for _, expr := range n.GroupBy {
			ReleaseAST(expr)
		}
		// Release GroupBy slice to pool
		if cap(n.GroupBy) > 0 {
			groupBy := n.GroupBy[:0]
			ReleaseExprSlice(&groupBy)
		}
		ReleaseAST(n.Having)
		for _, ob := range n.OrderBy {
			ReleaseAST(ob.Expr)
			ReleaseOrderByExpr(ob)
		}
		// Release OrderBy slice to pool
		if cap(n.OrderBy) > 0 {
			orderBy := n.OrderBy[:0]
			ReleaseOrderBySlice(&orderBy)
		}
		if n.Limit != nil {
			ReleaseAST(n.Limit.Count)
			ReleaseAST(n.Limit.Offset)
		}
		ReleaseSelectStmt(n)

	case *ColName:
		ReleaseColName(n)

	case *Literal:
		ReleaseLiteral(n)

	case *BinaryExpr:
		ReleaseAST(n.Left)
		ReleaseAST(n.Right)
		ReleaseBinaryExpr(n)

	case *FuncExpr:
		for _, arg := range n.Args {
			ReleaseAST(arg)
		}
		// Release Args slice to pool
		if cap(n.Args) > 0 {
			args := n.Args[:0]
			ReleaseExprSlice(&args)
		}
		ReleaseAST(n.Filter)
		ReleaseFuncExpr(n)

	case *AliasedExpr:
		ReleaseAST(n.Expr)
		ReleaseAliasedExpr(n)

	case *TableName:
		ReleaseTableName(n)

	case *AliasedTableExpr:
		ReleaseAST(n.Expr)
		ReleaseAliasedTableExpr(n)

	case *JoinExpr:
		ReleaseAST(n.Left)
		ReleaseAST(n.Right)
		ReleaseAST(n.On)
		ReleaseJoinExpr(n)

	case *UnaryExpr:
		ReleaseAST(n.Operand)
		ReleaseUnaryExpr(n)

	case *ParenExpr:
		ReleaseAST(n.Expr)

	case *Subquery:
		ReleaseAST(n.Select)

	case *InExpr:
		ReleaseAST(n.Expr)
		for _, v := range n.Values {
			ReleaseAST(v)
		}
		ReleaseAST(n.Select)

	case *BetweenExpr:
		ReleaseAST(n.Expr)
		ReleaseAST(n.Low)
		ReleaseAST(n.High)

	case *CaseExpr:
		ReleaseAST(n.Operand)
		for _, w := range n.Whens {
			ReleaseAST(w.Cond)
			ReleaseAST(w.Result)
		}
		ReleaseAST(n.Else)

	case *CastExpr:
		ReleaseAST(n.Expr)
	}
}
