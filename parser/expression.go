package parser

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/freeeve/machparse/ast"
	"github.com/freeeve/machparse/token"
)

// isNilExpr checks if an expression is nil, handling typed nils.
func isNilExpr(e ast.Expr) bool {
	if e == nil {
		return true
	}
	v := reflect.ValueOf(e)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// Operator precedence levels (higher = tighter binding)
const (
	precLowest     = 0
	precOr         = 1  // OR
	precXor        = 2  // XOR
	precAnd        = 3  // AND
	precNot        = 4  // NOT (prefix)
	precComparison = 5  // =, <>, <, >, <=, >=, IS, LIKE, IN, BETWEEN
	precBitOr      = 6  // |
	precBitXor     = 7  // ^
	precBitAnd     = 8  // &
	precShift      = 9  // <<, >>
	precAdditive   = 10 // +, -, ||
	precMultiply   = 11 // *, /, %
	precUnary      = 12 // -, ~, !
	precCollate    = 13 // COLLATE
	precHighest    = 14
)

// precedence returns the precedence of a binary operator.
func precedence(t token.Token) int {
	switch t {
	case token.OR:
		return precOr
	case token.XOR:
		return precXor
	case token.AND:
		return precAnd
	case token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE:
		return precComparison
	case token.BITOR:
		return precBitOr
	case token.BITXOR:
		return precBitXor
	case token.BITAND:
		return precBitAnd
	case token.LSHIFT, token.RSHIFT:
		return precShift
	case token.PLUS, token.MINUS, token.CONCAT:
		return precAdditive
	case token.ASTERISK, token.SLASH, token.PERCENT:
		return precMultiply
	default:
		return precLowest
	}
}

// parseExpr parses an expression using precedence climbing.
func (p *Parser) parseExpr() ast.Expr {
	return p.parseExprPrec(precLowest)
}

// parseExprPrec implements precedence climbing.
func (p *Parser) parseExprPrec(minPrec int) ast.Expr {
	left := p.parsePrimaryExpr()
	if left == nil {
		return nil
	}

	for {
		op := p.cur.Type

		// Handle special cases that aren't simple binary ops
		if p.curIs(token.IS) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseIsExpr(left)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.IN) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseInExpr(left, false)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.NOT) {
			next := p.peek()
			switch next.Type {
			case token.IN:
				if isNilExpr(left) {
					return nil
				}
				p.advance() // consume NOT
				left = p.parseInExpr(left, true)
				if isNilExpr(left) {
					return nil
				}
				continue
			case token.BETWEEN:
				if isNilExpr(left) {
					return nil
				}
				p.advance() // consume NOT
				left = p.parseBetweenExpr(left, true)
				if isNilExpr(left) {
					return nil
				}
				continue
			case token.LIKE, token.ILIKE:
				if isNilExpr(left) {
					return nil
				}
				p.advance() // consume NOT
				left = p.parseLikeExpr(left, true)
				if isNilExpr(left) {
					return nil
				}
				continue
			case token.SIMILAR:
				if isNilExpr(left) {
					return nil
				}
				p.advance() // consume NOT
				left = p.parseSimilarExpr(left, true)
				if isNilExpr(left) {
					return nil
				}
				continue
			}
		}
		if p.curIs(token.BETWEEN) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseBetweenExpr(left, false)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.LIKE) || p.curIs(token.ILIKE) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseLikeExpr(left, false)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.SIMILAR) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseSimilarExpr(left, false)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.COLLATE) {
			if isNilExpr(left) {
				return nil
			}
			left = p.parseCollateExpr(left)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.DCOLON) {
			// PostgreSQL cast: expr::type
			if isNilExpr(left) {
				return nil
			}
			left = p.parsePostgresCast(left)
			if isNilExpr(left) {
				return nil
			}
			continue
		}
		if p.curIs(token.LBRACKET) {
			// Array subscript
			if isNilExpr(left) {
				return nil
			}
			left = p.parseSubscript(left)
			if isNilExpr(left) {
				return nil
			}
			continue
		}

		// Standard binary operators
		prec := precedence(op)
		if prec < minPrec {
			break
		}
		if !isBinaryOp(op) {
			break
		}

		pos := p.cur.Pos
		p.advance() // consume operator

		right := p.parseExprPrec(prec + 1)
		if right == nil {
			return nil
		}

		bin := ast.GetBinaryExpr()
		bin.StartPos = pos
		bin.Op = op
		bin.Left = left
		bin.Right = right
		left = bin
	}

	return left
}

// parsePrimaryExpr parses primary expressions (atoms and prefix operators).
func (p *Parser) parsePrimaryExpr() ast.Expr {
	// Skip any comments before the expression
	p.skipComments()

	switch p.cur.Type {
	case token.INT:
		return p.parseLiteral(ast.LiteralInt)
	case token.FLOAT:
		return p.parseLiteral(ast.LiteralFloat)
	case token.STRING:
		return p.parseLiteral(ast.LiteralString)
	case token.NULL:
		pos := p.cur.Pos
		p.advance()
		return &ast.Literal{StartPos: pos, EndPos: pos, Type: ast.LiteralNull, Value: "NULL"}
	case token.TRUE:
		pos := p.cur.Pos
		p.advance()
		return &ast.Literal{StartPos: pos, EndPos: pos, Type: ast.LiteralBool, Value: "TRUE"}
	case token.FALSE:
		pos := p.cur.Pos
		p.advance()
		return &ast.Literal{StartPos: pos, EndPos: pos, Type: ast.LiteralBool, Value: "FALSE"}
	case token.IDENT:
		return p.parseIdentifierOrFunc()
	case token.PARAM:
		return p.parseParam()
	case token.LPAREN:
		return p.parseParenOrSubquery()
	case token.NOT:
		return p.parseNotExpr()
	case token.MINUS:
		return p.parseUnaryMinus()
	case token.BITNOT:
		return p.parseUnaryBitnot()
	case token.EXISTS:
		return p.parseExistsExpr()
	case token.CASE:
		return p.parseCaseExpr()
	case token.CAST:
		return p.parseCastExpr()
	case token.INTERVAL:
		return p.parseIntervalExpr()
	case token.EXTRACT:
		return p.parseExtractExpr()
	case token.TRIM:
		return p.parseTrimExpr()
	case token.SUBSTRING:
		return p.parseSubstringExpr()
	case token.POSITION:
		return p.parsePositionExpr()
	case token.ASTERISK:
		pos := p.cur.Pos
		p.advance()
		return &ast.StarExpr{StartPos: pos, EndPos: pos}
	case token.ARRAY:
		return p.parseArrayExpr()
	case token.DEFAULT:
		pos := p.cur.Pos
		p.advance()
		return &ast.Literal{StartPos: pos, EndPos: pos, Type: ast.LiteralNull, Value: "DEFAULT"}
	default:
		// Check if it's a keyword that could be a function name or column name
		if p.cur.Type.IsKeyword() {
			return p.parseIdentifierOrFunc()
		}
		p.errorf("unexpected token %v in expression", p.cur.Type)
		return nil
	}
}

func (p *Parser) parseLiteral(litType ast.LiteralType) *ast.Literal {
	lit := ast.GetLiteral()
	lit.StartPos = p.cur.Pos
	lit.EndPos = p.cur.Pos
	lit.Type = litType
	lit.Value = p.cur.Value
	p.advance()
	return lit
}

func (p *Parser) parseIdentifierOrFunc() ast.Expr {
	pos := p.cur.Pos
	name := p.cur.Value
	p.advance()

	// Check for function call first (before checking for dots)
	if p.curIs(token.LPAREN) {
		return p.parseFuncCall(pos, name)
	}

	// Collect all parts of a qualified identifier (a.b.c.d...)
	parts := []string{name}
	var endPos token.Pos = pos

	for p.curIs(token.DOT) {
		p.advance()

		// Check for table.* (qualified star)
		if p.curIs(token.ASTERISK) {
			endPos = p.cur.Pos
			p.advance()
			// Join all parts except the star for the table name
			tableName := parts[len(parts)-1]
			if len(parts) > 1 {
				tableName = parts[len(parts)-1]
			}
			return &ast.StarExpr{
				StartPos:     pos,
				EndPos:       endPos,
				TableName:    tableName,
				HasQualifier: true,
			}
		}

		if !p.curIs(token.IDENT) && !p.cur.Type.IsKeyword() {
			p.errorf("expected identifier after '.'")
			return nil
		}

		parts = append(parts, p.cur.Value)
		endPos = p.cur.Pos
		p.advance()
	}

	// Build ColName with all parts
	col := ast.GetColName()
	col.StartPos = pos
	col.EndPos = endPos
	col.Parts = parts
	return col
}

func (p *Parser) parseFuncCall(pos token.Pos, name string) *ast.FuncExpr {
	p.advance() // consume '('

	fn := ast.GetFuncExpr()
	fn.StartPos = pos
	fn.Name = strings.ToUpper(name)
	// Get args slice from pool
	if fn.Args == nil {
		slicePtr := ast.GetExprSlice()
		fn.Args = *slicePtr
	}

	// Check for COUNT(*) or DISTINCT
	if p.curIs(token.DISTINCT) {
		fn.Distinct = true
		p.advance()
	}

	// Parse arguments
	if !p.curIs(token.RPAREN) {
		if p.curIs(token.ASTERISK) {
			// COUNT(*)
			fn.Args = append(fn.Args, &ast.StarExpr{StartPos: p.cur.Pos, EndPos: p.cur.Pos})
			p.advance()
		} else {
			for {
				arg := p.parseExpr()
				if arg == nil {
					break
				}
				fn.Args = append(fn.Args, arg)
				if !p.curIs(token.COMMA) {
					break
				}
				p.advance() // consume comma
			}
		}
	}

	if !p.expect(token.RPAREN) {
		return nil
	}
	fn.EndPos = p.cur.Pos

	// Check for FILTER clause
	if p.curIs(token.FILTER) {
		p.advance()
		if !p.expect(token.LPAREN) {
			return nil
		}
		if !p.expect(token.WHERE) {
			return nil
		}
		fn.Filter = p.parseExpr()
		if !p.expect(token.RPAREN) {
			return nil
		}
	}

	// Check for OVER clause (window function)
	if p.curIs(token.OVER) {
		fn.Over = p.parseWindowSpec()
	}

	return fn
}

func (p *Parser) parseWindowSpec() *ast.WindowSpec {
	p.advance() // consume OVER
	pos := p.cur.Pos

	spec := &ast.WindowSpec{StartPos: pos}

	// Could be OVER window_name or OVER (...)
	if p.curIs(token.IDENT) {
		spec.Name = p.cur.Value
		p.advance()
		spec.EndPos = p.cur.Pos
		return spec
	}

	if !p.expect(token.LPAREN) {
		return nil
	}

	// Optional base window name
	if p.curIs(token.IDENT) && !p.peekIs(token.BY) {
		spec.Name = p.cur.Value
		p.advance()
	}

	// PARTITION BY
	if p.curIs(token.PARTITION) {
		p.advance()
		p.expect(token.BY)
		spec.PartitionBy = p.parseExprList()
	}

	// ORDER BY
	if p.curIs(token.ORDER) {
		spec.OrderBy = p.parseOrderBy()
	}

	// Frame clause
	if p.curIs(token.ROWS) || p.curIs(token.RANGE) || p.curIs(token.GROUPS) {
		spec.Frame = p.parseWindowFrame()
	}

	p.expect(token.RPAREN)
	spec.EndPos = p.cur.Pos
	return spec
}

func (p *Parser) parseWindowFrame() *ast.WindowFrame {
	frame := &ast.WindowFrame{}

	switch p.cur.Type {
	case token.ROWS:
		frame.Type = ast.FrameRows
	case token.RANGE:
		frame.Type = ast.FrameRange
	case token.GROUPS:
		frame.Type = ast.FrameGroups
	}
	p.advance()

	if p.curIs(token.BETWEEN) {
		p.advance()
		frame.Start = p.parseFrameBound()
		p.expect(token.AND)
		frame.End = p.parseFrameBound()
	} else {
		frame.Start = p.parseFrameBound()
	}

	return frame
}

func (p *Parser) parseFrameBound() *ast.FrameBound {
	bound := &ast.FrameBound{}

	if p.curIs(token.CURRENT) {
		p.advance()
		p.expect(token.ROW)
		bound.Type = ast.BoundCurrentRow
	} else if p.curIs(token.UNBOUNDED) {
		p.advance()
		if p.curIs(token.PRECEDING) {
			p.advance()
			bound.Type = ast.BoundUnboundedPreceding
		} else if p.curIs(token.FOLLOWING) {
			p.advance()
			bound.Type = ast.BoundUnboundedFollowing
		}
	} else {
		bound.Offset = p.parseExpr()
		if p.curIs(token.PRECEDING) {
			p.advance()
			bound.Type = ast.BoundPreceding
		} else if p.curIs(token.FOLLOWING) {
			p.advance()
			bound.Type = ast.BoundFollowing
		}
	}

	return bound
}

func (p *Parser) parseParam() *ast.Param {
	param := &ast.Param{
		StartPos: p.cur.Pos,
		EndPos:   p.cur.Pos,
	}

	val := p.cur.Value
	switch {
	case val == "?":
		param.Type = ast.ParamQuestion
	case strings.HasPrefix(val, "$"):
		param.Type = ast.ParamDollar
		if len(val) > 1 {
			param.Index, _ = strconv.Atoi(val[1:])
		}
	case strings.HasPrefix(val, ":"):
		param.Type = ast.ParamColon
		param.Name = val[1:]
	case strings.HasPrefix(val, "@"):
		param.Type = ast.ParamAt
		param.Name = val[1:]
	}

	p.advance()
	return param
}

func (p *Parser) parseParenOrSubquery() ast.Expr {
	pos := p.cur.Pos
	p.advance() // consume '('

	// Check if it's a subquery
	if p.curIs(token.SELECT) || p.curIs(token.WITH) {
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
		endPos := p.cur.Pos
		sel, ok := stmt.(*ast.SelectStmt)
		if !ok {
			p.errorf("expected SELECT statement in subquery")
			return nil
		}
		return &ast.Subquery{StartPos: pos, EndPos: endPos, Select: sel}
	}

	// Regular parenthesized expression
	expr := p.parseExpr()
	if !p.expect(token.RPAREN) {
		return nil
	}
	endPos := p.cur.Pos
	return &ast.ParenExpr{StartPos: pos, EndPos: endPos, Expr: expr}
}

func (p *Parser) parseNotExpr() *ast.UnaryExpr {
	pos := p.cur.Pos
	p.advance() // consume NOT

	u := ast.GetUnaryExpr()
	u.StartPos = pos
	u.Op = token.NOT
	u.Operand = p.parseExprPrec(precNot)
	return u
}

func (p *Parser) parseUnaryMinus() *ast.UnaryExpr {
	pos := p.cur.Pos
	p.advance() // consume -

	u := ast.GetUnaryExpr()
	u.StartPos = pos
	u.Op = token.MINUS
	u.Operand = p.parseExprPrec(precUnary)
	return u
}

func (p *Parser) parseUnaryBitnot() *ast.UnaryExpr {
	pos := p.cur.Pos
	p.advance() // consume ~

	u := ast.GetUnaryExpr()
	u.StartPos = pos
	u.Op = token.BITNOT
	u.Operand = p.parseExprPrec(precUnary)
	return u
}

func (p *Parser) parseExistsExpr() *ast.ExistsExpr {
	pos := p.cur.Pos
	p.advance() // consume EXISTS

	not := false
	if p.curIs(token.NOT) {
		not = true
		p.advance()
	}

	if !p.expect(token.LPAREN) {
		return nil
	}

	var sel *ast.SelectStmt
	if p.curIs(token.SELECT) {
		sel = p.parseSelect()
	} else if p.curIs(token.WITH) {
		stmt := p.parseWith()
		if s, ok := stmt.(*ast.SelectStmt); ok {
			sel = s
		}
	}

	if sel == nil {
		p.errorf("expected SELECT in EXISTS subquery")
		return nil
	}

	if !p.expect(token.RPAREN) {
		return nil
	}

	return &ast.ExistsExpr{
		StartPos: pos,
		EndPos:   p.cur.Pos,
		Not:      not,
		Subquery: &ast.Subquery{Select: sel},
	}
}

func (p *Parser) parseCaseExpr() *ast.CaseExpr {
	pos := p.cur.Pos
	p.advance() // consume CASE

	caseExpr := &ast.CaseExpr{StartPos: pos}

	// Check for simple CASE (CASE expr WHEN ...)
	if !p.curIs(token.WHEN) {
		caseExpr.Operand = p.parseExpr()
	}

	// Parse WHEN clauses
	for p.curIs(token.WHEN) {
		p.advance() // consume WHEN
		cond := p.parseExpr()
		if !p.expect(token.THEN) {
			return nil
		}
		result := p.parseExpr()
		caseExpr.Whens = append(caseExpr.Whens, &ast.When{
			Cond:   cond,
			Result: result,
		})
	}

	// Optional ELSE
	if p.curIs(token.ELSE) {
		p.advance()
		caseExpr.Else = p.parseExpr()
	}

	if !p.expect(token.END) {
		return nil
	}

	caseExpr.EndPos = p.cur.Pos
	return caseExpr
}

func (p *Parser) parseCastExpr() *ast.CastExpr {
	pos := p.cur.Pos
	p.advance() // consume CAST

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := p.parseExpr()

	if !p.expect(token.AS) {
		return nil
	}

	dataType := p.parseDataType()

	if !p.expect(token.RPAREN) {
		return nil
	}

	return &ast.CastExpr{
		StartPos: pos,
		EndPos:   p.cur.Pos,
		Expr:     expr,
		Type:     dataType,
	}
}

func (p *Parser) parsePostgresCast(left ast.Expr) *ast.CastExpr {
	p.advance() // consume ::
	dataType := p.parseDataType()

	return &ast.CastExpr{
		StartPos: left.Pos(),
		EndPos:   p.cur.Pos,
		Expr:     left,
		Type:     dataType,
	}
}

func (p *Parser) parseIntervalExpr() *ast.IntervalExpr {
	pos := p.cur.Pos
	p.advance() // consume INTERVAL

	expr := &ast.IntervalExpr{StartPos: pos}
	expr.Value = p.parseExpr()

	// Parse unit (YEAR, MONTH, DAY, etc.)
	if p.cur.Type.IsKeyword() {
		expr.Unit = p.cur.Value
		p.advance()
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseExtractExpr() *ast.ExtractExpr {
	pos := p.cur.Pos
	p.advance() // consume EXTRACT

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := &ast.ExtractExpr{StartPos: pos}

	// Parse field (YEAR, MONTH, DAY, etc.)
	if p.cur.Type.IsKeyword() || p.curIs(token.IDENT) {
		expr.Field = p.cur.Value
		p.advance()
	}

	if !p.expect(token.FROM) {
		return nil
	}

	expr.Source = p.parseExpr()

	if !p.expect(token.RPAREN) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseTrimExpr() *ast.TrimExpr {
	pos := p.cur.Pos
	p.advance() // consume TRIM

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := &ast.TrimExpr{StartPos: pos, TrimType: ast.TrimBoth}

	// Check for LEADING, TRAILING, BOTH
	switch p.cur.Type {
	case token.LEADING:
		expr.TrimType = ast.TrimLeading
		p.advance()
	case token.TRAILING:
		expr.TrimType = ast.TrimTrailing
		p.advance()
	case token.BOTH:
		expr.TrimType = ast.TrimBoth
		p.advance()
	}

	// Optional trim character(s)
	if !p.curIs(token.FROM) {
		expr.TrimChar = p.parseExpr()
	}

	if p.curIs(token.FROM) {
		p.advance()
	}

	expr.Expr = p.parseExpr()

	if !p.expect(token.RPAREN) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseSubstringExpr() *ast.SubstringExpr {
	pos := p.cur.Pos
	p.advance() // consume SUBSTRING

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := &ast.SubstringExpr{StartPos: pos}
	expr.Expr = p.parseExpr()

	if p.curIs(token.FROM) {
		p.advance()
		expr.From = p.parseExpr()
	} else if p.curIs(token.COMMA) {
		p.advance()
		expr.From = p.parseExpr()
	}

	if p.curIs(token.FOR) {
		p.advance()
		expr.For = p.parseExpr()
	} else if p.curIs(token.COMMA) {
		p.advance()
		expr.For = p.parseExpr()
	}

	if !p.expect(token.RPAREN) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parsePositionExpr() *ast.PositionExpr {
	pos := p.cur.Pos
	p.advance() // consume POSITION

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := &ast.PositionExpr{StartPos: pos}
	expr.Needle = p.parseExpr()

	if !p.expect(token.IN) {
		return nil
	}

	expr.Haystack = p.parseExpr()

	if !p.expect(token.RPAREN) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseArrayExpr() *ast.ArrayExpr {
	pos := p.cur.Pos
	p.advance() // consume ARRAY

	if !p.expect(token.LBRACKET) {
		return nil
	}

	expr := &ast.ArrayExpr{StartPos: pos}

	if !p.curIs(token.RBRACKET) {
		for {
			elem := p.parseExpr()
			if elem == nil {
				break
			}
			expr.Elements = append(expr.Elements, elem)
			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
	}

	if !p.expect(token.RBRACKET) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseSubscript(left ast.Expr) *ast.SubscriptExpr {
	p.advance() // consume [

	index := p.parseExpr()
	if index == nil {
		p.errorf("expected expression in subscript")
		return nil
	}

	expr := &ast.SubscriptExpr{
		StartPos: left.Pos(),
		Expr:     left,
		Index:    index,
	}

	if !p.expect(token.RBRACKET) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseIsExpr(left ast.Expr) *ast.IsExpr {
	pos := left.Pos()
	p.advance() // consume IS

	not := false
	if p.curIs(token.NOT) {
		not = true
		p.advance()
	}

	expr := &ast.IsExpr{
		StartPos: pos,
		Expr:     left,
		Not:      not,
	}

	switch p.cur.Type {
	case token.NULL:
		expr.What = ast.IsNull
	case token.TRUE:
		expr.What = ast.IsTrue
	case token.FALSE:
		expr.What = ast.IsFalse
	case token.UNKNOWN:
		expr.What = ast.IsUnknown
	default:
		p.errorf("expected NULL, TRUE, FALSE, or UNKNOWN after IS")
	}

	p.advance()
	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseInExpr(left ast.Expr, not bool) *ast.InExpr {
	pos := left.Pos()
	p.advance() // consume IN

	if !p.expect(token.LPAREN) {
		return nil
	}

	expr := &ast.InExpr{
		StartPos: pos,
		Expr:     left,
		Not:      not,
	}

	// Check for subquery
	if p.curIs(token.SELECT) || p.curIs(token.WITH) {
		var stmt ast.Statement
		if p.curIs(token.WITH) {
			stmt = p.parseWith()
		} else {
			stmt = p.parseSelect()
		}
		if stmt == nil {
			return nil
		}
		sel, ok := stmt.(*ast.SelectStmt)
		if !ok {
			p.errorf("expected SELECT statement in IN clause")
			return nil
		}
		expr.Select = sel
	} else {
		// Value list
		for {
			val := p.parseExpr()
			if val == nil {
				break
			}
			expr.Values = append(expr.Values, val)
			if !p.curIs(token.COMMA) {
				break
			}
			p.advance()
		}
	}

	if !p.expect(token.RPAREN) {
		return nil
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseBetweenExpr(left ast.Expr, not bool) *ast.BetweenExpr {
	pos := left.Pos()
	p.advance() // consume BETWEEN

	expr := &ast.BetweenExpr{
		StartPos: pos,
		Expr:     left,
		Not:      not,
	}

	// Handle SYMMETRIC/ASYMMETRIC
	if p.curIs(token.SYMMETRIC) || p.curIs(token.ASYMMETRIC) {
		p.advance()
	}

	expr.Low = p.parseExprPrec(precComparison + 1)

	if !p.expect(token.AND) {
		return nil
	}

	expr.High = p.parseExprPrec(precComparison + 1)

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseLikeExpr(left ast.Expr, not bool) *ast.LikeExpr {
	pos := left.Pos()
	ilike := p.curIs(token.ILIKE)
	p.advance() // consume LIKE/ILIKE

	expr := &ast.LikeExpr{
		StartPos: pos,
		Expr:     left,
		Not:      not,
		ILike:    ilike,
	}

	expr.Pattern = p.parseExprPrec(precComparison + 1)

	if p.curIs(token.ESCAPE) {
		p.advance()
		expr.Escape = p.parseExprPrec(precComparison + 1)
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseSimilarExpr(left ast.Expr, not bool) *ast.LikeExpr {
	pos := left.Pos()
	p.advance() // consume SIMILAR
	p.expect(token.TO)

	expr := &ast.LikeExpr{
		StartPos: pos,
		Expr:     left,
		Not:      not,
	}

	expr.Pattern = p.parseExprPrec(precComparison + 1)

	if p.curIs(token.ESCAPE) {
		p.advance()
		expr.Escape = p.parseExprPrec(precComparison + 1)
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseCollateExpr(left ast.Expr) *ast.CollateExpr {
	p.advance() // consume COLLATE

	expr := &ast.CollateExpr{
		StartPos: left.Pos(),
		Expr:     left,
	}

	if p.curIs(token.IDENT) || p.curIs(token.STRING) {
		expr.Collation = p.cur.Value
		p.advance()
	}

	expr.EndPos = p.cur.Pos
	return expr
}

func (p *Parser) parseExprList() []ast.Expr {
	// Get slice from pool
	slicePtr := ast.GetExprSlice()
	exprs := *slicePtr
	for {
		expr := p.parseExpr()
		if expr == nil {
			break
		}
		exprs = append(exprs, expr)
		if !p.curIs(token.COMMA) {
			break
		}
		p.advance()
	}
	return exprs
}

func isBinaryOp(t token.Token) bool {
	switch t {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT,
		token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE,
		token.AND, token.OR, token.XOR,
		token.BITAND, token.BITOR, token.BITXOR, token.LSHIFT, token.RSHIFT,
		token.CONCAT:
		return true
	default:
		return false
	}
}
