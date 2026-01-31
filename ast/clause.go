package ast

import "github.com/freeeve/machparse/token"

// TableName represents a table reference with optional qualifiers.
// Supports multi-level identifiers like catalog.schema.table.
type TableName struct {
	StartPos token.Pos
	EndPos   token.Pos
	Parts    []string // e.g., ["schema", "table"] or just ["table"]
}

func (*TableName) tableExprNode()   {}
func (t *TableName) Pos() token.Pos { return t.StartPos }
func (t *TableName) End() token.Pos { return t.EndPos }

// Name returns the table name (last part).
func (t *TableName) Name() string {
	if len(t.Parts) == 0 {
		return ""
	}
	return t.Parts[len(t.Parts)-1]
}

// Schema returns the schema qualifier (second-to-last part), or empty string.
func (t *TableName) Schema() string {
	if len(t.Parts) < 2 {
		return ""
	}
	return t.Parts[len(t.Parts)-2]
}

// Catalog returns the catalog qualifier (third-to-last part), or empty string.
func (t *TableName) Catalog() string {
	if len(t.Parts) < 3 {
		return ""
	}
	return t.Parts[len(t.Parts)-3]
}

// AliasedTableExpr represents a table with optional alias.
type AliasedTableExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     TableExpr
	Alias    string
	Hints    []*IndexHint // USE INDEX, FORCE INDEX, etc.
}

func (*AliasedTableExpr) tableExprNode()   {}
func (a *AliasedTableExpr) Pos() token.Pos { return a.StartPos }
func (a *AliasedTableExpr) End() token.Pos { return a.EndPos }

// IndexHint represents MySQL index hints.
type IndexHint struct {
	Type    IndexHintType // USE, FORCE, IGNORE
	For     IndexHintFor  // JOIN, ORDER BY, GROUP BY
	Indexes []string
}

// IndexHintType indicates the type of index hint.
type IndexHintType int

const (
	HintUse IndexHintType = iota
	HintForce
	HintIgnore
)

// IndexHintFor indicates what the hint applies to.
type IndexHintFor int

const (
	HintForAll IndexHintFor = iota
	HintForJoin
	HintForOrderBy
	HintForGroupBy
)

// JoinExpr represents a JOIN.
type JoinExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Type     JoinType
	Left     TableExpr
	Right    TableExpr
	On       Expr     // ON condition
	Using    []string // USING columns
	Natural  bool     // NATURAL JOIN
	Lateral  bool     // LATERAL (PostgreSQL)
}

// JoinType indicates the type of join.
type JoinType int

const (
	JoinInner JoinType = iota
	JoinLeft
	JoinRight
	JoinFull
	JoinCross
)

func (j JoinType) String() string {
	switch j {
	case JoinInner:
		return "INNER"
	case JoinLeft:
		return "LEFT"
	case JoinRight:
		return "RIGHT"
	case JoinFull:
		return "FULL"
	case JoinCross:
		return "CROSS"
	default:
		return "UNKNOWN"
	}
}

func (*JoinExpr) tableExprNode()   {}
func (j *JoinExpr) Pos() token.Pos { return j.StartPos }
func (j *JoinExpr) End() token.Pos { return j.EndPos }

// ParenTableExpr represents a parenthesized table expression.
type ParenTableExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     TableExpr
}

func (*ParenTableExpr) tableExprNode()   {}
func (p *ParenTableExpr) Pos() token.Pos { return p.StartPos }
func (p *ParenTableExpr) End() token.Pos { return p.EndPos }

// OrderByExpr represents an ORDER BY item.
type OrderByExpr struct {
	StartPos   token.Pos
	EndPos     token.Pos
	Expr       Expr
	Desc       bool
	NullsFirst *bool // nil = unspecified, true = NULLS FIRST, false = NULLS LAST
}

func (o *OrderByExpr) Pos() token.Pos { return o.StartPos }
func (o *OrderByExpr) End() token.Pos { return o.EndPos }

// Limit represents LIMIT/OFFSET clause.
type Limit struct {
	StartPos token.Pos
	EndPos   token.Pos
	Count    Expr // LIMIT count
	Offset   Expr // OFFSET value (optional)
}

func (l *Limit) Pos() token.Pos { return l.StartPos }
func (l *Limit) End() token.Pos { return l.EndPos }

// AliasedExpr represents a select expression with optional alias.
type AliasedExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Alias    string
}

func (*AliasedExpr) selectExprNode()  {}
func (a *AliasedExpr) Pos() token.Pos { return a.StartPos }
func (a *AliasedExpr) End() token.Pos { return a.EndPos }

// StarExpr represents * or table.*.
type StarExpr struct {
	StartPos     token.Pos
	EndPos       token.Pos
	TableName    string // table name for qualified table.*
	HasQualifier bool   // true if table qualifier was provided (even if empty)
}

func (*StarExpr) selectExprNode()  {}
func (*StarExpr) exprNode()        {} // Can appear as expression too (COUNT(*))
func (s *StarExpr) Pos() token.Pos { return s.StartPos }
func (s *StarExpr) End() token.Pos { return s.EndPos }

// WindowSpec represents window function specification.
type WindowSpec struct {
	StartPos    token.Pos
	EndPos      token.Pos
	Name        string // Reference to named window
	PartitionBy []Expr
	OrderBy     []*OrderByExpr
	Frame       *WindowFrame
}

func (w *WindowSpec) Pos() token.Pos { return w.StartPos }
func (w *WindowSpec) End() token.Pos { return w.EndPos }

// WindowDef represents a named window definition.
type WindowDef struct {
	Name string
	Spec *WindowSpec
}

// WindowFrame represents window frame specification.
type WindowFrame struct {
	Type  FrameType // ROWS, RANGE, GROUPS
	Start *FrameBound
	End   *FrameBound
}

// FrameType indicates the type of window frame.
type FrameType int

const (
	FrameRows FrameType = iota
	FrameRange
	FrameGroups
)

// FrameBound represents a window frame boundary.
type FrameBound struct {
	Type   BoundType
	Offset Expr // For N PRECEDING/FOLLOWING
}

// BoundType indicates the type of frame boundary.
type BoundType int

const (
	BoundCurrentRow BoundType = iota
	BoundUnboundedPreceding
	BoundUnboundedFollowing
	BoundPreceding
	BoundFollowing
)

// TableList represents a comma-separated list of tables (for multi-table UPDATE/DELETE).
type TableList struct {
	StartPos token.Pos
	EndPos   token.Pos
	Tables   []TableExpr
}

func (*TableList) tableExprNode()   {}
func (t *TableList) Pos() token.Pos { return t.StartPos }
func (t *TableList) End() token.Pos { return t.EndPos }

// ValuesStmt represents a VALUES statement (PostgreSQL).
type ValuesStmt struct {
	StartPos token.Pos
	EndPos   token.Pos
	Rows     [][]Expr
}

func (*ValuesStmt) statementNode()   {}
func (*ValuesStmt) tableExprNode()   {} // Can be used as table expr
func (v *ValuesStmt) Pos() token.Pos { return v.StartPos }
func (v *ValuesStmt) End() token.Pos { return v.EndPos }
