package visitor

import "github.com/freeeve/machparse/ast"

// ApplyFunc is called for each node during rewriting.
// Return the replacement node or the original to keep it.
type ApplyFunc func(ast.Node) ast.Node

// Rewrite traverses the AST and allows modifying nodes.
// The function is called in post-order (children first, then parent).
func Rewrite(node ast.Node, f ApplyFunc) ast.Node {
	if node == nil {
		return nil
	}

	// First, recursively rewrite children
	rewriteChildren(node, f)

	// Then apply the function to this node
	return f(node)
}

func rewriteChildren(node ast.Node, f ApplyFunc) {
	switch n := node.(type) {
	case *ast.SelectStmt:
		if n.With != nil {
			for i, cte := range n.With.CTEs {
				if result := Rewrite(cte.Query, f); result != nil {
					n.With.CTEs[i].Query = result.(ast.Statement)
				}
			}
		}
		for i, col := range n.Columns {
			if result := Rewrite(col, f); result != nil {
				n.Columns[i] = result.(ast.SelectExpr)
			}
		}
		if n.From != nil {
			if result := Rewrite(n.From, f); result != nil {
				n.From = result.(ast.TableExpr)
			}
		}
		if n.Where != nil {
			if result := Rewrite(n.Where, f); result != nil {
				n.Where = result.(ast.Expr)
			}
		}
		for i, expr := range n.GroupBy {
			if result := Rewrite(expr, f); result != nil {
				n.GroupBy[i] = result.(ast.Expr)
			}
		}
		if n.Having != nil {
			if result := Rewrite(n.Having, f); result != nil {
				n.Having = result.(ast.Expr)
			}
		}
		for i, ob := range n.OrderBy {
			if result := Rewrite(ob.Expr, f); result != nil {
				n.OrderBy[i].Expr = result.(ast.Expr)
			}
		}
		if n.Limit != nil {
			if n.Limit.Count != nil {
				if result := Rewrite(n.Limit.Count, f); result != nil {
					n.Limit.Count = result.(ast.Expr)
				}
			}
			if n.Limit.Offset != nil {
				if result := Rewrite(n.Limit.Offset, f); result != nil {
					n.Limit.Offset = result.(ast.Expr)
				}
			}
		}

	case *ast.InsertStmt:
		if result := Rewrite(n.Table, f); result != nil {
			n.Table = result.(*ast.TableName)
		}
		for i, row := range n.Values {
			for j, val := range row {
				if result := Rewrite(val, f); result != nil {
					n.Values[i][j] = result.(ast.Expr)
				}
			}
		}
		if n.Select != nil {
			if result := Rewrite(n.Select, f); result != nil {
				n.Select = result.(*ast.SelectStmt)
			}
		}

	case *ast.UpdateStmt:
		if result := Rewrite(n.Table, f); result != nil {
			n.Table = result.(ast.TableExpr)
		}
		for i, ue := range n.Set {
			if result := Rewrite(ue.Expr, f); result != nil {
				n.Set[i].Expr = result.(ast.Expr)
			}
		}
		if n.From != nil {
			if result := Rewrite(n.From, f); result != nil {
				n.From = result.(ast.TableExpr)
			}
		}
		if n.Where != nil {
			if result := Rewrite(n.Where, f); result != nil {
				n.Where = result.(ast.Expr)
			}
		}

	case *ast.DeleteStmt:
		if result := Rewrite(n.Table, f); result != nil {
			n.Table = result.(ast.TableExpr)
		}
		if n.Using != nil {
			if result := Rewrite(n.Using, f); result != nil {
				n.Using = result.(ast.TableExpr)
			}
		}
		if n.Where != nil {
			if result := Rewrite(n.Where, f); result != nil {
				n.Where = result.(ast.Expr)
			}
		}

	case *ast.BinaryExpr:
		if result := Rewrite(n.Left, f); result != nil {
			n.Left = result.(ast.Expr)
		}
		if result := Rewrite(n.Right, f); result != nil {
			n.Right = result.(ast.Expr)
		}

	case *ast.UnaryExpr:
		if result := Rewrite(n.Operand, f); result != nil {
			n.Operand = result.(ast.Expr)
		}

	case *ast.ParenExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}

	case *ast.FuncExpr:
		for i, arg := range n.Args {
			if result := Rewrite(arg, f); result != nil {
				n.Args[i] = result.(ast.Expr)
			}
		}
		if n.Filter != nil {
			if result := Rewrite(n.Filter, f); result != nil {
				n.Filter = result.(ast.Expr)
			}
		}

	case *ast.CaseExpr:
		if n.Operand != nil {
			if result := Rewrite(n.Operand, f); result != nil {
				n.Operand = result.(ast.Expr)
			}
		}
		for i, w := range n.Whens {
			if result := Rewrite(w.Cond, f); result != nil {
				n.Whens[i].Cond = result.(ast.Expr)
			}
			if result := Rewrite(w.Result, f); result != nil {
				n.Whens[i].Result = result.(ast.Expr)
			}
		}
		if n.Else != nil {
			if result := Rewrite(n.Else, f); result != nil {
				n.Else = result.(ast.Expr)
			}
		}

	case *ast.InExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}
		for i, val := range n.Values {
			if result := Rewrite(val, f); result != nil {
				n.Values[i] = result.(ast.Expr)
			}
		}
		if n.Select != nil {
			if result := Rewrite(n.Select, f); result != nil {
				n.Select = result.(*ast.SelectStmt)
			}
		}

	case *ast.BetweenExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}
		if result := Rewrite(n.Low, f); result != nil {
			n.Low = result.(ast.Expr)
		}
		if result := Rewrite(n.High, f); result != nil {
			n.High = result.(ast.Expr)
		}

	case *ast.LikeExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}
		if result := Rewrite(n.Pattern, f); result != nil {
			n.Pattern = result.(ast.Expr)
		}
		if n.Escape != nil {
			if result := Rewrite(n.Escape, f); result != nil {
				n.Escape = result.(ast.Expr)
			}
		}

	case *ast.IsExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}

	case *ast.CastExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}

	case *ast.Subquery:
		if result := Rewrite(n.Select, f); result != nil {
			n.Select = result.(*ast.SelectStmt)
		}

	case *ast.ExistsExpr:
		if result := Rewrite(n.Subquery, f); result != nil {
			n.Subquery = result.(*ast.Subquery)
		}

	case *ast.AliasedExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.Expr)
		}

	case *ast.AliasedTableExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.TableExpr)
		}

	case *ast.JoinExpr:
		if result := Rewrite(n.Left, f); result != nil {
			n.Left = result.(ast.TableExpr)
		}
		if result := Rewrite(n.Right, f); result != nil {
			n.Right = result.(ast.TableExpr)
		}
		if n.On != nil {
			if result := Rewrite(n.On, f); result != nil {
				n.On = result.(ast.Expr)
			}
		}

	case *ast.ParenTableExpr:
		if result := Rewrite(n.Expr, f); result != nil {
			n.Expr = result.(ast.TableExpr)
		}
	}
}

// RewriteExpr is a convenience wrapper for rewriting only expressions.
func RewriteExpr(expr ast.Expr, f func(ast.Expr) ast.Expr) ast.Expr {
	result := Rewrite(expr, func(n ast.Node) ast.Node {
		if e, ok := n.(ast.Expr); ok {
			return f(e)
		}
		return n
	})
	if result == nil {
		return nil
	}
	return result.(ast.Expr)
}
