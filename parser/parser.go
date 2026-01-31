// Package parser provides a recursive descent SQL parser.
package parser

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/lexer"
	"github.com/freeeve/machparse/token"
)

// Parser is a recursive descent SQL parser.
type Parser struct {
	lexer  *lexer.Lexer
	errors []ParseError
	cur    token.Item // current token
}

// ParseError represents a parse error with position.
type ParseError struct {
	Pos     token.Pos
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s", e.Pos.Line, e.Pos.Column, e.Message)
}

// New creates a new parser for the given input.
func New(input string) *Parser {
	p := &Parser{
		lexer: lexer.New(input),
	}
	p.advance() // Prime the first token
	return p
}

var parserPool = sync.Pool{
	New: func() any { return &Parser{} },
}

// Get returns a parser from the pool for the given input.
// Call Put(p) when done to return it to the pool.
func Get(input string) *Parser {
	p := parserPool.Get().(*Parser)
	p.lexer = lexer.Get(input)
	p.errors = p.errors[:0]
	p.cur = token.Item{}
	p.advance()
	return p
}

// Put returns the parser and its lexer to the pool.
func Put(p *Parser) {
	if p.lexer != nil {
		lexer.Put(p.lexer)
		p.lexer = nil
	}
	parserPool.Put(p)
}

// Parse parses a single statement.
func (p *Parser) Parse() (ast.Statement, error) {
	p.skipComments()
	if p.curIs(token.EOF) {
		return nil, nil
	}
	stmt := p.parseStatement()
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}
	// Verify all input was consumed (allow trailing semicolons and comments)
	p.skipComments()
	for p.curIs(token.SEMICOLON) {
		p.advance()
		p.skipComments()
	}
	if !p.curIs(token.EOF) {
		p.errorf("unexpected token %v after statement", p.cur.Type)
		return nil, p.errors[0]
	}
	return stmt, nil
}

// ParseAll parses all statements until EOF.
func (p *Parser) ParseAll() ([]ast.Statement, error) {
	var stmts []ast.Statement
	for !p.curIs(token.EOF) {
		p.skipComments()
		if p.curIs(token.EOF) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		// Skip optional semicolons between statements
		for p.curIs(token.SEMICOLON) {
			p.advance()
		}
		p.skipComments()
	}
	if len(p.errors) > 0 {
		return stmts, p.errors[0]
	}
	return stmts, nil
}

// Token navigation methods

func (p *Parser) advance() {
	p.cur = p.lexer.Next()
}

func (p *Parser) curIs(t token.Token) bool {
	return p.cur.Type == t
}

// curIsIdent returns true if the current token can be used as an identifier.
// This includes both IDENT tokens and keywords (which can be used as identifiers
// in certain contexts like table/column names).
func (p *Parser) curIsIdent() bool {
	return p.cur.Type == token.IDENT || p.cur.Type.IsKeyword()
}

// curIdentValue returns the identifier value of the current token.
// Works for both IDENT tokens and keywords used as identifiers.
func (p *Parser) curIdentValue() string {
	return p.cur.Value
}

func (p *Parser) curIsKeyword(keywords ...token.Token) bool {
	for _, kw := range keywords {
		if p.cur.Type == kw {
			return true
		}
	}
	return false
}

func (p *Parser) peek() token.Item {
	return p.lexer.Peek()
}

func (p *Parser) peekIs(t token.Token) bool {
	return p.peek().Type == t
}

func (p *Parser) expect(t token.Token) bool {
	if p.curIs(t) {
		p.advance()
		return true
	}
	p.errorf("expected %v, got %v", t, p.cur.Type)
	return false
}

func (p *Parser) skipComments() {
	for p.curIs(token.COMMENT) {
		p.advance()
	}
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.errors = append(p.errors, ParseError{
		Pos:     p.cur.Pos,
		Message: fmt.Sprintf(format, args...),
	})
}

// parseStatement dispatches to the appropriate statement parser.
func (p *Parser) parseStatement() ast.Statement {
	switch p.cur.Type {
	case token.SELECT:
		return p.parseSelect()
	case token.INSERT, token.REPLACE:
		return p.parseInsert()
	case token.UPDATE:
		return p.parseUpdate()
	case token.DELETE:
		return p.parseDelete()
	case token.CREATE:
		return p.parseCreate()
	case token.ALTER:
		return p.parseAlter()
	case token.DROP:
		return p.parseDrop()
	case token.WITH:
		return p.parseWith()
	case token.TRUNCATE:
		return p.parseTruncate()
	case token.EXPLAIN, token.ANALYZE:
		return p.parseExplain()
	case token.LPAREN:
		return p.parseParenthesizedStatement()
	default:
		p.errorf("unexpected token %v at start of statement", p.cur.Type)
		p.advance() // Skip to recover
		return nil
	}
}

// parseWith handles WITH clause (CTEs).
func (p *Parser) parseWith() ast.Statement {
	withClause := p.parseWithClause()

	p.skipComments()
	switch p.cur.Type {
	case token.SELECT:
		stmt := p.parseSelect()
		if stmt != nil {
			stmt.With = withClause
		}
		return stmt
	case token.INSERT, token.REPLACE:
		stmt := p.parseInsert()
		if stmt != nil {
			stmt.With = withClause
		}
		return stmt
	case token.UPDATE:
		stmt := p.parseUpdate()
		if stmt != nil {
			stmt.With = withClause
		}
		return stmt
	case token.DELETE:
		stmt := p.parseDelete()
		if stmt != nil {
			stmt.With = withClause
		}
		return stmt
	default:
		p.errorf("expected SELECT, INSERT, UPDATE, or DELETE after WITH")
		return nil
	}
}

func (p *Parser) parseWithClause() *ast.WithClause {
	p.advance() // consume WITH

	with := &ast.WithClause{}

	if p.curIs(token.RECURSIVE) {
		with.Recursive = true
		p.advance()
	}

	for {
		cte := p.parseCTE()
		if cte != nil {
			with.CTEs = append(with.CTEs, cte)
		}

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance() // consume comma
	}

	return with
}

func (p *Parser) parseCTE() *ast.CTE {
	if !p.curIs(token.IDENT) {
		p.errorf("expected CTE name")
		return nil
	}

	cte := &ast.CTE{
		Name: p.cur.Value,
	}
	p.advance()

	// Optional column list
	if p.curIs(token.LPAREN) {
		cte.Columns = p.parseColumnNameList()
	}

	if !p.expect(token.AS) {
		return nil
	}

	if !p.expect(token.LPAREN) {
		return nil
	}

	cte.Query = p.parseStatement()

	if !p.expect(token.RPAREN) {
		return nil
	}

	return cte
}

func (p *Parser) parseColumnNameList() []string {
	p.advance() // consume (

	var names []string
	for {
		if !p.curIs(token.IDENT) {
			break
		}
		names = append(names, p.cur.Value)
		p.advance()

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance() // consume comma
	}

	p.expect(token.RPAREN)
	return names
}

// Placeholder implementations for statements we'll complete later
func (p *Parser) parseCreate() ast.Statement {
	pos := p.cur.Pos
	p.advance() // consume CREATE

	// Skip TEMPORARY/TEMP
	if p.curIs(token.TEMPORARY) || p.curIs(token.TEMP) {
		p.advance()
	}

	switch p.cur.Type {
	case token.TABLE:
		return p.parseCreateTable(pos)
	case token.INDEX, token.UNIQUE:
		return p.parseCreateIndex(pos)
	default:
		p.errorf("expected TABLE or INDEX after CREATE")
		return nil
	}
}

func (p *Parser) parseCreateTable(pos token.Pos) ast.Statement {
	p.advance() // consume TABLE

	stmt := &ast.CreateTableStmt{StartPos: pos}

	if p.curIs(token.IF) {
		p.advance()
		if p.curIs(token.NOT) {
			p.advance()
			if p.curIs(token.EXISTS) {
				stmt.IfNotExists = true
				p.advance()
			}
		}
	}

	stmt.Table = p.parseTableName()

	// Check for CREATE TABLE AS SELECT
	if p.curIs(token.AS) {
		p.advance()
		stmt.As = p.parseSelect()
		stmt.EndPos = p.cur.Pos
		return stmt
	}

	if !p.expect(token.LPAREN) {
		return nil
	}

	// Parse column definitions and table constraints
	for !p.curIs(token.RPAREN) && !p.curIs(token.EOF) {
		if p.curIs(token.PRIMARY) || p.curIs(token.FOREIGN) ||
			p.curIs(token.UNIQUE) || p.curIs(token.CHECK) || p.curIs(token.CONSTRAINT) {
			constraint := p.parseTableConstraint()
			if constraint != nil {
				stmt.Constraints = append(stmt.Constraints, constraint)
			}
		} else {
			col := p.parseColumnDef()
			if col != nil {
				stmt.Columns = append(stmt.Columns, col)
			}
		}

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	p.expect(token.RPAREN)

	// Parse table options (ENGINE, CHARSET, etc.)
	stmt.Options = p.parseTableOptions()

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseColumnDef() *ast.ColumnDef {
	if !p.curIs(token.IDENT) {
		p.errorf("expected column name")
		return nil
	}

	col := &ast.ColumnDef{
		Name: p.cur.Value,
	}
	p.advance()

	col.Type = p.parseDataType()
	col.Constraints = p.parseColumnConstraints()

	return col
}

func (p *Parser) parseDataType() *ast.DataType {
	dt := &ast.DataType{}

	// Get base type name
	if p.cur.Type.IsKeyword() || p.curIs(token.IDENT) {
		dt.Name = p.cur.Value
		p.advance()
	} else {
		p.errorf("expected data type")
		return dt
	}

	// Handle multi-word types like DOUBLE PRECISION, CHARACTER VARYING
	if p.curIs(token.PRECISION) || p.curIs(token.VARYING) {
		dt.Name += " " + p.cur.Value
		p.advance()
	}

	// Parse length/precision
	if p.curIs(token.LPAREN) {
		p.advance()
		if p.curIs(token.INT) {
			n := parseInt(p.cur.Value)
			dt.Length = &n
			p.advance()

			if p.curIs(token.COMMA) {
				p.advance()
				if p.curIs(token.INT) {
					s := parseInt(p.cur.Value)
					dt.Precision = dt.Length
					dt.Scale = &s
					p.advance()
				}
			}
		}
		p.expect(token.RPAREN)
	}

	// Parse modifiers
	for {
		switch p.cur.Type {
		case token.UNSIGNED:
			dt.Unsigned = true
			p.advance()
		case token.SIGNED:
			p.advance()
		case token.ZEROFILL:
			p.advance()
		case token.CHARACTER, token.CHAR:
			if p.peekIs(token.SET) || p.peekIs(token.CHARSET) {
				p.advance()
				p.advance()
				if p.curIs(token.IDENT) || p.curIs(token.STRING) {
					dt.Charset = p.cur.Value
					p.advance()
				}
			} else {
				return dt
			}
		case token.COLLATE:
			p.advance()
			if p.curIs(token.IDENT) || p.curIs(token.STRING) {
				dt.Collation = p.cur.Value
				p.advance()
			}
		case token.ARRAY:
			dt.Array = true
			p.advance()
		case token.LBRACKET:
			// PostgreSQL array syntax: type[]
			p.advance()
			p.expect(token.RBRACKET)
			dt.Array = true
		default:
			return dt
		}
	}
}

func (p *Parser) parseColumnConstraints() []*ast.ColumnConstraint {
	var constraints []*ast.ColumnConstraint

	for {
		var constraint *ast.ColumnConstraint

		// Optional CONSTRAINT name
		name := ""
		if p.curIs(token.CONSTRAINT) {
			p.advance()
			if p.curIs(token.IDENT) {
				name = p.cur.Value
				p.advance()
			}
		}

		switch p.cur.Type {
		case token.NOT:
			p.advance()
			if p.curIs(token.NULL) {
				p.advance()
				constraint = &ast.ColumnConstraint{
					Name:    name,
					Type:    ast.ConstraintNotNull,
					NotNull: true,
				}
			}
		case token.NULL:
			p.advance()
			// NULL is the default, no constraint needed
		case token.PRIMARY:
			p.advance()
			p.expect(token.KEY)
			constraint = &ast.ColumnConstraint{
				Name: name,
				Type: ast.ConstraintPrimaryKey,
			}
		case token.UNIQUE:
			p.advance()
			constraint = &ast.ColumnConstraint{
				Name: name,
				Type: ast.ConstraintUnique,
			}
		case token.DEFAULT:
			p.advance()
			constraint = &ast.ColumnConstraint{
				Name:    name,
				Type:    ast.ConstraintDefault,
				Default: p.parseExpr(),
			}
		case token.CHECK:
			p.advance()
			p.expect(token.LPAREN)
			constraint = &ast.ColumnConstraint{
				Name:  name,
				Type:  ast.ConstraintCheck,
				Check: p.parseExpr(),
			}
			p.expect(token.RPAREN)
		case token.REFERENCES:
			p.advance()
			constraint = &ast.ColumnConstraint{
				Name:       name,
				Type:       ast.ConstraintForeignKey,
				References: p.parseForeignKeyRef(),
			}
		case token.AUTO_INCREMENT, token.AUTOINCREMENT:
			p.advance()
			// MySQL/SQLite auto increment - treated as column property
		case token.GENERATED:
			p.advance()
			constraint = p.parseGeneratedConstraint(name)
		default:
			return constraints
		}

		if constraint != nil {
			constraints = append(constraints, constraint)
		}
	}
}

func (p *Parser) parseGeneratedConstraint(name string) *ast.ColumnConstraint {
	gen := &ast.GeneratedColumn{}

	// GENERATED ALWAYS AS (expr) [STORED | VIRTUAL]
	if p.curIs(token.ALWAYS) {
		p.advance()
	}

	if p.curIs(token.AS) {
		p.advance()
	}

	p.expect(token.LPAREN)
	gen.Expr = p.parseExpr()
	p.expect(token.RPAREN)

	if p.curIs(token.STORED) {
		gen.Stored = true
		p.advance()
	} else if p.curIs(token.VIRTUAL) {
		p.advance()
	}

	return &ast.ColumnConstraint{
		Name:      name,
		Type:      ast.ConstraintGenerated,
		Generated: gen,
	}
}

func (p *Parser) parseForeignKeyRef() *ast.ForeignKeyRef {
	ref := &ast.ForeignKeyRef{
		Table: p.parseTableName(),
	}

	if p.curIs(token.LPAREN) {
		ref.Columns = p.parseColumnNameList()
	}

	// ON DELETE / ON UPDATE
	for p.curIs(token.ON) {
		p.advance()
		var action *ast.RefAction
		switch p.cur.Type {
		case token.DELETE:
			p.advance()
			a := p.parseRefAction()
			ref.OnDelete = a
			action = &ref.OnDelete
		case token.UPDATE:
			p.advance()
			a := p.parseRefAction()
			ref.OnUpdate = a
			action = &ref.OnUpdate
		}
		_ = action
	}

	return ref
}

func (p *Parser) parseRefAction() ast.RefAction {
	switch p.cur.Type {
	case token.CASCADE:
		p.advance()
		return ast.RefCascade
	case token.RESTRICT:
		p.advance()
		return ast.RefRestrict
	case token.SET:
		p.advance()
		if p.curIs(token.NULL) {
			p.advance()
			return ast.RefSetNull
		} else if p.curIs(token.DEFAULT) {
			p.advance()
			return ast.RefSetDefault
		}
	case token.NO:
		p.advance()
		p.expect(token.ACTION)
		return ast.RefNoAction
	}
	return ast.RefNoAction
}

func (p *Parser) parseTableConstraint() *ast.TableConstraint {
	tc := &ast.TableConstraint{}

	// Optional CONSTRAINT name
	if p.curIs(token.CONSTRAINT) {
		p.advance()
		if p.curIs(token.IDENT) {
			tc.Name = p.cur.Value
			p.advance()
		}
	}

	switch p.cur.Type {
	case token.PRIMARY:
		p.advance()
		p.expect(token.KEY)
		tc.Type = ast.ConstraintPrimaryKey
		if p.curIs(token.LPAREN) {
			tc.Columns = p.parseColumnNameList()
		}
	case token.UNIQUE:
		p.advance()
		tc.Type = ast.ConstraintUnique
		if p.curIs(token.KEY) {
			p.advance()
		}
		if p.curIs(token.LPAREN) {
			tc.Columns = p.parseColumnNameList()
		}
	case token.FOREIGN:
		p.advance()
		p.expect(token.KEY)
		tc.Type = ast.ConstraintForeignKey
		if p.curIs(token.LPAREN) {
			tc.Columns = p.parseColumnNameList()
		}
		p.expect(token.REFERENCES)
		tc.References = p.parseForeignKeyRef()
	case token.CHECK:
		p.advance()
		tc.Type = ast.ConstraintCheck
		p.expect(token.LPAREN)
		tc.Check = p.parseExpr()
		p.expect(token.RPAREN)
	}

	return tc
}

func (p *Parser) parseTableOptions() []*ast.TableOption {
	var opts []*ast.TableOption

	for {
		switch p.cur.Type {
		case token.ENGINE:
			p.advance()
			if p.curIs(token.EQ) {
				p.advance()
			}
			if p.curIs(token.IDENT) {
				opts = append(opts, &ast.TableOption{Name: "ENGINE", Value: p.cur.Value})
				p.advance()
			}
		case token.CHARSET, token.CHARACTER:
			p.advance()
			if p.curIs(token.SET) {
				p.advance()
			}
			if p.curIs(token.EQ) {
				p.advance()
			}
			if p.curIs(token.IDENT) {
				opts = append(opts, &ast.TableOption{Name: "CHARSET", Value: p.cur.Value})
				p.advance()
			}
		case token.COLLATE:
			p.advance()
			if p.curIs(token.EQ) {
				p.advance()
			}
			if p.curIs(token.IDENT) {
				opts = append(opts, &ast.TableOption{Name: "COLLATE", Value: p.cur.Value})
				p.advance()
			}
		case token.COMMENT_KW:
			p.advance()
			if p.curIs(token.EQ) {
				p.advance()
			}
			if p.curIs(token.STRING) {
				opts = append(opts, &ast.TableOption{Name: "COMMENT", Value: p.cur.Value})
				p.advance()
			}
		case token.AUTO_INCREMENT:
			p.advance()
			if p.curIs(token.EQ) {
				p.advance()
			}
			if p.curIs(token.INT) {
				opts = append(opts, &ast.TableOption{Name: "AUTO_INCREMENT", Value: p.cur.Value})
				p.advance()
			}
		default:
			return opts
		}
	}
}

func (p *Parser) parseCreateIndex(pos token.Pos) ast.Statement {
	stmt := &ast.CreateIndexStmt{StartPos: pos}

	if p.curIs(token.UNIQUE) {
		stmt.Unique = true
		p.advance()
	}

	p.expect(token.INDEX)

	if p.curIs(token.CONCURRENTLY) {
		stmt.Concurrent = true
		p.advance()
	}

	if p.curIs(token.IF) {
		p.advance()
		if p.curIs(token.NOT) {
			p.advance()
			if p.curIs(token.EXISTS) {
				stmt.IfNotExists = true
				p.advance()
			}
		}
	}

	if p.curIs(token.IDENT) {
		stmt.Name = p.cur.Value
		p.advance()
	}

	p.expect(token.ON)
	stmt.Table = p.parseTableName()

	// USING method
	if p.curIs(token.USING) {
		p.advance()
		if p.curIs(token.IDENT) {
			stmt.Using = p.cur.Value
			p.advance()
		}
	}

	// Column list
	p.expect(token.LPAREN)
	for !p.curIs(token.RPAREN) && !p.curIs(token.EOF) {
		col := &ast.IndexColumn{}
		if p.curIsIdent() {
			col.Column = p.curIdentValue()
			p.advance()
		} else if p.curIs(token.LPAREN) {
			// Expression index (must be parenthesized)
			col.Expr = p.parseExpr()
		} else {
			p.errorf("expected column name or expression")
			return nil
		}

		if p.curIs(token.DESC) {
			col.Desc = true
			p.advance()
		} else if p.curIs(token.ASC) {
			p.advance()
		}

		if p.curIs(token.NULLS) {
			p.advance()
			if p.curIs(token.FIRST) {
				col.Nulls = "FIRST"
				p.advance()
			} else if p.curIs(token.LAST) {
				col.Nulls = "LAST"
				p.advance()
			}
		}

		stmt.Columns = append(stmt.Columns, col)

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}
	p.expect(token.RPAREN)

	// WHERE clause for partial index
	if p.curIs(token.WHERE) {
		p.advance()
		stmt.Where = p.parseExpr()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseAlter() ast.Statement {
	pos := p.cur.Pos
	p.advance() // consume ALTER

	if !p.curIs(token.TABLE) {
		p.errorf("expected TABLE after ALTER")
		return nil
	}
	p.advance()

	stmt := &ast.AlterTableStmt{
		StartPos: pos,
		Table:    p.parseTableName(),
	}

	// Parse alter actions
	for {
		action := p.parseAlterTableAction()
		if action != nil {
			stmt.Actions = append(stmt.Actions, action)
		}

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseAlterTableAction() ast.AlterTableAction {
	switch p.cur.Type {
	case token.ADD:
		p.advance()
		if p.curIs(token.COLUMN) {
			p.advance()
		}
		if p.curIs(token.CONSTRAINT) || p.curIs(token.PRIMARY) ||
			p.curIs(token.FOREIGN) || p.curIs(token.UNIQUE) || p.curIs(token.CHECK) {
			return &ast.AddConstraint{Constraint: p.parseTableConstraint()}
		}
		return &ast.AddColumn{Column: p.parseColumnDef()}

	case token.DROP:
		p.advance()
		if p.curIs(token.COLUMN) {
			p.advance()
			action := &ast.DropColumn{}
			if p.curIs(token.IF) {
				p.advance()
				p.expect(token.EXISTS)
				action.IfExists = true
			}
			if p.curIsIdent() {
				action.Name = p.curIdentValue()
				p.advance()
			}
			if p.curIs(token.CASCADE) {
				action.Cascade = true
				p.advance()
			}
			return action
		}
		if p.curIs(token.CONSTRAINT) {
			p.advance()
			action := &ast.DropConstraint{}
			if p.curIs(token.IF) {
				p.advance()
				p.expect(token.EXISTS)
				action.IfExists = true
			}
			if p.curIsIdent() {
				action.Name = p.curIdentValue()
				p.advance()
			}
			if p.curIs(token.CASCADE) {
				action.Cascade = true
				p.advance()
			}
			return action
		}

	case token.RENAME:
		p.advance()
		if p.curIs(token.COLUMN) {
			p.advance()
			action := &ast.RenameColumn{}
			if p.curIsIdent() {
				action.OldName = p.curIdentValue()
				p.advance()
			}
			p.expect(token.TO)
			if p.curIsIdent() {
				action.NewName = p.curIdentValue()
				p.advance()
			}
			return action
		}
		if p.curIs(token.TO) {
			p.advance()
			return &ast.RenameTable{NewName: p.parseTableName()}
		}

	case token.MODIFY, token.ALTER:
		p.advance()
		if p.curIs(token.COLUMN) {
			p.advance()
		}
		action := &ast.ModifyColumn{}
		if p.curIsIdent() {
			action.Name = p.curIdentValue()
			p.advance()
		}
		// Various modifications
		if p.curIs(token.SET) {
			p.advance()
			if p.curIs(token.NOT) {
				p.advance()
				p.expect(token.NULL)
				action.SetNotNull = true
			} else if p.curIs(token.DEFAULT) {
				p.advance()
				action.SetDefault = p.parseExpr()
			}
		} else if p.curIs(token.DROP) {
			p.advance()
			if p.curIs(token.NOT) {
				p.advance()
				p.expect(token.NULL)
				action.DropNotNull = true
			} else if p.curIs(token.DEFAULT) {
				p.advance()
				action.DropDefault = true
			}
		} else {
			// MySQL MODIFY COLUMN name type - parse type and constraints directly
			colDef := &ast.ColumnDef{Name: action.Name}
			colDef.Type = p.parseDataType()
			colDef.Constraints = p.parseColumnConstraints()
			action.NewDef = colDef
		}
		return action
	}

	return nil
}

func (p *Parser) parseDrop() ast.Statement {
	pos := p.cur.Pos
	p.advance() // consume DROP

	switch p.cur.Type {
	case token.TABLE:
		return p.parseDropTable(pos)
	case token.INDEX:
		return p.parseDropIndex(pos)
	default:
		p.errorf("expected TABLE or INDEX after DROP")
		return nil
	}
}

func (p *Parser) parseDropTable(pos token.Pos) ast.Statement {
	p.advance() // consume TABLE

	stmt := &ast.DropTableStmt{StartPos: pos}

	if p.curIs(token.IF) {
		p.advance()
		p.expect(token.EXISTS)
		stmt.IfExists = true
	}

	// Parse table names
	for {
		stmt.Tables = append(stmt.Tables, p.parseTableName())
		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	if p.curIs(token.CASCADE) {
		stmt.Cascade = true
		p.advance()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseDropIndex(pos token.Pos) ast.Statement {
	p.advance() // consume INDEX

	stmt := &ast.DropIndexStmt{StartPos: pos}

	if p.curIs(token.CONCURRENTLY) {
		stmt.Concurrent = true
		p.advance()
	}

	if p.curIs(token.IF) {
		p.advance()
		p.expect(token.EXISTS)
		stmt.IfExists = true
	}

	if p.curIs(token.IDENT) {
		stmt.Name = p.cur.Value
		p.advance()
	}

	// MySQL: DROP INDEX name ON table
	if p.curIs(token.ON) {
		p.advance()
		stmt.Table = p.parseTableName()
	}

	if p.curIs(token.CASCADE) {
		stmt.Cascade = true
		p.advance()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseTruncate() ast.Statement {
	pos := p.cur.Pos
	p.advance() // consume TRUNCATE

	if p.curIs(token.TABLE) {
		p.advance()
	}

	stmt := &ast.TruncateStmt{StartPos: pos}

	// Parse table names
	for {
		stmt.Tables = append(stmt.Tables, p.parseTableName())
		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	if p.curIs(token.CASCADE) {
		stmt.Cascade = true
		p.advance()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

// parseParenthesizedStatement handles statements that start with parentheses,
// like (SELECT ...) UNION (SELECT ...).
func (p *Parser) parseParenthesizedStatement() ast.Statement {
	pos := p.cur.Pos
	p.advance() // consume '('

	// Parse inner statement
	inner := p.parseStatement()
	if inner == nil {
		return nil
	}

	if !p.expect(token.RPAREN) {
		return nil
	}

	// Only SELECT can be in parentheses for set operations
	sel, ok := inner.(*ast.SelectStmt)
	if !ok {
		return inner
	}

	// Check for set operations (UNION, INTERSECT, EXCEPT)
	if p.curIs(token.UNION) || p.curIs(token.INTERSECT) || p.curIs(token.EXCEPT) {
		return p.parseSetOp(sel)
	}

	// Check for ORDER BY / LIMIT on parenthesized select
	if p.curIs(token.ORDER) {
		sel.OrderBy = p.parseOrderBy()
	}
	if p.curIs(token.LIMIT) {
		sel.Limit = p.parseLimit()
	}
	sel.StartPos = pos
	sel.EndPos = p.cur.Pos

	return sel
}

func (p *Parser) parseExplain() ast.Statement {
	pos := p.cur.Pos

	stmt := &ast.ExplainStmt{StartPos: pos}

	if p.curIs(token.EXPLAIN) {
		p.advance()
	}

	// Parse options
	for {
		switch p.cur.Type {
		case token.ANALYZE:
			stmt.Analyze = true
			p.advance()
		case token.VERBOSE:
			stmt.Verbose = true
			p.advance()
		case token.FORMAT:
			p.advance()
			if p.curIs(token.IDENT) {
				stmt.Format = p.cur.Value
				p.advance()
			}
		case token.LPAREN:
			// PostgreSQL style: EXPLAIN (ANALYZE, VERBOSE, ...)
			p.advance()
			for !p.curIs(token.RPAREN) && !p.curIs(token.EOF) {
				switch p.cur.Type {
				case token.ANALYZE:
					stmt.Analyze = true
				case token.VERBOSE:
					stmt.Verbose = true
				case token.FORMAT:
					p.advance()
					if p.curIs(token.IDENT) {
						stmt.Format = p.cur.Value
					}
				}
				p.advance()
				if p.curIs(token.COMMA) {
					p.advance()
				}
			}
			p.expect(token.RPAREN)
		default:
			goto parseStmt
		}
	}

parseStmt:
	stmt.Stmt = p.parseStatement()
	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseTableName() *ast.TableName {
	if !p.curIsIdent() {
		p.errorf("expected table name")
		return nil
	}

	pos := p.cur.Pos
	parts := []string{p.curIdentValue()}
	p.advance()

	// Collect all parts (catalog.schema.table)
	for p.curIs(token.DOT) {
		p.advance()
		if !p.curIsIdent() {
			p.errorf("expected identifier after '.'")
			return nil
		}
		parts = append(parts, p.curIdentValue())
		p.advance()
	}

	tn := ast.GetTableName()
	tn.StartPos = pos
	tn.EndPos = p.cur.Pos
	tn.Parts = parts
	return tn
}

func parseInt(s string) int {
	// Use strconv to properly handle overflow
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// On overflow or error, return max int to avoid negative values
		return int(^uint(0) >> 1)
	}
	// Clamp to int range
	if n > int64(int(^uint(0)>>1)) {
		return int(^uint(0) >> 1)
	}
	if n < int64(-int(^uint(0)>>1)-1) {
		return -int(^uint(0)>>1) - 1
	}
	return int(n)
}
