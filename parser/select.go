package parser

import (
	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/token"
)

func (p *Parser) parseSelect() *ast.SelectStmt {
	pos := p.cur.Pos
	if !p.expect(token.SELECT) {
		return nil
	}

	stmt := ast.GetSelectStmt()
	stmt.StartPos = pos

	// Skip hints like SQL_CALC_FOUND_ROWS
	for p.curIs(token.SQL_CALC_FOUND_ROWS) || p.curIs(token.SQL_SMALL_RESULT) ||
		p.curIs(token.SQL_BIG_RESULT) || p.curIs(token.SQL_BUFFER_RESULT) ||
		p.curIs(token.HIGH_PRIORITY) || p.curIs(token.STRAIGHT_JOIN) {
		p.advance()
	}

	// Check for DISTINCT/ALL
	if p.curIs(token.DISTINCT) {
		stmt.Distinct = true
		p.advance()
	} else if p.curIs(token.ALL) {
		p.advance()
	}

	// Parse select expressions
	stmt.Columns = p.parseSelectExprs()

	// Optional INTO clause (MySQL)
	if p.curIs(token.INTO) {
		stmt.Into = p.parseSelectInto()
	}

	// FROM clause (optional for things like SELECT 1+1)
	if p.curIs(token.FROM) {
		p.advance()
		stmt.From = p.parseTableExpr()
	}

	// WHERE clause
	if p.curIs(token.WHERE) {
		p.advance()
		stmt.Where = p.parseExpr()
	}

	// GROUP BY clause
	if p.curIs(token.GROUP) {
		p.advance()
		if !p.expect(token.BY) {
			return nil
		}
		stmt.GroupBy = p.parseExprList()
	}

	// HAVING clause
	if p.curIs(token.HAVING) {
		p.advance()
		stmt.Having = p.parseExpr()
	}

	// WINDOW clause
	if p.curIs(token.WINDOW) {
		stmt.WindowDefs = p.parseWindowDefs()
	}

	// ORDER BY clause
	if p.curIs(token.ORDER) {
		stmt.OrderBy = p.parseOrderBy()
	}

	// LIMIT clause
	if p.curIs(token.LIMIT) {
		stmt.Limit = p.parseLimit()
	}

	// OFFSET clause (PostgreSQL style without LIMIT)
	if p.curIs(token.OFFSET) && stmt.Limit == nil {
		stmt.Limit = &ast.Limit{StartPos: p.cur.Pos}
		p.advance()
		stmt.Limit.Offset = p.parseExpr()
		stmt.Limit.EndPos = p.cur.Pos
	}

	// FETCH clause (SQL standard)
	if p.curIs(token.FETCH) {
		if stmt.Limit == nil {
			stmt.Limit = &ast.Limit{StartPos: p.cur.Pos}
		}
		p.advance()
		if p.curIs(token.FIRST) || p.curIs(token.NEXT) {
			p.advance()
		}
		stmt.Limit.Count = p.parseExpr()
		if p.curIs(token.ROW) || p.curIs(token.ROWS) {
			p.advance()
		}
		if p.curIs(token.ONLY) {
			p.advance()
		}
		stmt.Limit.EndPos = p.cur.Pos
	}

	// FOR UPDATE/SHARE
	if p.curIs(token.FOR) {
		stmt.Lock = p.parseLockClause()
	}

	stmt.EndPos = p.cur.Pos

	// Check for set operations (UNION, INTERSECT, EXCEPT)
	if p.curIs(token.UNION) || p.curIs(token.INTERSECT) || p.curIs(token.EXCEPT) {
		return p.parseSetOp(stmt)
	}

	return stmt
}

func (p *Parser) parseSelectExprs() []ast.SelectExpr {
	// Get slice from pool (pre-allocated with typical capacity)
	slicePtr := ast.GetSelectExprSlice()
	exprs := *slicePtr
	for {
		expr := p.parseSelectExpr()
		if expr == nil {
			break
		}
		exprs = append(exprs, expr)
		if !p.curIs(token.COMMA) {
			break
		}
		p.advance() // consume comma
	}
	return exprs
}

func (p *Parser) parseSelectExpr() ast.SelectExpr {
	// Skip any comments
	p.skipComments()
	pos := p.cur.Pos

	// Check for *
	if p.curIs(token.ASTERISK) {
		p.advance()
		return &ast.StarExpr{StartPos: pos, EndPos: pos}
	}

	// Parse as expression with optional alias
	expr := p.parseExpr()
	if expr == nil {
		return nil
	}

	// Check if it's table.* (parseExpr returns StarExpr for this)
	if _, ok := expr.(*ast.StarExpr); ok {
		return expr.(*ast.StarExpr)
	}

	alias := ""
	if p.curIs(token.AS) {
		p.advance()
		if !p.curIs(token.IDENT) && !p.curIs(token.STRING) {
			p.errorf("expected alias after AS")
			return nil
		}
		alias = p.cur.Value
		p.advance()
	} else if p.curIs(token.IDENT) {
		// Check if this looks like an alias (not a keyword that starts a clause)
		if !isClauseKeyword(p.cur.Type) {
			alias = p.cur.Value
			p.advance()
		}
	}

	ae := ast.GetAliasedExpr()
	ae.StartPos = pos
	ae.EndPos = p.cur.Pos
	ae.Expr = expr
	ae.Alias = alias
	return ae
}

func (p *Parser) parseSelectInto() *ast.SelectInto {
	p.advance() // consume INTO

	into := &ast.SelectInto{}

	if p.curIs(token.OUTFILE) {
		p.advance()
		if p.curIs(token.STRING) {
			into.Outfile = p.cur.Value
			p.advance()
		}
	} else if p.curIs(token.IDENT) && p.cur.Value == "DUMPFILE" {
		p.advance()
		if p.curIs(token.STRING) {
			into.Dumpfile = p.cur.Value
			p.advance()
		}
	} else {
		// Variable list
		for {
			if p.curIs(token.PARAM) || p.curIs(token.IDENT) {
				into.Vars = append(into.Vars, p.cur.Value)
				p.advance()
			} else {
				break
			}
			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
	}

	return into
}

func (p *Parser) parseTableExpr() ast.TableExpr {
	left := p.parseTablePrimary()
	if left == nil {
		return nil
	}

	// Parse joins
	for {
		joinType, natural, hasJoin := p.checkJoinKeyword()
		if !hasJoin {
			break
		}

		join := ast.GetJoinExpr()
		join.StartPos = p.cur.Pos
		join.Type = joinType
		join.Natural = natural
		join.Left = left

		// Consume join keywords
		p.consumeJoinKeywords()

		// Check for LATERAL
		if p.curIs(token.LATERAL) {
			join.Lateral = true
			p.advance()
		}

		join.Right = p.parseTablePrimary()

		// ON or USING clause (not for CROSS JOIN or NATURAL JOIN)
		if joinType != ast.JoinCross && !natural {
			if p.curIs(token.ON) {
				p.advance()
				join.On = p.parseExpr()
			} else if p.curIs(token.USING) {
				p.advance()
				join.Using = p.parseColumnNameList()
			}
		}

		join.EndPos = p.cur.Pos
		left = join
	}

	return left
}

func (p *Parser) parseTablePrimary() ast.TableExpr {
	var expr ast.TableExpr

	// Check for LATERAL
	lateral := false
	if p.curIs(token.LATERAL) {
		lateral = true
		p.advance()
	}

	if p.curIs(token.LPAREN) {
		pos := p.cur.Pos
		p.advance()
		if p.curIs(token.SELECT) || p.curIs(token.WITH) {
			// Derived table (subquery)
			var stmt ast.Statement
			if p.curIs(token.WITH) {
				stmt = p.parseWith()
			} else {
				stmt = p.parseSelect()
			}
			if stmt == nil {
				return nil
			}
			if !p.expect(token.RPAREN) {
				return nil
			}
			sel, ok := stmt.(*ast.SelectStmt)
			if !ok {
				p.errorf("expected SELECT statement in subquery")
				return nil
			}
			expr = &ast.Subquery{StartPos: pos, EndPos: p.cur.Pos, Select: sel}
		} else {
			// Parenthesized table expression
			inner := p.parseTableExpr()
			if !p.expect(token.RPAREN) {
				return nil
			}
			expr = &ast.ParenTableExpr{StartPos: pos, EndPos: p.cur.Pos, Expr: inner}
		}
	} else if p.curIsIdent() {
		tn := p.parseTableName()
		if tn == nil {
			return nil
		}
		expr = tn
	} else if p.curIs(token.VALUES) {
		expr = p.parseValuesClause()
	} else {
		p.errorf("expected table name or subquery")
		return nil
	}

	// Parse optional alias
	alias := ""
	if p.curIs(token.AS) {
		p.advance()
	}
	if p.curIs(token.IDENT) && !isClauseKeyword(p.cur.Type) {
		alias = p.cur.Value
		p.advance()
	}

	// Parse column alias list for derived tables
	var colAliases []string
	if p.curIs(token.LPAREN) {
		colAliases = p.parseColumnNameList()
	}
	_ = colAliases // Would store in AliasedTableExpr if needed

	// Parse index hints (MySQL)
	var hints []*ast.IndexHint
	for p.curIs(token.USE) || p.curIs(token.FORCE) || p.curIs(token.IGNORE) {
		hints = append(hints, p.parseIndexHint())
	}

	if alias != "" || len(hints) > 0 || lateral {
		aliased := ast.GetAliasedTableExpr()
		aliased.StartPos = expr.Pos()
		aliased.EndPos = p.cur.Pos
		aliased.Expr = expr
		aliased.Alias = alias
		aliased.Hints = hints
		if lateral {
			if join, ok := expr.(*ast.JoinExpr); ok {
				join.Lateral = true
			}
		}
		return aliased
	}

	return expr
}

func (p *Parser) parseValuesClause() *ast.ValuesStmt {
	pos := p.cur.Pos
	p.advance() // consume VALUES

	stmt := &ast.ValuesStmt{StartPos: pos}

	for {
		if !p.expect(token.LPAREN) {
			break
		}

		var row []ast.Expr
		for {
			expr := p.parseExpr()
			if expr == nil {
				break
			}
			row = append(row, expr)
			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
		stmt.Rows = append(stmt.Rows, row)

		if !p.expect(token.RPAREN) {
			break
		}

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseIndexHint() *ast.IndexHint {
	hint := &ast.IndexHint{}

	switch p.cur.Type {
	case token.USE:
		hint.Type = ast.HintUse
	case token.FORCE:
		hint.Type = ast.HintForce
	case token.IGNORE:
		hint.Type = ast.HintIgnore
	}
	p.advance()

	// INDEX or KEY
	if p.curIs(token.INDEX) || p.curIs(token.KEY) {
		p.advance()
	}

	// FOR clause
	if p.curIs(token.FOR) {
		p.advance()
		switch p.cur.Type {
		case token.JOIN:
			hint.For = ast.HintForJoin
			p.advance()
		case token.ORDER:
			hint.For = ast.HintForOrderBy
			p.advance()
			p.expect(token.BY)
		case token.GROUP:
			hint.For = ast.HintForGroupBy
			p.advance()
			p.expect(token.BY)
		}
	}

	// Index list
	if p.curIs(token.LPAREN) {
		p.advance()
		for {
			if p.curIs(token.IDENT) || p.curIs(token.PRIMARY) {
				hint.Indexes = append(hint.Indexes, p.cur.Value)
				p.advance()
			} else {
				break
			}
			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
		p.expect(token.RPAREN)
	}

	return hint
}

func (p *Parser) parseOrderBy() []*ast.OrderByExpr {
	p.advance() // consume ORDER
	if !p.expect(token.BY) {
		return nil
	}

	slicePtr := ast.GetOrderBySlice()
	items := *slicePtr
	for {
		pos := p.cur.Pos
		expr := p.parseExpr()
		if expr == nil {
			break
		}

		item := ast.GetOrderByExpr()
		item.StartPos = pos
		item.Expr = expr

		if p.curIs(token.ASC) {
			p.advance()
		} else if p.curIs(token.DESC) {
			item.Desc = true
			p.advance()
		}

		// NULLS FIRST/LAST
		if p.curIs(token.NULLS) {
			p.advance()
			if p.curIs(token.FIRST) {
				t := true
				item.NullsFirst = &t
				p.advance()
			} else if p.curIs(token.LAST) {
				f := false
				item.NullsFirst = &f
				p.advance()
			}
		}

		item.EndPos = p.cur.Pos
		items = append(items, item)

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	return items
}

func (p *Parser) parseLimit() *ast.Limit {
	pos := p.cur.Pos
	p.advance() // consume LIMIT

	limit := &ast.Limit{StartPos: pos}

	// MySQL style: LIMIT count [OFFSET offset] or LIMIT offset, count
	limit.Count = p.parseExpr()

	if p.curIs(token.OFFSET) {
		p.advance()
		limit.Offset = p.parseExpr()
	} else if p.curIs(token.COMMA) {
		// MySQL: LIMIT offset, count
		p.advance()
		limit.Offset = limit.Count
		limit.Count = p.parseExpr()
	}

	limit.EndPos = p.cur.Pos
	return limit
}

func (p *Parser) parseLockClause() string {
	p.advance() // consume FOR

	var lock string
	if p.curIs(token.UPDATE) {
		lock = "UPDATE"
		p.advance()
	} else if p.curIs(token.SHARE) {
		lock = "SHARE"
		p.advance()
	}

	// NOWAIT, SKIP LOCKED
	if p.curIs(token.NOWAIT) {
		lock += " NOWAIT"
		p.advance()
	} else if p.curIs(token.SKIP) {
		p.advance()
		if p.curIs(token.LOCKED) {
			lock += " SKIP LOCKED"
			p.advance()
		}
	}

	return lock
}

func (p *Parser) parseWindowDefs() []*ast.WindowDef {
	p.advance() // consume WINDOW

	var defs []*ast.WindowDef
	for {
		if !p.curIs(token.IDENT) {
			break
		}

		def := &ast.WindowDef{Name: p.cur.Value}
		p.advance()

		if !p.expect(token.AS) {
			break
		}

		def.Spec = p.parseWindowSpec()
		defs = append(defs, def)

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	return defs
}

func (p *Parser) parseSetOp(left *ast.SelectStmt) *ast.SelectStmt {
	// This is a simplification - properly handling set operations would
	// return a SetOp node, but for now we'll just parse the right side
	// and return the left
	for p.curIs(token.UNION) || p.curIs(token.INTERSECT) || p.curIs(token.EXCEPT) {
		p.advance()
		if p.curIs(token.ALL) || p.curIs(token.DISTINCT) {
			p.advance()
		}
		// Parse right side - can be parenthesized or direct SELECT
		if p.curIs(token.LPAREN) {
			p.advance() // consume '('
			p.parseSelect()
			p.expect(token.RPAREN)
		} else {
			p.parseSelect()
		}
	}
	return left
}

func (p *Parser) checkJoinKeyword() (ast.JoinType, bool, bool) {
	natural := false
	if p.curIs(token.NATURAL) {
		natural = true
	}

	switch p.cur.Type {
	case token.JOIN:
		return ast.JoinInner, natural, true
	case token.INNER:
		return ast.JoinInner, natural, true
	case token.LEFT:
		return ast.JoinLeft, natural, true
	case token.RIGHT:
		return ast.JoinRight, natural, true
	case token.FULL:
		return ast.JoinFull, natural, true
	case token.CROSS:
		return ast.JoinCross, natural, true
	case token.NATURAL:
		return ast.JoinInner, true, true
	case token.STRAIGHT_JOIN:
		return ast.JoinInner, false, true
	case token.COMMA:
		// Comma is an implicit cross join
		return ast.JoinCross, false, true
	default:
		return 0, false, false
	}
}

func (p *Parser) consumeJoinKeywords() {
	// Consume join type keywords (including comma for implicit cross join)
	for p.curIs(token.NATURAL) || p.curIs(token.INNER) || p.curIs(token.LEFT) ||
		p.curIs(token.RIGHT) || p.curIs(token.FULL) || p.curIs(token.OUTER) ||
		p.curIs(token.CROSS) || p.curIs(token.JOIN) || p.curIs(token.STRAIGHT_JOIN) ||
		p.curIs(token.COMMA) {
		p.advance()
	}
}

func isClauseKeyword(t token.Token) bool {
	switch t {
	case token.FROM, token.WHERE, token.GROUP, token.HAVING, token.ORDER,
		token.LIMIT, token.OFFSET, token.UNION, token.INTERSECT, token.EXCEPT,
		token.FOR, token.INTO, token.ON, token.USING, token.JOIN, token.INNER,
		token.LEFT, token.RIGHT, token.FULL, token.CROSS, token.NATURAL,
		token.AND, token.OR, token.THEN, token.ELSE, token.END, token.WHEN,
		token.AS, token.SET, token.VALUES, token.RETURNING:
		return true
	default:
		return false
	}
}
