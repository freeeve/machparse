// Package machparse provides a high-performance SQL parser.
//
// machparse is a dialect-agnostic SQL parser that supports MySQL, PostgreSQL,
// and SQLite query syntax. It provides Parse, Walk, and Rewrite functionality
// similar to vitess-sqlparser.
//
// Basic usage:
//
//	stmt, err := machparse.Parse("SELECT * FROM users WHERE id = 1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(machparse.String(stmt))
//
// Walking the AST:
//
//	machparse.Walk(stmt, func(node ast.Node) bool {
//	    if col, ok := node.(*ast.ColName); ok {
//	        fmt.Printf("Found column: %s\n", col.Name)
//	    }
//	    return true
//	})
//
// Rewriting nodes:
//
//	rewritten := machparse.Rewrite(stmt, func(n ast.Node) ast.Node {
//	    // Transform nodes as needed
//	    return n
//	})
package machparse

import (
	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/format"
	"github.com/freeeve/machparse/parser"
	"github.com/freeeve/machparse/visitor"
)

// Parse parses a single SQL statement.
// The parser uses internal pooling for efficiency.
// For maximum performance when parsing many queries, call Repool(stmt)
// when done with the statement (optional, see Repool).
func Parse(sql string) (ast.Statement, error) {
	p := parser.Get(sql)
	stmt, err := p.Parse()
	parser.Put(p)
	return stmt, err
}

// ParseAll parses all statements in the input.
// For maximum performance, call Repool on each statement when done (optional).
func ParseAll(sql string) ([]ast.Statement, error) {
	p := parser.Get(sql)
	stmts, err := p.ParseAll()
	parser.Put(p)
	return stmts, err
}

// Repool returns AST nodes to internal pools for reuse.
// This is optional - if not called, nodes are garbage collected normally.
// Calling Repool after you're done with a statement improves performance
// when parsing many queries by reducing allocations.
//
// Example:
//
//	stmt, err := machparse.Parse(sql)
//	if err != nil {
//	    return err
//	}
//	defer machparse.Repool(stmt)
//	// ... use stmt ...
func Repool(stmt Statement) {
	ast.ReleaseAST(stmt)
}

// String formats an AST node back to SQL.
func String(node ast.Node) string {
	return format.String(node)
}

// Walk traverses the AST calling the function for each node.
// If the function returns false, children are not visited.
func Walk(node ast.Node, fn func(ast.Node) bool) {
	visitor.WalkFunc(node, fn)
}

// Rewrite traverses the AST allowing node replacement.
// The function is called in post-order (children first, then parent).
// Return the replacement node or the original to keep it.
func Rewrite(node ast.Node, fn func(ast.Node) ast.Node) ast.Node {
	return visitor.Rewrite(node, fn)
}

// Statement is the interface for all SQL statements.
type Statement = ast.Statement

// Expr is the interface for all expressions.
type Expr = ast.Expr

// Node is the base interface for all AST nodes.
type Node = ast.Node

// Common type aliases for convenience.
type (
	SelectStmt       = ast.SelectStmt
	InsertStmt       = ast.InsertStmt
	UpdateStmt       = ast.UpdateStmt
	DeleteStmt       = ast.DeleteStmt
	CreateTableStmt  = ast.CreateTableStmt
	AlterTableStmt   = ast.AlterTableStmt
	DropTableStmt    = ast.DropTableStmt
	CreateIndexStmt  = ast.CreateIndexStmt
	DropIndexStmt    = ast.DropIndexStmt
	TruncateStmt     = ast.TruncateStmt
	ExplainStmt      = ast.ExplainStmt
	ColName          = ast.ColName
	TableName        = ast.TableName
	Literal          = ast.Literal
	BinaryExpr       = ast.BinaryExpr
	UnaryExpr        = ast.UnaryExpr
	FuncExpr         = ast.FuncExpr
	CaseExpr         = ast.CaseExpr
	CastExpr         = ast.CastExpr
	Subquery         = ast.Subquery
	JoinExpr         = ast.JoinExpr
	AliasedExpr      = ast.AliasedExpr
	AliasedTableExpr = ast.AliasedTableExpr
	StarExpr         = ast.StarExpr
	ParenExpr        = ast.ParenExpr
	InExpr           = ast.InExpr
	BetweenExpr      = ast.BetweenExpr
	LikeExpr         = ast.LikeExpr
	IsExpr           = ast.IsExpr
	ExistsExpr       = ast.ExistsExpr
	OrderByExpr      = ast.OrderByExpr
	Limit            = ast.Limit
	WithClause       = ast.WithClause
	CTE              = ast.CTE
)

// Join types
const (
	JoinInner = ast.JoinInner
	JoinLeft  = ast.JoinLeft
	JoinRight = ast.JoinRight
	JoinFull  = ast.JoinFull
	JoinCross = ast.JoinCross
)

// Literal types
const (
	LiteralNull   = ast.LiteralNull
	LiteralInt    = ast.LiteralInt
	LiteralFloat  = ast.LiteralFloat
	LiteralString = ast.LiteralString
	LiteralBool   = ast.LiteralBool
)
