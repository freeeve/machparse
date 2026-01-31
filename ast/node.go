// Package ast defines the abstract syntax tree for SQL statements.
package ast

import "github.com/freeeve/machparse/token"

// Node is the base interface for all AST nodes.
type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// Statement represents a SQL statement.
type Statement interface {
	Node
	statementNode()
}

// Expr represents an expression.
type Expr interface {
	Node
	exprNode()
}

// TableExpr represents a table expression (in FROM clause).
type TableExpr interface {
	Node
	tableExprNode()
}

// SelectExpr represents a select expression (in SELECT clause).
type SelectExpr interface {
	Node
	selectExprNode()
}

// SQLNode is an alias for compatibility with vitess API.
type SQLNode = Node
