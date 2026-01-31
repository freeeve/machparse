package ast

import "github.com/freeeve/machparse/token"

// ColName represents a column reference with optional qualifiers.
// Supports multi-level identifiers like catalog.schema.table.column.
type ColName struct {
	StartPos token.Pos
	EndPos   token.Pos
	Parts    []string // e.g., ["schema", "table", "column"] or just ["column"]
}

func (*ColName) exprNode()        {}
func (c *ColName) Pos() token.Pos { return c.StartPos }
func (c *ColName) End() token.Pos { return c.EndPos }

// Name returns the column name (last part).
func (c *ColName) Name() string {
	if len(c.Parts) == 0 {
		return ""
	}
	return c.Parts[len(c.Parts)-1]
}

// Table returns the table qualifier (second-to-last part), or empty string.
func (c *ColName) Table() string {
	if len(c.Parts) < 2 {
		return ""
	}
	return c.Parts[len(c.Parts)-2]
}

// Schema returns the schema qualifier (third-to-last part), or empty string.
func (c *ColName) Schema() string {
	if len(c.Parts) < 3 {
		return ""
	}
	return c.Parts[len(c.Parts)-3]
}

// Catalog returns the catalog qualifier (fourth-to-last part), or empty string.
func (c *ColName) Catalog() string {
	if len(c.Parts) < 4 {
		return ""
	}
	return c.Parts[len(c.Parts)-4]
}

// Literal represents a literal value.
type Literal struct {
	StartPos token.Pos
	EndPos   token.Pos
	Type     LiteralType
	Value    string
}

// LiteralType indicates the type of literal.
type LiteralType int

const (
	LiteralNull LiteralType = iota
	LiteralInt
	LiteralFloat
	LiteralString
	LiteralBool
	LiteralBlob
)

func (*Literal) exprNode()        {}
func (l *Literal) Pos() token.Pos { return l.StartPos }
func (l *Literal) End() token.Pos { return l.EndPos }

// BinaryExpr represents a binary operation.
type BinaryExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Op       token.Token
	Left     Expr
	Right    Expr
}

func (*BinaryExpr) exprNode()        {}
func (b *BinaryExpr) Pos() token.Pos { return b.StartPos }
func (b *BinaryExpr) End() token.Pos { return b.EndPos }

// UnaryExpr represents a unary operation.
type UnaryExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Op       token.Token // NOT, -, ~, etc.
	Operand  Expr
}

func (*UnaryExpr) exprNode()        {}
func (u *UnaryExpr) Pos() token.Pos { return u.StartPos }
func (u *UnaryExpr) End() token.Pos { return u.EndPos }

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
}

func (*ParenExpr) exprNode()        {}
func (p *ParenExpr) Pos() token.Pos { return p.StartPos }
func (p *ParenExpr) End() token.Pos { return p.EndPos }

// FuncExpr represents a function call.
type FuncExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Name     string
	Distinct bool // COUNT(DISTINCT ...)
	Args     []Expr
	OrderBy  []*OrderByExpr // For aggregate functions with ORDER BY
	Filter   Expr           // FILTER (WHERE ...) clause
	Over     *WindowSpec    // Window function OVER clause
}

func (*FuncExpr) exprNode()        {}
func (f *FuncExpr) Pos() token.Pos { return f.StartPos }
func (f *FuncExpr) End() token.Pos { return f.EndPos }

// CastExpr represents CAST(expr AS type).
type CastExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Type     *DataType
}

func (*CastExpr) exprNode()        {}
func (c *CastExpr) Pos() token.Pos { return c.StartPos }
func (c *CastExpr) End() token.Pos { return c.EndPos }

// CaseExpr represents CASE expressions.
type CaseExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Operand  Expr // For simple CASE (optional)
	Whens    []*When
	Else     Expr // ELSE clause (optional)
}

func (*CaseExpr) exprNode()        {}
func (c *CaseExpr) Pos() token.Pos { return c.StartPos }
func (c *CaseExpr) End() token.Pos { return c.EndPos }

// When represents a WHEN clause in a CASE expression.
type When struct {
	Cond   Expr
	Result Expr
}

// InExpr represents IN expression.
type InExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Not      bool
	Values   []Expr      // List of values
	Select   *SelectStmt // Subquery (alternative to Values)
}

func (*InExpr) exprNode()        {}
func (i *InExpr) Pos() token.Pos { return i.StartPos }
func (i *InExpr) End() token.Pos { return i.EndPos }

// BetweenExpr represents BETWEEN expression.
type BetweenExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Not      bool
	Low      Expr
	High     Expr
}

func (*BetweenExpr) exprNode()        {}
func (b *BetweenExpr) Pos() token.Pos { return b.StartPos }
func (b *BetweenExpr) End() token.Pos { return b.EndPos }

// LikeExpr represents LIKE/ILIKE expression.
type LikeExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Pattern  Expr
	Not      bool
	Escape   Expr // ESCAPE character
	ILike    bool // case-insensitive (PostgreSQL)
}

func (*LikeExpr) exprNode()        {}
func (l *LikeExpr) Pos() token.Pos { return l.StartPos }
func (l *LikeExpr) End() token.Pos { return l.EndPos }

// IsExpr represents IS [NOT] NULL/TRUE/FALSE.
type IsExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Not      bool
	What     IsType // NULL, TRUE, FALSE, UNKNOWN
}

// IsType indicates what the IS expression tests for.
type IsType int

const (
	IsNull IsType = iota
	IsTrue
	IsFalse
	IsUnknown
)

func (*IsExpr) exprNode()        {}
func (i *IsExpr) Pos() token.Pos { return i.StartPos }
func (i *IsExpr) End() token.Pos { return i.EndPos }

// Subquery represents a subquery expression.
type Subquery struct {
	StartPos token.Pos
	EndPos   token.Pos
	Select   *SelectStmt
}

func (*Subquery) exprNode()        {}
func (*Subquery) tableExprNode()   {}
func (s *Subquery) Pos() token.Pos { return s.StartPos }
func (s *Subquery) End() token.Pos { return s.EndPos }

// ExistsExpr represents EXISTS (subquery).
type ExistsExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Not      bool
	Subquery *Subquery
}

func (*ExistsExpr) exprNode()        {}
func (e *ExistsExpr) Pos() token.Pos { return e.StartPos }
func (e *ExistsExpr) End() token.Pos { return e.EndPos }

// Param represents a query parameter.
type Param struct {
	StartPos token.Pos
	EndPos   token.Pos
	Type     ParamType // ?, $1, :name
	Name     string    // For named params
	Index    int       // For positional params
}

// ParamType indicates the type of parameter.
type ParamType int

const (
	ParamQuestion ParamType = iota // ?
	ParamDollar                    // $1, $2
	ParamColon                     // :name
	ParamAt                        // @name (MySQL)
)

func (*Param) exprNode()        {}
func (p *Param) Pos() token.Pos { return p.StartPos }
func (p *Param) End() token.Pos { return p.EndPos }

// ArrayExpr represents an array constructor.
type ArrayExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Elements []Expr
}

func (*ArrayExpr) exprNode()        {}
func (a *ArrayExpr) Pos() token.Pos { return a.StartPos }
func (a *ArrayExpr) End() token.Pos { return a.EndPos }

// SubscriptExpr represents array subscript access.
type SubscriptExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	Index    Expr
}

func (*SubscriptExpr) exprNode()        {}
func (s *SubscriptExpr) Pos() token.Pos { return s.StartPos }
func (s *SubscriptExpr) End() token.Pos { return s.EndPos }

// IntervalExpr represents INTERVAL expression.
type IntervalExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Value    Expr
	Unit     string // YEAR, MONTH, DAY, etc.
}

func (*IntervalExpr) exprNode()        {}
func (i *IntervalExpr) Pos() token.Pos { return i.StartPos }
func (i *IntervalExpr) End() token.Pos { return i.EndPos }

// ExtractExpr represents EXTRACT(field FROM source).
type ExtractExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Field    string // YEAR, MONTH, DAY, etc.
	Source   Expr
}

func (*ExtractExpr) exprNode()        {}
func (e *ExtractExpr) Pos() token.Pos { return e.StartPos }
func (e *ExtractExpr) End() token.Pos { return e.EndPos }

// TrimExpr represents TRIM expressions.
type TrimExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	TrimType TrimType // LEADING, TRAILING, BOTH
	TrimChar Expr     // Characters to trim (optional)
	Expr     Expr     // Expression to trim
}

// TrimType indicates the trim direction.
type TrimType int

const (
	TrimBoth TrimType = iota
	TrimLeading
	TrimTrailing
)

func (*TrimExpr) exprNode()        {}
func (t *TrimExpr) Pos() token.Pos { return t.StartPos }
func (t *TrimExpr) End() token.Pos { return t.EndPos }

// SubstringExpr represents SUBSTRING expressions.
type SubstringExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Expr     Expr
	From     Expr // Starting position
	For      Expr // Length (optional)
}

func (*SubstringExpr) exprNode()        {}
func (s *SubstringExpr) Pos() token.Pos { return s.StartPos }
func (s *SubstringExpr) End() token.Pos { return s.EndPos }

// PositionExpr represents POSITION(substring IN string).
type PositionExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
	Needle   Expr
	Haystack Expr
}

func (*PositionExpr) exprNode()        {}
func (p *PositionExpr) Pos() token.Pos { return p.StartPos }
func (p *PositionExpr) End() token.Pos { return p.EndPos }

// CollateExpr represents COLLATE expression.
type CollateExpr struct {
	StartPos  token.Pos
	EndPos    token.Pos
	Expr      Expr
	Collation string
}

func (*CollateExpr) exprNode()        {}
func (c *CollateExpr) Pos() token.Pos { return c.StartPos }
func (c *CollateExpr) End() token.Pos { return c.EndPos }
