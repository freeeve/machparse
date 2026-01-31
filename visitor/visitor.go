// Package visitor provides AST traversal and rewriting utilities.
package visitor

import "github.com/freeeve/machparse/ast"

// Visitor is the interface for AST traversal.
type Visitor interface {
	Visit(node ast.Node) Visitor
}

// Walk traverses an AST in depth-first order.
func Walk(v Visitor, node ast.Node) {
	if node == nil {
		return
	}
	if v = v.Visit(node); v == nil {
		return
	}
	walkChildren(v, node)
}

func walkChildren(v Visitor, node ast.Node) {
	switch n := node.(type) {
	case *ast.SelectStmt:
		if n.With != nil {
			for _, cte := range n.With.CTEs {
				Walk(v, cte.Query)
			}
		}
		for _, col := range n.Columns {
			Walk(v, col)
		}
		if n.From != nil {
			Walk(v, n.From)
		}
		if n.Where != nil {
			Walk(v, n.Where)
		}
		for _, expr := range n.GroupBy {
			Walk(v, expr)
		}
		if n.Having != nil {
			Walk(v, n.Having)
		}
		for _, ob := range n.OrderBy {
			Walk(v, ob.Expr)
		}
		if n.Limit != nil {
			if n.Limit.Count != nil {
				Walk(v, n.Limit.Count)
			}
			if n.Limit.Offset != nil {
				Walk(v, n.Limit.Offset)
			}
		}

	case *ast.InsertStmt:
		Walk(v, n.Table)
		for _, col := range n.Columns {
			Walk(v, col)
		}
		for _, row := range n.Values {
			for _, val := range row {
				Walk(v, val)
			}
		}
		if n.Select != nil {
			Walk(v, n.Select)
		}
		for _, ue := range n.OnDuplicateUpdate {
			Walk(v, ue.Expr)
		}
		for _, se := range n.Returning {
			Walk(v, se)
		}

	case *ast.UpdateStmt:
		Walk(v, n.Table)
		for _, ue := range n.Set {
			Walk(v, ue.Column)
			Walk(v, ue.Expr)
		}
		if n.From != nil {
			Walk(v, n.From)
		}
		if n.Where != nil {
			Walk(v, n.Where)
		}
		for _, se := range n.Returning {
			Walk(v, se)
		}

	case *ast.DeleteStmt:
		Walk(v, n.Table)
		if n.Using != nil {
			Walk(v, n.Using)
		}
		if n.Where != nil {
			Walk(v, n.Where)
		}
		for _, se := range n.Returning {
			Walk(v, se)
		}

	case *ast.BinaryExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *ast.UnaryExpr:
		Walk(v, n.Operand)

	case *ast.ParenExpr:
		Walk(v, n.Expr)

	case *ast.FuncExpr:
		for _, arg := range n.Args {
			Walk(v, arg)
		}
		if n.Filter != nil {
			Walk(v, n.Filter)
		}
		if n.Over != nil {
			for _, pb := range n.Over.PartitionBy {
				Walk(v, pb)
			}
			for _, ob := range n.Over.OrderBy {
				Walk(v, ob.Expr)
			}
		}

	case *ast.CaseExpr:
		if n.Operand != nil {
			Walk(v, n.Operand)
		}
		for _, w := range n.Whens {
			Walk(v, w.Cond)
			Walk(v, w.Result)
		}
		if n.Else != nil {
			Walk(v, n.Else)
		}

	case *ast.InExpr:
		Walk(v, n.Expr)
		for _, val := range n.Values {
			Walk(v, val)
		}
		if n.Select != nil {
			Walk(v, n.Select)
		}

	case *ast.BetweenExpr:
		Walk(v, n.Expr)
		Walk(v, n.Low)
		Walk(v, n.High)

	case *ast.LikeExpr:
		Walk(v, n.Expr)
		Walk(v, n.Pattern)
		if n.Escape != nil {
			Walk(v, n.Escape)
		}

	case *ast.IsExpr:
		Walk(v, n.Expr)

	case *ast.CastExpr:
		Walk(v, n.Expr)

	case *ast.Subquery:
		Walk(v, n.Select)

	case *ast.ExistsExpr:
		Walk(v, n.Subquery)

	case *ast.ColName:
		// Parts are strings, not AST nodes - nothing to walk

	case *ast.AliasedExpr:
		Walk(v, n.Expr)

	case *ast.AliasedTableExpr:
		Walk(v, n.Expr)

	case *ast.JoinExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)
		if n.On != nil {
			Walk(v, n.On)
		}

	case *ast.ParenTableExpr:
		Walk(v, n.Expr)

	case *ast.IntervalExpr:
		Walk(v, n.Value)

	case *ast.ExtractExpr:
		Walk(v, n.Source)

	case *ast.TrimExpr:
		if n.TrimChar != nil {
			Walk(v, n.TrimChar)
		}
		Walk(v, n.Expr)

	case *ast.SubstringExpr:
		Walk(v, n.Expr)
		if n.From != nil {
			Walk(v, n.From)
		}
		if n.For != nil {
			Walk(v, n.For)
		}

	case *ast.PositionExpr:
		Walk(v, n.Needle)
		Walk(v, n.Haystack)

	case *ast.ArrayExpr:
		for _, elem := range n.Elements {
			Walk(v, elem)
		}

	case *ast.SubscriptExpr:
		Walk(v, n.Expr)
		Walk(v, n.Index)

	case *ast.CollateExpr:
		Walk(v, n.Expr)

	case *ast.CreateTableStmt:
		Walk(v, n.Table)
		if n.As != nil {
			Walk(v, n.As)
		}
		for _, col := range n.Columns {
			for _, cons := range col.Constraints {
				if cons.Default != nil {
					Walk(v, cons.Default)
				}
				if cons.Check != nil {
					Walk(v, cons.Check)
				}
			}
		}

	case *ast.AlterTableStmt:
		Walk(v, n.Table)

	case *ast.DropTableStmt:
		for _, t := range n.Tables {
			Walk(v, t)
		}

	case *ast.CreateIndexStmt:
		Walk(v, n.Table)
		if n.Where != nil {
			Walk(v, n.Where)
		}

	case *ast.ExplainStmt:
		Walk(v, n.Stmt)

	case *ast.SetOp:
		Walk(v, n.Left)
		Walk(v, n.Right)

	case *ast.ValuesStmt:
		for _, row := range n.Rows {
			for _, val := range row {
				Walk(v, val)
			}
		}
	}
}

// WalkFunc is a convenience wrapper that calls a function for each node.
func WalkFunc(node ast.Node, fn func(ast.Node) bool) {
	Walk(&funcVisitor{fn: fn}, node)
}

type funcVisitor struct {
	fn func(ast.Node) bool
}

func (v *funcVisitor) Visit(node ast.Node) Visitor {
	if v.fn(node) {
		return v
	}
	return nil
}

// Inspect calls f for each node in the AST.
// If f returns false, children are not visited.
func Inspect(node ast.Node, f func(ast.Node) bool) {
	WalkFunc(node, f)
}
