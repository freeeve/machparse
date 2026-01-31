package ast

import "github.com/freeeve/machparse/token"

// SelectStmt represents a SELECT statement.
type SelectStmt struct {
	StartPos   token.Pos
	EndPos     token.Pos
	With       *WithClause    // WITH clause (CTEs)
	Distinct   bool           // DISTINCT
	Columns    []SelectExpr   // SELECT expressions
	From       TableExpr      // FROM clause
	Where      Expr           // WHERE clause (optional)
	GroupBy    []Expr         // GROUP BY expressions
	Having     Expr           // HAVING clause (optional)
	OrderBy    []*OrderByExpr // ORDER BY expressions
	Limit      *Limit         // LIMIT clause (optional)
	Lock       string         // FOR UPDATE, etc.
	Into       *SelectInto    // INTO clause (optional)
	WindowDefs []*WindowDef   // WINDOW definitions
}

func (*SelectStmt) statementNode()   {}
func (s *SelectStmt) Pos() token.Pos { return s.StartPos }
func (s *SelectStmt) End() token.Pos { return s.EndPos }

// SelectInto represents SELECT ... INTO.
type SelectInto struct {
	Outfile  string
	Dumpfile string
	Vars     []string
}

// InsertStmt represents an INSERT statement.
type InsertStmt struct {
	StartPos          token.Pos
	EndPos            token.Pos
	With              *WithClause // WITH clause (CTEs)
	Replace           bool        // REPLACE INTO (MySQL)
	Ignore            bool        // INSERT IGNORE (MySQL)
	Table             *TableName
	Columns           []*ColName    // Column list (optional)
	Values            [][]Expr      // VALUES rows
	Select            *SelectStmt   // INSERT ... SELECT
	OnDuplicateUpdate []*UpdateExpr // ON DUPLICATE KEY UPDATE (MySQL)
	OnConflict        *OnConflict   // ON CONFLICT (PostgreSQL)
	Returning         []SelectExpr  // RETURNING clause (PostgreSQL)
}

func (*InsertStmt) statementNode()   {}
func (i *InsertStmt) Pos() token.Pos { return i.StartPos }
func (i *InsertStmt) End() token.Pos { return i.EndPos }

// OnConflict represents PostgreSQL ON CONFLICT clause.
type OnConflict struct {
	Columns   []string // Conflict columns
	Where     Expr     // Optional WHERE for partial index
	DoNothing bool
	Updates   []*UpdateExpr // SET expressions for DO UPDATE
}

// UpdateStmt represents an UPDATE statement.
type UpdateStmt struct {
	StartPos  token.Pos
	EndPos    token.Pos
	With      *WithClause // WITH clause (CTEs)
	Table     TableExpr
	Set       []*UpdateExpr
	From      TableExpr // PostgreSQL FROM clause
	Where     Expr
	OrderBy   []*OrderByExpr // MySQL extension
	Limit     *Limit         // MySQL extension
	Returning []SelectExpr   // PostgreSQL
}

func (*UpdateStmt) statementNode()   {}
func (u *UpdateStmt) Pos() token.Pos { return u.StartPos }
func (u *UpdateStmt) End() token.Pos { return u.EndPos }

// UpdateExpr represents SET column = value.
type UpdateExpr struct {
	Column *ColName
	Expr   Expr
}

// DeleteStmt represents a DELETE statement.
type DeleteStmt struct {
	StartPos  token.Pos
	EndPos    token.Pos
	With      *WithClause // WITH clause (CTEs)
	Table     TableExpr
	Using     TableExpr // USING clause (PostgreSQL)
	Where     Expr
	OrderBy   []*OrderByExpr // MySQL extension
	Limit     *Limit         // MySQL extension
	Returning []SelectExpr   // PostgreSQL
}

func (*DeleteStmt) statementNode()   {}
func (d *DeleteStmt) Pos() token.Pos { return d.StartPos }
func (d *DeleteStmt) End() token.Pos { return d.EndPos }

// SetOp represents UNION/INTERSECT/EXCEPT.
type SetOp struct {
	StartPos token.Pos
	EndPos   token.Pos
	Type     SetOpType // UNION, INTERSECT, EXCEPT
	All      bool
	Left     Statement
	Right    Statement
	OrderBy  []*OrderByExpr
	Limit    *Limit
}

// SetOpType indicates the type of set operation.
type SetOpType int

const (
	Union SetOpType = iota
	Intersect
	Except
)

func (*SetOp) statementNode()   {}
func (s *SetOp) Pos() token.Pos { return s.StartPos }
func (s *SetOp) End() token.Pos { return s.EndPos }

// WithClause represents a WITH clause (common table expressions).
type WithClause struct {
	Recursive bool
	CTEs      []*CTE
}

// CTE represents a single common table expression.
type CTE struct {
	Name    string
	Columns []string
	Query   Statement
}

// CreateTableStmt represents CREATE TABLE.
type CreateTableStmt struct {
	StartPos    token.Pos
	EndPos      token.Pos
	IfNotExists bool
	Temporary   bool
	Table       *TableName
	Columns     []*ColumnDef
	Constraints []*TableConstraint
	Options     []*TableOption
	As          *SelectStmt // CREATE TABLE AS SELECT
}

func (*CreateTableStmt) statementNode()   {}
func (c *CreateTableStmt) Pos() token.Pos { return c.StartPos }
func (c *CreateTableStmt) End() token.Pos { return c.EndPos }

// ColumnDef represents a column definition.
type ColumnDef struct {
	Name        string
	Type        *DataType
	Constraints []*ColumnConstraint
}

// DataType represents a SQL data type.
type DataType struct {
	Name      string // INT, VARCHAR, etc.
	Length    *int   // VARCHAR(255)
	Precision *int   // DECIMAL(10,2)
	Scale     *int
	Array     bool   // PostgreSQL array type
	Unsigned  bool   // MySQL UNSIGNED
	Charset   string // MySQL CHARACTER SET
	Collation string // COLLATE
}

// ColumnConstraint represents a column-level constraint.
type ColumnConstraint struct {
	Name       string // optional constraint name
	Type       ConstraintType
	NotNull    bool
	Default    Expr
	Check      Expr
	References *ForeignKeyRef
	Generated  *GeneratedColumn
}

// ConstraintType indicates the type of constraint.
type ConstraintType int

const (
	ConstraintPrimaryKey ConstraintType = iota
	ConstraintUnique
	ConstraintNotNull
	ConstraintDefault
	ConstraintCheck
	ConstraintForeignKey
	ConstraintGenerated
)

// GeneratedColumn represents a generated column specification.
type GeneratedColumn struct {
	Expr   Expr
	Stored bool // STORED vs VIRTUAL
}

// TableConstraint represents a table-level constraint.
type TableConstraint struct {
	Name       string
	Type       ConstraintType
	Columns    []string
	References *ForeignKeyRef
	Check      Expr
}

// ForeignKeyRef represents foreign key reference.
type ForeignKeyRef struct {
	Table    *TableName
	Columns  []string
	OnDelete RefAction
	OnUpdate RefAction
}

// RefAction indicates foreign key referential action.
type RefAction int

const (
	RefNoAction RefAction = iota
	RefCascade
	RefSetNull
	RefSetDefault
	RefRestrict
)

// TableOption represents a table option.
type TableOption struct {
	Name  string
	Value string
}

// AlterTableStmt represents ALTER TABLE.
type AlterTableStmt struct {
	StartPos token.Pos
	EndPos   token.Pos
	Table    *TableName
	Actions  []AlterTableAction
}

func (*AlterTableStmt) statementNode()   {}
func (a *AlterTableStmt) Pos() token.Pos { return a.StartPos }
func (a *AlterTableStmt) End() token.Pos { return a.EndPos }

// AlterTableAction is an interface for ALTER TABLE actions.
type AlterTableAction interface {
	alterTableAction()
}

// AddColumn represents ADD COLUMN.
type AddColumn struct {
	Column *ColumnDef
}

func (*AddColumn) alterTableAction() {}

// DropColumn represents DROP COLUMN.
type DropColumn struct {
	Name     string
	IfExists bool
	Cascade  bool
}

func (*DropColumn) alterTableAction() {}

// ModifyColumn represents MODIFY/ALTER COLUMN.
type ModifyColumn struct {
	Name        string
	NewDef      *ColumnDef
	SetDefault  Expr
	DropDefault bool
	SetNotNull  bool
	DropNotNull bool
}

func (*ModifyColumn) alterTableAction() {}

// RenameColumn represents RENAME COLUMN.
type RenameColumn struct {
	OldName string
	NewName string
}

func (*RenameColumn) alterTableAction() {}

// AddConstraint represents ADD CONSTRAINT.
type AddConstraint struct {
	Constraint *TableConstraint
}

func (*AddConstraint) alterTableAction() {}

// DropConstraint represents DROP CONSTRAINT.
type DropConstraint struct {
	Name     string
	IfExists bool
	Cascade  bool
}

func (*DropConstraint) alterTableAction() {}

// RenameTable represents RENAME TO.
type RenameTable struct {
	NewName *TableName
}

func (*RenameTable) alterTableAction() {}

// DropTableStmt represents DROP TABLE.
type DropTableStmt struct {
	StartPos token.Pos
	EndPos   token.Pos
	IfExists bool
	Tables   []*TableName
	Cascade  bool
}

func (*DropTableStmt) statementNode()   {}
func (d *DropTableStmt) Pos() token.Pos { return d.StartPos }
func (d *DropTableStmt) End() token.Pos { return d.EndPos }

// CreateIndexStmt represents CREATE INDEX.
type CreateIndexStmt struct {
	StartPos    token.Pos
	EndPos      token.Pos
	IfNotExists bool
	Unique      bool
	Concurrent  bool // PostgreSQL CONCURRENTLY
	Name        string
	Table       *TableName
	Columns     []*IndexColumn
	Using       string // btree, hash, etc.
	Where       Expr   // Partial index (PostgreSQL)
}

func (*CreateIndexStmt) statementNode()   {}
func (c *CreateIndexStmt) Pos() token.Pos { return c.StartPos }
func (c *CreateIndexStmt) End() token.Pos { return c.EndPos }

// IndexColumn represents a column in an index.
type IndexColumn struct {
	Column string
	Expr   Expr // Expression index
	Desc   bool
	Nulls  string // FIRST, LAST
}

// DropIndexStmt represents DROP INDEX.
type DropIndexStmt struct {
	StartPos   token.Pos
	EndPos     token.Pos
	IfExists   bool
	Concurrent bool // PostgreSQL CONCURRENTLY
	Name       string
	Table      *TableName // MySQL requires table name
	Cascade    bool
}

func (*DropIndexStmt) statementNode()   {}
func (d *DropIndexStmt) Pos() token.Pos { return d.StartPos }
func (d *DropIndexStmt) End() token.Pos { return d.EndPos }

// TruncateStmt represents TRUNCATE TABLE.
type TruncateStmt struct {
	StartPos token.Pos
	EndPos   token.Pos
	Tables   []*TableName
	Cascade  bool
}

func (*TruncateStmt) statementNode()   {}
func (t *TruncateStmt) Pos() token.Pos { return t.StartPos }
func (t *TruncateStmt) End() token.Pos { return t.EndPos }

// ExplainStmt represents EXPLAIN.
type ExplainStmt struct {
	StartPos token.Pos
	EndPos   token.Pos
	Analyze  bool
	Verbose  bool
	Format   string // TEXT, JSON, YAML, XML
	Stmt     Statement
}

func (*ExplainStmt) statementNode()   {}
func (e *ExplainStmt) Pos() token.Pos { return e.StartPos }
func (e *ExplainStmt) End() token.Pos { return e.EndPos }
