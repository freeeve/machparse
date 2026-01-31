// Package format provides SQL generation from AST nodes.
package format

import (
	"bytes"
	"strings"

	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/token"
)

// Options controls formatting behavior.
type Options struct {
	Uppercase bool   // Uppercase keywords
	Indent    string // Indentation string (unused for single-line output)
}

// DefaultOptions are the default formatting options.
var DefaultOptions = Options{
	Uppercase: true,
	Indent:    "  ",
}

// Formatter generates SQL from AST nodes.
type Formatter struct {
	buf  bytes.Buffer
	opts Options
}

// New creates a new formatter with the given options.
func New(opts Options) *Formatter {
	return &Formatter{opts: opts}
}

// String formats an AST node to a SQL string.
func String(node ast.Node) string {
	f := New(DefaultOptions)
	f.Format(node)
	return f.String()
}

// Format formats a node to the internal buffer.
func (f *Formatter) Format(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.SelectStmt:
		f.formatSelect(n)
	case *ast.InsertStmt:
		f.formatInsert(n)
	case *ast.UpdateStmt:
		f.formatUpdate(n)
	case *ast.DeleteStmt:
		f.formatDelete(n)
	case *ast.CreateTableStmt:
		f.formatCreateTable(n)
	case *ast.AlterTableStmt:
		f.formatAlterTable(n)
	case *ast.DropTableStmt:
		f.formatDropTable(n)
	case *ast.CreateIndexStmt:
		f.formatCreateIndex(n)
	case *ast.DropIndexStmt:
		f.formatDropIndex(n)
	case *ast.TruncateStmt:
		f.formatTruncate(n)
	case *ast.ExplainStmt:
		f.formatExplain(n)
	case *ast.SetOp:
		f.formatSetOp(n)
	case *ast.BinaryExpr:
		f.formatBinaryExpr(n)
	case *ast.UnaryExpr:
		f.formatUnaryExpr(n)
	case *ast.ParenExpr:
		f.write("(")
		f.Format(n.Expr)
		f.write(")")
	case *ast.FuncExpr:
		f.formatFuncExpr(n)
	case *ast.CaseExpr:
		f.formatCaseExpr(n)
	case *ast.CastExpr:
		f.formatCastExpr(n)
	case *ast.ColName:
		f.formatColName(n)
	case *ast.Literal:
		f.formatLiteral(n)
	case *ast.Param:
		f.formatParam(n)
	case *ast.TableName:
		f.formatTableName(n)
	case *ast.AliasedTableExpr:
		f.formatAliasedTableExpr(n)
	case *ast.JoinExpr:
		f.formatJoinExpr(n)
	case *ast.ParenTableExpr:
		f.write("(")
		f.Format(n.Expr)
		f.write(")")
	case *ast.Subquery:
		f.write("(")
		f.Format(n.Select)
		f.write(")")
	case *ast.AliasedExpr:
		f.Format(n.Expr)
		if n.Alias != "" {
			f.write(" ")
			f.writeKeyword("AS")
			f.write(" ")
			f.writeIdent(n.Alias)
		}
	case *ast.StarExpr:
		if n.HasQualifier {
			f.writeIdent(n.TableName)
			f.write(".")
		}
		f.write("*")
	case *ast.InExpr:
		f.formatInExpr(n)
	case *ast.BetweenExpr:
		f.formatBetweenExpr(n)
	case *ast.LikeExpr:
		f.formatLikeExpr(n)
	case *ast.IsExpr:
		f.formatIsExpr(n)
	case *ast.ExistsExpr:
		f.formatExistsExpr(n)
	case *ast.IntervalExpr:
		f.formatIntervalExpr(n)
	case *ast.ExtractExpr:
		f.formatExtractExpr(n)
	case *ast.TrimExpr:
		f.formatTrimExpr(n)
	case *ast.SubstringExpr:
		f.formatSubstringExpr(n)
	case *ast.ArrayExpr:
		f.formatArrayExpr(n)
	case *ast.SubscriptExpr:
		// Format array subscripts with space after [ to distinguish from SQL Server
		// bracket identifiers. The lexer treats [ followed by space as LBRACKET,
		// not as start of bracket identifier.
		f.Format(n.Expr)
		f.write("[ ")
		f.Format(n.Index)
		f.write(" ]")
	case *ast.CollateExpr:
		f.Format(n.Expr)
		f.write(" ")
		f.writeKeyword("COLLATE")
		f.write(" ")
		f.write(n.Collation)
	case *ast.ValuesStmt:
		f.formatValuesStmt(n)
	}
}

// String returns the formatted SQL.
func (f *Formatter) String() string {
	return f.buf.String()
}

func (f *Formatter) write(s string) {
	f.buf.WriteString(s)
}

func (f *Formatter) writeKeyword(kw string) {
	if f.opts.Uppercase {
		f.buf.WriteString(strings.ToUpper(kw))
	} else {
		f.buf.WriteString(strings.ToLower(kw))
	}
}

func (f *Formatter) writeIdent(id string) {
	if needsQuoting(id) {
		f.buf.WriteByte('"')
		f.buf.WriteString(strings.ReplaceAll(id, `"`, `""`))
		f.buf.WriteByte('"')
	} else {
		f.buf.WriteString(id)
	}
}

// writeFuncName writes a function name. Unlike writeIdent, it doesn't quote
// keywords since many SQL functions have keyword names (ANY, ALL, COUNT, etc.)
func (f *Formatter) writeFuncName(name string) {
	if needsQuotingNonKeyword(name) {
		f.buf.WriteByte('"')
		f.buf.WriteString(strings.ReplaceAll(name, `"`, `""`))
		f.buf.WriteByte('"')
	} else {
		f.buf.WriteString(name)
	}
}

func (f *Formatter) formatSelect(s *ast.SelectStmt) {
	if s.With != nil {
		f.formatWithClause(s.With)
		f.write(" ")
	}

	f.writeKeyword("SELECT")

	if s.Distinct {
		f.write(" ")
		f.writeKeyword("DISTINCT")
	}

	f.write(" ")

	// Columns
	for i, col := range s.Columns {
		if i > 0 {
			f.write(", ")
		}
		f.Format(col)
	}

	// FROM
	if s.From != nil {
		f.write(" ")
		f.writeKeyword("FROM")
		f.write(" ")
		f.Format(s.From)
	}

	// WHERE
	if s.Where != nil {
		f.write(" ")
		f.writeKeyword("WHERE")
		f.write(" ")
		f.Format(s.Where)
	}

	// GROUP BY
	if len(s.GroupBy) > 0 {
		f.write(" ")
		f.writeKeyword("GROUP BY")
		f.write(" ")
		for i, expr := range s.GroupBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(expr)
		}
	}

	// HAVING
	if s.Having != nil {
		f.write(" ")
		f.writeKeyword("HAVING")
		f.write(" ")
		f.Format(s.Having)
	}

	// ORDER BY
	if len(s.OrderBy) > 0 {
		f.write(" ")
		f.writeKeyword("ORDER BY")
		f.write(" ")
		for i, ob := range s.OrderBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(ob.Expr)
			if ob.Desc {
				f.write(" ")
				f.writeKeyword("DESC")
			}
			if ob.NullsFirst != nil {
				f.write(" ")
				f.writeKeyword("NULLS")
				f.write(" ")
				if *ob.NullsFirst {
					f.writeKeyword("FIRST")
				} else {
					f.writeKeyword("LAST")
				}
			}
		}
	}

	// LIMIT
	if s.Limit != nil {
		if s.Limit.Count != nil {
			f.write(" ")
			f.writeKeyword("LIMIT")
			f.write(" ")
			f.Format(s.Limit.Count)
		}
		if s.Limit.Offset != nil {
			f.write(" ")
			f.writeKeyword("OFFSET")
			f.write(" ")
			f.Format(s.Limit.Offset)
		}
	}

	// FOR UPDATE/SHARE
	if s.Lock != "" {
		f.write(" ")
		f.writeKeyword("FOR")
		f.write(" ")
		f.writeKeyword(s.Lock)
	}
}

func (f *Formatter) formatWithClause(w *ast.WithClause) {
	f.writeKeyword("WITH")
	if w.Recursive {
		f.write(" ")
		f.writeKeyword("RECURSIVE")
	}
	f.write(" ")
	for i, cte := range w.CTEs {
		if i > 0 {
			f.write(", ")
		}
		f.writeIdent(cte.Name)
		if len(cte.Columns) > 0 {
			f.write(" (")
			for j, col := range cte.Columns {
				if j > 0 {
					f.write(", ")
				}
				f.writeIdent(col)
			}
			f.write(")")
		}
		f.write(" ")
		f.writeKeyword("AS")
		f.write(" (")
		f.Format(cte.Query)
		f.write(")")
	}
}

func (f *Formatter) formatInsert(s *ast.InsertStmt) {
	if s.With != nil {
		f.formatWithClause(s.With)
		f.write(" ")
	}

	if s.Replace {
		f.writeKeyword("REPLACE")
	} else {
		f.writeKeyword("INSERT")
	}

	if s.Ignore {
		f.write(" ")
		f.writeKeyword("IGNORE")
	}

	f.write(" ")
	f.writeKeyword("INTO")
	f.write(" ")
	f.Format(s.Table)

	if len(s.Columns) > 0 {
		f.write(" (")
		for i, col := range s.Columns {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(col.Name())
		}
		f.write(")")
	}

	if s.Select != nil {
		f.write(" ")
		f.Format(s.Select)
	} else if len(s.Values) > 0 {
		f.write(" ")
		f.writeKeyword("VALUES")
		f.write(" ")
		for i, row := range s.Values {
			if i > 0 {
				f.write(", ")
			}
			f.write("(")
			for j, val := range row {
				if j > 0 {
					f.write(", ")
				}
				f.Format(val)
			}
			f.write(")")
		}
	}

	if len(s.OnDuplicateUpdate) > 0 {
		f.write(" ")
		f.writeKeyword("ON DUPLICATE KEY UPDATE")
		f.write(" ")
		for i, ue := range s.OnDuplicateUpdate {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(ue.Column.Name())
			f.write(" = ")
			f.Format(ue.Expr)
		}
	}

	if s.OnConflict != nil {
		f.write(" ")
		f.writeKeyword("ON CONFLICT")
		if len(s.OnConflict.Columns) > 0 {
			f.write(" (")
			for i, col := range s.OnConflict.Columns {
				if i > 0 {
					f.write(", ")
				}
				f.writeIdent(col)
			}
			f.write(")")
		}
		f.write(" ")
		f.writeKeyword("DO")
		f.write(" ")
		if s.OnConflict.DoNothing {
			f.writeKeyword("NOTHING")
		} else {
			f.writeKeyword("UPDATE SET")
			f.write(" ")
			for i, ue := range s.OnConflict.Updates {
				if i > 0 {
					f.write(", ")
				}
				f.writeIdent(ue.Column.Name())
				f.write(" = ")
				f.Format(ue.Expr)
			}
		}
	}

	if len(s.Returning) > 0 {
		f.write(" ")
		f.writeKeyword("RETURNING")
		f.write(" ")
		for i, col := range s.Returning {
			if i > 0 {
				f.write(", ")
			}
			f.Format(col)
		}
	}
}

func (f *Formatter) formatUpdate(s *ast.UpdateStmt) {
	if s.With != nil {
		f.formatWithClause(s.With)
		f.write(" ")
	}

	f.writeKeyword("UPDATE")
	f.write(" ")
	f.Format(s.Table)
	f.write(" ")
	f.writeKeyword("SET")
	f.write(" ")

	for i, ue := range s.Set {
		if i > 0 {
			f.write(", ")
		}
		f.formatColName(ue.Column)
		f.write(" = ")
		f.Format(ue.Expr)
	}

	if s.From != nil {
		f.write(" ")
		f.writeKeyword("FROM")
		f.write(" ")
		f.Format(s.From)
	}

	if s.Where != nil {
		f.write(" ")
		f.writeKeyword("WHERE")
		f.write(" ")
		f.Format(s.Where)
	}

	if len(s.OrderBy) > 0 {
		f.write(" ")
		f.writeKeyword("ORDER BY")
		f.write(" ")
		for i, ob := range s.OrderBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(ob.Expr)
			if ob.Desc {
				f.write(" ")
				f.writeKeyword("DESC")
			}
		}
	}

	if s.Limit != nil && s.Limit.Count != nil {
		f.write(" ")
		f.writeKeyword("LIMIT")
		f.write(" ")
		f.Format(s.Limit.Count)
	}

	if len(s.Returning) > 0 {
		f.write(" ")
		f.writeKeyword("RETURNING")
		f.write(" ")
		for i, col := range s.Returning {
			if i > 0 {
				f.write(", ")
			}
			f.Format(col)
		}
	}
}

func (f *Formatter) formatDelete(s *ast.DeleteStmt) {
	if s.With != nil {
		f.formatWithClause(s.With)
		f.write(" ")
	}

	f.writeKeyword("DELETE FROM")
	f.write(" ")
	f.Format(s.Table)

	if s.Using != nil {
		f.write(" ")
		f.writeKeyword("USING")
		f.write(" ")
		f.Format(s.Using)
	}

	if s.Where != nil {
		f.write(" ")
		f.writeKeyword("WHERE")
		f.write(" ")
		f.Format(s.Where)
	}

	if len(s.OrderBy) > 0 {
		f.write(" ")
		f.writeKeyword("ORDER BY")
		f.write(" ")
		for i, ob := range s.OrderBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(ob.Expr)
			if ob.Desc {
				f.write(" ")
				f.writeKeyword("DESC")
			}
		}
	}

	if s.Limit != nil && s.Limit.Count != nil {
		f.write(" ")
		f.writeKeyword("LIMIT")
		f.write(" ")
		f.Format(s.Limit.Count)
	}

	if len(s.Returning) > 0 {
		f.write(" ")
		f.writeKeyword("RETURNING")
		f.write(" ")
		for i, col := range s.Returning {
			if i > 0 {
				f.write(", ")
			}
			f.Format(col)
		}
	}
}

func (f *Formatter) formatCreateTable(s *ast.CreateTableStmt) {
	f.writeKeyword("CREATE")
	if s.Temporary {
		f.write(" ")
		f.writeKeyword("TEMPORARY")
	}
	f.write(" ")
	f.writeKeyword("TABLE")

	if s.IfNotExists {
		f.write(" ")
		f.writeKeyword("IF NOT EXISTS")
	}

	f.write(" ")
	f.Format(s.Table)

	if s.As != nil {
		f.write(" ")
		f.writeKeyword("AS")
		f.write(" ")
		f.Format(s.As)
		return
	}

	f.write(" (")
	for i, col := range s.Columns {
		if i > 0 {
			f.write(", ")
		}
		f.formatColumnDef(col)
	}
	for i, cons := range s.Constraints {
		if len(s.Columns) > 0 || i > 0 {
			f.write(", ")
		}
		f.formatTableConstraint(cons)
	}
	f.write(")")

	for _, opt := range s.Options {
		f.write(" ")
		f.write(opt.Name)
		f.write("=")
		f.write(opt.Value)
	}
}

func (f *Formatter) formatColumnDef(col *ast.ColumnDef) {
	f.writeIdent(col.Name)
	f.write(" ")
	f.formatDataType(col.Type)

	for _, cons := range col.Constraints {
		f.write(" ")
		f.formatColumnConstraint(cons)
	}
}

func (f *Formatter) formatDataType(dt *ast.DataType) {
	if dt == nil {
		return
	}
	// Use writeIdent to handle quoted identifiers as type names
	if needsQuoting(dt.Name) {
		f.writeIdent(dt.Name)
	} else {
		f.writeKeyword(dt.Name)
	}
	if dt.Length != nil {
		f.write("(")
		f.write(itoa(*dt.Length))
		if dt.Scale != nil {
			f.write(", ")
			f.write(itoa(*dt.Scale))
		}
		f.write(")")
	}
	if dt.Unsigned {
		f.write(" ")
		f.writeKeyword("UNSIGNED")
	}
	if dt.Array {
		f.write("[]")
	}
}

func (f *Formatter) formatColumnConstraint(cons *ast.ColumnConstraint) {
	switch cons.Type {
	case ast.ConstraintNotNull:
		f.writeKeyword("NOT NULL")
	case ast.ConstraintPrimaryKey:
		f.writeKeyword("PRIMARY KEY")
	case ast.ConstraintUnique:
		f.writeKeyword("UNIQUE")
	case ast.ConstraintDefault:
		f.writeKeyword("DEFAULT")
		f.write(" ")
		f.Format(cons.Default)
	case ast.ConstraintCheck:
		f.writeKeyword("CHECK")
		f.write(" (")
		f.Format(cons.Check)
		f.write(")")
	case ast.ConstraintForeignKey:
		f.writeKeyword("REFERENCES")
		f.write(" ")
		f.Format(cons.References.Table)
		if len(cons.References.Columns) > 0 {
			f.write(" (")
			for i, col := range cons.References.Columns {
				if i > 0 {
					f.write(", ")
				}
				f.writeIdent(col)
			}
			f.write(")")
		}
	}
}

func (f *Formatter) formatTableConstraint(cons *ast.TableConstraint) {
	if cons.Name != "" {
		f.writeKeyword("CONSTRAINT")
		f.write(" ")
		f.writeIdent(cons.Name)
		f.write(" ")
	}

	switch cons.Type {
	case ast.ConstraintPrimaryKey:
		f.writeKeyword("PRIMARY KEY")
		f.write(" (")
		for i, col := range cons.Columns {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(col)
		}
		f.write(")")
	case ast.ConstraintUnique:
		f.writeKeyword("UNIQUE")
		f.write(" (")
		for i, col := range cons.Columns {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(col)
		}
		f.write(")")
	case ast.ConstraintForeignKey:
		f.writeKeyword("FOREIGN KEY")
		f.write(" (")
		for i, col := range cons.Columns {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(col)
		}
		f.write(") ")
		f.writeKeyword("REFERENCES")
		f.write(" ")
		f.Format(cons.References.Table)
		if len(cons.References.Columns) > 0 {
			f.write(" (")
			for i, col := range cons.References.Columns {
				if i > 0 {
					f.write(", ")
				}
				f.writeIdent(col)
			}
			f.write(")")
		}
	case ast.ConstraintCheck:
		f.writeKeyword("CHECK")
		f.write(" (")
		f.Format(cons.Check)
		f.write(")")
	}
}

func (f *Formatter) formatAlterTable(s *ast.AlterTableStmt) {
	f.writeKeyword("ALTER TABLE")
	f.write(" ")
	f.Format(s.Table)

	for i, action := range s.Actions {
		if i > 0 {
			f.write(",")
		}
		f.write(" ")
		switch a := action.(type) {
		case *ast.AddColumn:
			f.writeKeyword("ADD COLUMN")
			f.write(" ")
			f.formatColumnDef(a.Column)
		case *ast.DropColumn:
			f.writeKeyword("DROP COLUMN")
			if a.IfExists {
				f.write(" ")
				f.writeKeyword("IF EXISTS")
			}
			f.write(" ")
			f.writeIdent(a.Name)
			if a.Cascade {
				f.write(" ")
				f.writeKeyword("CASCADE")
			}
		case *ast.RenameColumn:
			f.writeKeyword("RENAME COLUMN")
			f.write(" ")
			f.writeIdent(a.OldName)
			f.write(" ")
			f.writeKeyword("TO")
			f.write(" ")
			f.writeIdent(a.NewName)
		case *ast.RenameTable:
			f.writeKeyword("RENAME TO")
			f.write(" ")
			f.Format(a.NewName)
		case *ast.ModifyColumn:
			f.writeKeyword("MODIFY COLUMN")
			f.write(" ")
			if a.NewDef != nil {
				f.formatColumnDef(a.NewDef)
			} else {
				f.writeIdent(a.Name)
				if a.SetNotNull {
					f.write(" ")
					f.writeKeyword("SET NOT NULL")
				}
				if a.SetDefault != nil {
					f.write(" ")
					f.writeKeyword("SET DEFAULT")
					f.write(" ")
					f.Format(a.SetDefault)
				}
				if a.DropNotNull {
					f.write(" ")
					f.writeKeyword("DROP NOT NULL")
				}
				if a.DropDefault {
					f.write(" ")
					f.writeKeyword("DROP DEFAULT")
				}
			}
		case *ast.AddConstraint:
			f.writeKeyword("ADD")
			f.write(" ")
			f.formatTableConstraint(a.Constraint)
		case *ast.DropConstraint:
			f.writeKeyword("DROP CONSTRAINT")
			if a.IfExists {
				f.write(" ")
				f.writeKeyword("IF EXISTS")
			}
			f.write(" ")
			f.writeIdent(a.Name)
			if a.Cascade {
				f.write(" ")
				f.writeKeyword("CASCADE")
			}
		}
	}
}

func (f *Formatter) formatDropTable(s *ast.DropTableStmt) {
	f.writeKeyword("DROP TABLE")
	if s.IfExists {
		f.write(" ")
		f.writeKeyword("IF EXISTS")
	}
	f.write(" ")
	for i, t := range s.Tables {
		if i > 0 {
			f.write(", ")
		}
		f.Format(t)
	}
	if s.Cascade {
		f.write(" ")
		f.writeKeyword("CASCADE")
	}
}

func (f *Formatter) formatCreateIndex(s *ast.CreateIndexStmt) {
	f.writeKeyword("CREATE")
	if s.Unique {
		f.write(" ")
		f.writeKeyword("UNIQUE")
	}
	f.write(" ")
	f.writeKeyword("INDEX")
	if s.Concurrent {
		f.write(" ")
		f.writeKeyword("CONCURRENTLY")
	}
	if s.IfNotExists {
		f.write(" ")
		f.writeKeyword("IF NOT EXISTS")
	}
	if s.Name != "" {
		f.write(" ")
		f.writeIdent(s.Name)
	}
	f.write(" ")
	f.writeKeyword("ON")
	f.write(" ")
	f.Format(s.Table)
	if s.Using != "" {
		f.write(" ")
		f.writeKeyword("USING")
		f.write(" ")
		f.write(s.Using)
	}
	f.write(" (")
	for i, col := range s.Columns {
		if i > 0 {
			f.write(", ")
		}
		if col.Expr != nil {
			f.Format(col.Expr)
		} else {
			f.writeIdent(col.Column)
		}
		if col.Desc {
			f.write(" ")
			f.writeKeyword("DESC")
		}
	}
	f.write(")")
	if s.Where != nil {
		f.write(" ")
		f.writeKeyword("WHERE")
		f.write(" ")
		f.Format(s.Where)
	}
}

func (f *Formatter) formatDropIndex(s *ast.DropIndexStmt) {
	f.writeKeyword("DROP INDEX")
	if s.Concurrent {
		f.write(" ")
		f.writeKeyword("CONCURRENTLY")
	}
	if s.IfExists {
		f.write(" ")
		f.writeKeyword("IF EXISTS")
	}
	f.write(" ")
	f.writeIdent(s.Name)
	if s.Table != nil {
		f.write(" ")
		f.writeKeyword("ON")
		f.write(" ")
		f.Format(s.Table)
	}
	if s.Cascade {
		f.write(" ")
		f.writeKeyword("CASCADE")
	}
}

func (f *Formatter) formatTruncate(s *ast.TruncateStmt) {
	f.writeKeyword("TRUNCATE TABLE")
	f.write(" ")
	for i, t := range s.Tables {
		if i > 0 {
			f.write(", ")
		}
		f.Format(t)
	}
	if s.Cascade {
		f.write(" ")
		f.writeKeyword("CASCADE")
	}
}

func (f *Formatter) formatExplain(s *ast.ExplainStmt) {
	f.writeKeyword("EXPLAIN")
	if s.Analyze {
		f.write(" ")
		f.writeKeyword("ANALYZE")
	}
	if s.Verbose {
		f.write(" ")
		f.writeKeyword("VERBOSE")
	}
	if s.Format != "" {
		f.write(" ")
		f.writeKeyword("FORMAT")
		f.write(" ")
		f.write(s.Format)
	}
	f.write(" ")
	f.Format(s.Stmt)
}

func (f *Formatter) formatSetOp(s *ast.SetOp) {
	f.Format(s.Left)
	f.write(" ")
	switch s.Type {
	case ast.Union:
		f.writeKeyword("UNION")
	case ast.Intersect:
		f.writeKeyword("INTERSECT")
	case ast.Except:
		f.writeKeyword("EXCEPT")
	}
	if s.All {
		f.write(" ")
		f.writeKeyword("ALL")
	}
	f.write(" ")
	f.Format(s.Right)
}

func (f *Formatter) formatBinaryExpr(e *ast.BinaryExpr) {
	f.Format(e.Left)
	f.write(" ")
	f.writeKeyword(tokenToString(e.Op))
	f.write(" ")
	f.Format(e.Right)
}

func (f *Formatter) formatUnaryExpr(e *ast.UnaryExpr) {
	switch e.Op {
	case token.NOT:
		f.writeKeyword("NOT")
		f.write(" ")
	case token.MINUS:
		f.write("-")
		// Add space if operand is also unary minus to avoid -- comment syntax
		if inner, ok := e.Operand.(*ast.UnaryExpr); ok && inner.Op == token.MINUS {
			f.write(" ")
		}
	case token.BITNOT:
		f.write("~")
	}
	f.Format(e.Operand)
}

func (f *Formatter) formatFuncExpr(e *ast.FuncExpr) {
	f.writeFuncName(e.Name)
	f.write("(")
	if e.Distinct {
		f.writeKeyword("DISTINCT")
		f.write(" ")
	}
	for i, arg := range e.Args {
		if i > 0 {
			f.write(", ")
		}
		f.Format(arg)
	}
	f.write(")")
	if e.Filter != nil {
		f.write(" ")
		f.writeKeyword("FILTER")
		f.write(" (")
		f.writeKeyword("WHERE")
		f.write(" ")
		f.Format(e.Filter)
		f.write(")")
	}
	if e.Over != nil {
		f.write(" ")
		f.formatWindowSpec(e.Over)
	}
}

func (f *Formatter) formatWindowSpec(spec *ast.WindowSpec) {
	f.writeKeyword("OVER")
	f.write(" ")
	if spec.Name != "" && len(spec.PartitionBy) == 0 && len(spec.OrderBy) == 0 && spec.Frame == nil {
		f.writeIdent(spec.Name)
		return
	}
	f.write("(")
	if spec.Name != "" {
		f.writeIdent(spec.Name)
	}
	if len(spec.PartitionBy) > 0 {
		if spec.Name != "" {
			f.write(" ")
		}
		f.writeKeyword("PARTITION BY")
		f.write(" ")
		for i, pb := range spec.PartitionBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(pb)
		}
	}
	if len(spec.OrderBy) > 0 {
		if spec.Name != "" || len(spec.PartitionBy) > 0 {
			f.write(" ")
		}
		f.writeKeyword("ORDER BY")
		f.write(" ")
		for i, ob := range spec.OrderBy {
			if i > 0 {
				f.write(", ")
			}
			f.Format(ob.Expr)
			if ob.Desc {
				f.write(" ")
				f.writeKeyword("DESC")
			}
		}
	}
	if spec.Frame != nil {
		f.write(" ")
		f.formatWindowFrame(spec.Frame)
	}
	f.write(")")
}

func (f *Formatter) formatWindowFrame(frame *ast.WindowFrame) {
	switch frame.Type {
	case ast.FrameRows:
		f.writeKeyword("ROWS")
	case ast.FrameRange:
		f.writeKeyword("RANGE")
	case ast.FrameGroups:
		f.writeKeyword("GROUPS")
	}
	f.write(" ")
	if frame.End != nil {
		f.writeKeyword("BETWEEN")
		f.write(" ")
		f.formatFrameBound(frame.Start)
		f.write(" ")
		f.writeKeyword("AND")
		f.write(" ")
		f.formatFrameBound(frame.End)
	} else {
		f.formatFrameBound(frame.Start)
	}
}

func (f *Formatter) formatFrameBound(bound *ast.FrameBound) {
	switch bound.Type {
	case ast.BoundCurrentRow:
		f.writeKeyword("CURRENT ROW")
	case ast.BoundUnboundedPreceding:
		f.writeKeyword("UNBOUNDED PRECEDING")
	case ast.BoundUnboundedFollowing:
		f.writeKeyword("UNBOUNDED FOLLOWING")
	case ast.BoundPreceding:
		f.Format(bound.Offset)
		f.write(" ")
		f.writeKeyword("PRECEDING")
	case ast.BoundFollowing:
		f.Format(bound.Offset)
		f.write(" ")
		f.writeKeyword("FOLLOWING")
	}
}

func (f *Formatter) formatCaseExpr(e *ast.CaseExpr) {
	f.writeKeyword("CASE")
	if e.Operand != nil {
		f.write(" ")
		f.Format(e.Operand)
	}
	for _, w := range e.Whens {
		f.write(" ")
		f.writeKeyword("WHEN")
		f.write(" ")
		f.Format(w.Cond)
		f.write(" ")
		f.writeKeyword("THEN")
		f.write(" ")
		f.Format(w.Result)
	}
	if e.Else != nil {
		f.write(" ")
		f.writeKeyword("ELSE")
		f.write(" ")
		f.Format(e.Else)
	}
	f.write(" ")
	f.writeKeyword("END")
}

func (f *Formatter) formatCastExpr(e *ast.CastExpr) {
	f.writeKeyword("CAST")
	f.write("(")
	f.Format(e.Expr)
	f.write(" ")
	f.writeKeyword("AS")
	f.write(" ")
	f.formatDataType(e.Type)
	f.write(")")
}

func (f *Formatter) formatColName(c *ast.ColName) {
	for i, part := range c.Parts {
		if i > 0 {
			f.write(".")
		}
		f.writeIdent(part)
	}
}

func (f *Formatter) formatTableName(t *ast.TableName) {
	for i, part := range t.Parts {
		if i > 0 {
			f.write(".")
		}
		f.writeIdent(part)
	}
}

func (f *Formatter) formatLiteral(l *ast.Literal) {
	switch l.Type {
	case ast.LiteralNull:
		f.writeKeyword("NULL")
	case ast.LiteralString:
		f.formatStringLiteral(l.Value)
	case ast.LiteralBool:
		f.writeKeyword(l.Value)
	default:
		f.write(l.Value)
	}
}

func (f *Formatter) formatStringLiteral(s string) {
	// The lexer returns string content without enclosing quotes.
	// We need to add quotes and escape any internal quotes/backslashes.
	f.write("'")
	// Escape both single quotes and backslashes for round-trip safety
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "'", "''")
	f.write(escaped)
	f.write("'")
}

func (f *Formatter) formatParam(p *ast.Param) {
	switch p.Type {
	case ast.ParamQuestion:
		f.write("?")
	case ast.ParamDollar:
		f.write("$")
		f.write(itoa(p.Index))
	case ast.ParamColon:
		f.write(":")
		f.write(p.Name)
	case ast.ParamAt:
		f.write("@")
		f.write(p.Name)
	}
}

func (f *Formatter) formatAliasedTableExpr(a *ast.AliasedTableExpr) {
	f.Format(a.Expr)
	if a.Alias != "" {
		f.write(" ")
		f.writeKeyword("AS")
		f.write(" ")
		f.writeIdent(a.Alias)
	}
}

func (f *Formatter) formatJoinExpr(j *ast.JoinExpr) {
	f.Format(j.Left)
	f.write(" ")
	if j.Natural {
		f.writeKeyword("NATURAL")
		f.write(" ")
	}
	switch j.Type {
	case ast.JoinInner:
		f.writeKeyword("JOIN")
	case ast.JoinLeft:
		f.writeKeyword("LEFT JOIN")
	case ast.JoinRight:
		f.writeKeyword("RIGHT JOIN")
	case ast.JoinFull:
		f.writeKeyword("FULL JOIN")
	case ast.JoinCross:
		f.writeKeyword("CROSS JOIN")
	}
	f.write(" ")
	f.Format(j.Right)
	if j.On != nil {
		f.write(" ")
		f.writeKeyword("ON")
		f.write(" ")
		f.Format(j.On)
	}
	if len(j.Using) > 0 {
		f.write(" ")
		f.writeKeyword("USING")
		f.write(" (")
		for i, col := range j.Using {
			if i > 0 {
				f.write(", ")
			}
			f.writeIdent(col)
		}
		f.write(")")
	}
}

func (f *Formatter) formatInExpr(e *ast.InExpr) {
	f.Format(e.Expr)
	if e.Not {
		f.write(" ")
		f.writeKeyword("NOT")
	}
	f.write(" ")
	f.writeKeyword("IN")
	f.write(" (")
	if e.Select != nil {
		f.Format(e.Select)
	} else {
		for i, val := range e.Values {
			if i > 0 {
				f.write(", ")
			}
			f.Format(val)
		}
	}
	f.write(")")
}

func (f *Formatter) formatBetweenExpr(e *ast.BetweenExpr) {
	f.Format(e.Expr)
	if e.Not {
		f.write(" ")
		f.writeKeyword("NOT")
	}
	f.write(" ")
	f.writeKeyword("BETWEEN")
	f.write(" ")
	f.Format(e.Low)
	f.write(" ")
	f.writeKeyword("AND")
	f.write(" ")
	f.Format(e.High)
}

func (f *Formatter) formatLikeExpr(e *ast.LikeExpr) {
	f.Format(e.Expr)
	if e.Not {
		f.write(" ")
		f.writeKeyword("NOT")
	}
	f.write(" ")
	if e.ILike {
		f.writeKeyword("ILIKE")
	} else {
		f.writeKeyword("LIKE")
	}
	f.write(" ")
	f.Format(e.Pattern)
	if e.Escape != nil {
		f.write(" ")
		f.writeKeyword("ESCAPE")
		f.write(" ")
		f.Format(e.Escape)
	}
}

func (f *Formatter) formatIsExpr(e *ast.IsExpr) {
	f.Format(e.Expr)
	f.write(" ")
	f.writeKeyword("IS")
	if e.Not {
		f.write(" ")
		f.writeKeyword("NOT")
	}
	f.write(" ")
	switch e.What {
	case ast.IsNull:
		f.writeKeyword("NULL")
	case ast.IsTrue:
		f.writeKeyword("TRUE")
	case ast.IsFalse:
		f.writeKeyword("FALSE")
	case ast.IsUnknown:
		f.writeKeyword("UNKNOWN")
	}
}

func (f *Formatter) formatExistsExpr(e *ast.ExistsExpr) {
	if e.Not {
		f.writeKeyword("NOT")
		f.write(" ")
	}
	f.writeKeyword("EXISTS")
	f.write(" ")
	f.Format(e.Subquery)
}

func (f *Formatter) formatIntervalExpr(e *ast.IntervalExpr) {
	f.writeKeyword("INTERVAL")
	f.write(" ")
	f.Format(e.Value)
	if e.Unit != "" {
		f.write(" ")
		f.writeKeyword(e.Unit)
	}
}

func (f *Formatter) formatExtractExpr(e *ast.ExtractExpr) {
	f.writeKeyword("EXTRACT")
	f.write("(")
	// Use writeIdent to handle empty or special field names
	f.writeIdent(e.Field)
	f.write(" ")
	f.writeKeyword("FROM")
	f.write(" ")
	f.Format(e.Source)
	f.write(")")
}

func (f *Formatter) formatTrimExpr(e *ast.TrimExpr) {
	f.writeKeyword("TRIM")
	f.write("(")
	switch e.TrimType {
	case ast.TrimLeading:
		f.writeKeyword("LEADING")
		f.write(" ")
	case ast.TrimTrailing:
		f.writeKeyword("TRAILING")
		f.write(" ")
	case ast.TrimBoth:
		f.writeKeyword("BOTH")
		f.write(" ")
	}
	if e.TrimChar != nil {
		f.Format(e.TrimChar)
		f.write(" ")
	}
	f.writeKeyword("FROM")
	f.write(" ")
	f.Format(e.Expr)
	f.write(")")
}

func (f *Formatter) formatSubstringExpr(e *ast.SubstringExpr) {
	f.writeKeyword("SUBSTRING")
	f.write("(")
	f.Format(e.Expr)
	if e.From != nil {
		f.write(" ")
		f.writeKeyword("FROM")
		f.write(" ")
		f.Format(e.From)
	}
	if e.For != nil {
		f.write(" ")
		f.writeKeyword("FOR")
		f.write(" ")
		f.Format(e.For)
	}
	f.write(")")
}

func (f *Formatter) formatArrayExpr(e *ast.ArrayExpr) {
	// Format ARRAY constructor with spaces inside brackets to distinguish from
	// SQL Server bracket identifiers. The lexer treats [ followed by space as
	// LBRACKET, not as start of bracket identifier.
	f.writeKeyword("ARRAY")
	f.write("[ ")
	for i, elem := range e.Elements {
		if i > 0 {
			f.write(", ")
		}
		f.Format(elem)
	}
	f.write(" ]")
}

func (f *Formatter) formatValuesStmt(s *ast.ValuesStmt) {
	f.writeKeyword("VALUES")
	f.write(" ")
	for i, row := range s.Rows {
		if i > 0 {
			f.write(", ")
		}
		f.write("(")
		for j, val := range row {
			if j > 0 {
				f.write(", ")
			}
			f.Format(val)
		}
		f.write(")")
	}
}

func needsQuoting(id string) bool {
	if needsQuotingNonKeyword(id) {
		return true
	}
	// Check if it's a reserved keyword
	return token.IsKeyword(id)
}

// needsQuotingNonKeyword checks if an identifier needs quoting for non-keyword
// reasons (empty, special characters, etc.)
func needsQuotingNonKeyword(id string) bool {
	if len(id) == 0 {
		return true
	}
	// Check first char
	ch := id[0]
	if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
		return true
	}
	// Check remaining chars
	for i := 1; i < len(id); i++ {
		ch := id[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '$') {
			return true
		}
	}
	return false
}

func tokenToString(t token.Token) string {
	switch t {
	case token.EQ:
		return "="
	case token.NEQ:
		return "<>"
	case token.LT:
		return "<"
	case token.GT:
		return ">"
	case token.LTE:
		return "<="
	case token.GTE:
		return ">="
	case token.PLUS:
		return "+"
	case token.MINUS:
		return "-"
	case token.ASTERISK:
		return "*"
	case token.SLASH:
		return "/"
	case token.PERCENT:
		return "%"
	case token.AND:
		return "AND"
	case token.OR:
		return "OR"
	case token.XOR:
		return "XOR"
	case token.CONCAT:
		return "||"
	case token.BITAND:
		return "&"
	case token.BITOR:
		return "|"
	case token.BITXOR:
		return "^"
	case token.LSHIFT:
		return "<<"
	case token.RSHIFT:
		return ">>"
	default:
		return t.String()
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
