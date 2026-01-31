package parser

import (
	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/token"
)

func (p *Parser) parseInsert() *ast.InsertStmt {
	pos := p.cur.Pos
	stmt := &ast.InsertStmt{StartPos: pos}

	if p.curIs(token.REPLACE) {
		stmt.Replace = true
	}
	p.advance() // consume INSERT or REPLACE

	// IGNORE (MySQL)
	if p.curIs(token.IGNORE) {
		stmt.Ignore = true
		p.advance()
	}

	if !p.expect(token.INTO) {
		return nil
	}

	stmt.Table = p.parseTableName()

	// Optional column list
	if p.curIs(token.LPAREN) && !p.peekIs(token.SELECT) {
		p.advance()
		for {
			if !p.curIs(token.IDENT) {
				break
			}
			col := &ast.ColName{
				StartPos: p.cur.Pos,
				EndPos:   p.cur.Pos,
				Parts:    []string{p.cur.Value},
			}
			stmt.Columns = append(stmt.Columns, col)
			p.advance()

			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
		p.expect(token.RPAREN)
	}

	// VALUES, SELECT, or SET
	if p.curIs(token.VALUES) || p.curIs(token.VALUE) {
		p.advance()
		stmt.Values = p.parseValuesList()
	} else if p.curIs(token.SELECT) || p.curIs(token.WITH) {
		var innerStmt ast.Statement
		if p.curIs(token.WITH) {
			innerStmt = p.parseWith()
		} else {
			innerStmt = p.parseSelect()
		}
		if innerStmt == nil {
			return nil
		}
		sel, ok := innerStmt.(*ast.SelectStmt)
		if !ok {
			p.errorf("expected SELECT statement")
			return nil
		}
		stmt.Select = sel
	} else if p.curIs(token.SET) {
		// MySQL INSERT ... SET syntax: INSERT INTO t SET col1=val1, col2=val2
		p.advance()
		// Must have at least one column=value assignment
		if !p.curIsIdent() {
			p.errorf("expected column name after SET")
			return nil
		}
		stmt.Values = [][]ast.Expr{{}}
		for {
			if !p.curIsIdent() {
				break
			}
			// Find column index
			colName := p.curIdentValue()
			p.advance()
			if !p.expect(token.EQ) {
				return nil
			}
			val := p.parseExpr()
			if val == nil {
				return nil
			}

			// Add column and value
			stmt.Columns = append(stmt.Columns, &ast.ColName{Parts: []string{colName}})
			stmt.Values[0] = append(stmt.Values[0], val)

			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
	} else if p.curIs(token.DEFAULT) {
		p.advance()
		p.expect(token.VALUES)
		// INSERT ... DEFAULT VALUES
		stmt.Values = [][]ast.Expr{{}}
	}

	// ON DUPLICATE KEY UPDATE (MySQL)
	if p.curIs(token.ON) {
		p.advance()
		if p.curIs(token.DUPLICATE) {
			p.advance()
			p.expect(token.KEY)
			p.expect(token.UPDATE)
			stmt.OnDuplicateUpdate = p.parseUpdateExprs()
		}
	}

	// ON CONFLICT (PostgreSQL)
	if p.curIs(token.CONFLICT) || (p.curIs(token.ON) && p.peekIs(token.CONFLICT)) {
		if p.curIs(token.ON) {
			p.advance()
		}
		stmt.OnConflict = p.parseOnConflict()
	}

	// RETURNING (PostgreSQL)
	if p.curIs(token.RETURNING) {
		p.advance()
		stmt.Returning = p.parseSelectExprs()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseValuesList() [][]ast.Expr {
	var rows [][]ast.Expr

	for {
		if !p.curIs(token.LPAREN) {
			break
		}
		p.advance()

		var row []ast.Expr
		for {
			if p.curIs(token.DEFAULT) {
				row = append(row, &ast.Literal{
					StartPos: p.cur.Pos,
					EndPos:   p.cur.Pos,
					Type:     ast.LiteralNull,
					Value:    "DEFAULT",
				})
				p.advance()
			} else {
				expr := p.parseExpr()
				if expr == nil {
					break
				}
				row = append(row, expr)
			}

			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}

		rows = append(rows, row)
		p.expect(token.RPAREN)

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	return rows
}

func (p *Parser) parseOnConflict() *ast.OnConflict {
	p.advance() // consume CONFLICT

	conflict := &ast.OnConflict{}

	// Target columns
	if p.curIs(token.LPAREN) {
		p.advance()
		for {
			if p.curIs(token.IDENT) {
				conflict.Columns = append(conflict.Columns, p.cur.Value)
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

	// WHERE clause for partial index
	if p.curIs(token.WHERE) {
		p.advance()
		conflict.Where = p.parseExpr()
	}

	// DO NOTHING or DO UPDATE
	p.expect(token.DO)
	if p.curIs(token.NOTHING) {
		conflict.DoNothing = true
		p.advance()
	} else if p.curIs(token.UPDATE) {
		p.advance()
		p.expect(token.SET)
		conflict.Updates = p.parseUpdateExprs()
	}

	return conflict
}

func (p *Parser) parseUpdate() *ast.UpdateStmt {
	pos := p.cur.Pos
	p.advance() // consume UPDATE

	stmt := &ast.UpdateStmt{StartPos: pos}

	// Table reference
	stmt.Table = p.parseTableExpr()

	// SET clause
	if !p.expect(token.SET) {
		return nil
	}
	stmt.Set = p.parseUpdateExprs()

	// FROM clause (PostgreSQL)
	if p.curIs(token.FROM) {
		p.advance()
		stmt.From = p.parseTableExpr()
	}

	// WHERE clause
	if p.curIs(token.WHERE) {
		p.advance()
		stmt.Where = p.parseExpr()
	}

	// ORDER BY (MySQL)
	if p.curIs(token.ORDER) {
		stmt.OrderBy = p.parseOrderBy()
	}

	// LIMIT (MySQL)
	if p.curIs(token.LIMIT) {
		stmt.Limit = p.parseLimit()
	}

	// RETURNING (PostgreSQL)
	if p.curIs(token.RETURNING) {
		p.advance()
		stmt.Returning = p.parseSelectExprs()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}

func (p *Parser) parseUpdateExprs() []*ast.UpdateExpr {
	var exprs []*ast.UpdateExpr

	for {
		if !p.curIs(token.IDENT) {
			break
		}

		startPos := p.cur.Pos
		parts := []string{p.cur.Value}
		p.advance()

		// Check for qualified column name (table.column or schema.table.column)
		for p.curIs(token.DOT) {
			p.advance()
			if p.curIs(token.IDENT) {
				parts = append(parts, p.cur.Value)
				p.advance()
			} else {
				break
			}
		}

		ue := &ast.UpdateExpr{
			Column: &ast.ColName{
				StartPos: startPos,
				EndPos:   p.cur.Pos,
				Parts:    parts,
			},
		}

		p.expect(token.EQ)
		ue.Expr = p.parseExpr()

		exprs = append(exprs, ue)

		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}

	return exprs
}

func (p *Parser) parseDelete() *ast.DeleteStmt {
	pos := p.cur.Pos
	p.advance() // consume DELETE

	stmt := &ast.DeleteStmt{StartPos: pos}

	// Optional FROM
	if p.curIs(token.FROM) {
		p.advance()
	}

	stmt.Table = p.parseTableExpr()

	// USING clause (PostgreSQL)
	if p.curIs(token.USING) {
		p.advance()
		stmt.Using = p.parseTableExpr()
	}

	// WHERE clause
	if p.curIs(token.WHERE) {
		p.advance()
		stmt.Where = p.parseExpr()
	}

	// ORDER BY (MySQL)
	if p.curIs(token.ORDER) {
		stmt.OrderBy = p.parseOrderBy()
	}

	// LIMIT (MySQL)
	if p.curIs(token.LIMIT) {
		stmt.Limit = p.parseLimit()
	}

	// RETURNING (PostgreSQL)
	if p.curIs(token.RETURNING) {
		p.advance()
		stmt.Returning = p.parseSelectExprs()
	}

	stmt.EndPos = p.cur.Pos
	return stmt
}
