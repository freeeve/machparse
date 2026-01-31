// Package lexer provides a lexical scanner for SQL.
package lexer

import (
	"sync"

	"github.com/freeeve/machparse/token"
)

// Lexer tokenizes SQL input.
type Lexer struct {
	input   string
	start   int        // start position of current token
	pos     int        // current position in input
	line    int        // current line number (1-indexed)
	linePos int        // position of current line start
	item    token.Item // most recently scanned item
	peeked  bool       // whether item contains a peeked token
}

var lexerPool = sync.Pool{
	New: func() any { return &Lexer{} },
}

// New creates a new Lexer for the input string.
func New(input string) *Lexer {
	return &Lexer{
		input:   input,
		line:    1,
		linePos: 0,
	}
}

// Get returns a Lexer from the pool, initialized with the input.
func Get(input string) *Lexer {
	l := lexerPool.Get().(*Lexer)
	l.Reset(input)
	return l
}

// Put returns the Lexer to the pool.
func Put(l *Lexer) {
	lexerPool.Put(l)
}

// Reset resets the lexer to scan new input.
func (l *Lexer) Reset(input string) {
	l.input = input
	l.start = 0
	l.pos = 0
	l.line = 1
	l.linePos = 0
	l.item = token.Item{}
	l.peeked = false
}

// Next returns the next token.
func (l *Lexer) Next() token.Item {
	if l.peeked {
		l.peeked = false
		return l.item
	}
	l.item = l.scan()
	return l.item
}

// Peek returns the next token without consuming it.
func (l *Lexer) Peek() token.Item {
	if !l.peeked {
		l.item = l.scan()
		l.peeked = true
	}
	return l.item
}

// scan performs the actual lexical analysis.
func (l *Lexer) scan() token.Item {
	l.skipWhitespace()
	l.start = l.pos

	if l.pos >= len(l.input) {
		return l.makeItem(token.EOF, "")
	}

	ch := l.input[l.pos]

	// Fast path for common single-character tokens
	switch ch {
	case '(':
		l.pos++
		return l.makeItem(token.LPAREN, "(")
	case ')':
		l.pos++
		return l.makeItem(token.RPAREN, ")")
	case '[':
		// Check if this is a SQL Server bracket-quoted identifier
		return l.scanBracketOrLBracket()
	case ']':
		l.pos++
		return l.makeItem(token.RBRACKET, "]")
	case ',':
		l.pos++
		return l.makeItem(token.COMMA, ",")
	case ';':
		l.pos++
		return l.makeItem(token.SEMICOLON, ";")
	case '+':
		l.pos++
		return l.makeItem(token.PLUS, "+")
	case '*':
		l.pos++
		return l.makeItem(token.ASTERISK, "*")
	case '%':
		l.pos++
		return l.makeItem(token.PERCENT, "%")
	case '~':
		l.pos++
		return l.makeItem(token.BITNOT, "~")
	case '^':
		l.pos++
		return l.makeItem(token.BITXOR, "^")
	case '@':
		return l.scanAt()
	case '.':
		if l.pos+1 < len(l.input) && isDigit(l.input[l.pos+1]) {
			return l.scanNumber()
		}
		l.pos++
		return l.makeItem(token.DOT, ".")
	case '-':
		return l.scanMinus()
	case '/':
		return l.scanSlash()
	case '\'':
		return l.scanString('\'')
	case '"':
		return l.scanQuotedIdentifier()
	case '`':
		return l.scanBacktickIdentifier()
	case '=':
		l.pos++
		return l.makeItem(token.EQ, "=")
	case '<':
		return l.scanLessThan()
	case '>':
		return l.scanGreaterThan()
	case '!':
		return l.scanBang()
	case '|':
		return l.scanPipe()
	case '&':
		l.pos++
		return l.makeItem(token.BITAND, "&")
	case '?':
		return l.scanQuestion()
	case '$':
		return l.scanDollar()
	case ':':
		return l.scanColon()
	case '#':
		return l.scanHash()
	}

	// Identifiers and keywords
	if isIdentStart(ch) {
		return l.scanIdentifier()
	}

	// Numbers
	if isDigit(ch) {
		return l.scanNumber()
	}

	// Unknown character
	l.pos++
	return l.makeItem(token.ILLEGAL, string(ch))
}

func (l *Lexer) makeItem(typ token.Token, val string) token.Item {
	return token.Item{
		Type:  typ,
		Value: val,
		Pos: token.Pos{
			Offset: l.start,
			Line:   l.line,
			Column: l.start - l.linePos + 1,
		},
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
		} else if ch == '\n' {
			l.pos++
			l.line++
			l.linePos = l.pos
		} else {
			break
		}
	}
}

func (l *Lexer) scanIdentifier() token.Item {
	for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
		l.pos++
	}
	val := l.input[l.start:l.pos]
	tok := token.LookupIdent(val)
	return l.makeItem(tok, val)
}

func (l *Lexer) scanNumber() token.Item {
	tok := token.INT

	// Handle hex numbers: 0x...
	if l.pos+1 < len(l.input) && l.input[l.pos] == '0' &&
		(l.input[l.pos+1] == 'x' || l.input[l.pos+1] == 'X') {
		l.pos += 2
		for l.pos < len(l.input) && isHexDigit(l.input[l.pos]) {
			l.pos++
		}
		return l.makeItem(token.INT, l.input[l.start:l.pos])
	}

	// Integer part
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}

	// Decimal part
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		// Check it's not a range operator (..)
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '.' {
			return l.makeItem(tok, l.input[l.start:l.pos])
		}
		tok = token.FLOAT
		l.pos++
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	// Exponent
	if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
		tok = token.FLOAT
		l.pos++
		if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
			l.pos++
		}
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	return l.makeItem(tok, l.input[l.start:l.pos])
}

func (l *Lexer) scanString(quote byte) token.Item {
	l.pos++ // skip opening quote
	var buf []byte
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == quote {
			// Check for escaped quote ('')
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == quote {
				buf = append(buf, quote)
				l.pos += 2
				continue
			}
			// End of string
			l.pos++
			return l.makeItem(token.STRING, string(buf))
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			// Handle escape sequences - interpret them
			next := l.input[l.pos+1]
			switch next {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			case '\\':
				buf = append(buf, '\\')
			case '\'':
				buf = append(buf, '\'')
			case '"':
				buf = append(buf, '"')
			default:
				// Unknown escape - keep the backslash and char
				buf = append(buf, '\\', next)
			}
			l.pos += 2
			continue
		}
		if ch == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		buf = append(buf, ch)
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanQuotedIdentifier() token.Item {
	l.pos++ // skip opening "
	var buf []byte
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			// Check for escaped quote
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '"' {
				buf = append(buf, '"')
				l.pos += 2
				continue
			}
			l.pos++
			// Extract the identifier without quotes, handling escapes
			if buf == nil {
				return l.makeItem(token.IDENT, l.input[l.start+1:l.pos-1])
			}
			return l.makeItem(token.IDENT, string(buf))
		}
		if ch == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		buf = append(buf, ch)
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanBacktickIdentifier() token.Item {
	l.pos++ // skip opening `
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '`' {
			// Check for escaped backtick
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '`' {
				l.pos += 2
				continue
			}
			l.pos++
			// Extract the identifier without backticks
			val := l.input[l.start+1 : l.pos-1]
			return l.makeItem(token.IDENT, val)
		}
		if ch == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanBracketOrLBracket() token.Item {
	// Peek at next character to see if this is a bracket-quoted identifier
	if l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		// If followed by identifier-start char (letter, underscore) or # @ for temp tables/variables,
		// treat as SQL Server bracket-quoted identifier.
		// Do NOT include space here - that allows array subscripts to use [ expr ] format.
		if isIdentStart(next) || next == '#' || next == '@' {
			return l.scanBracketIdentifier()
		}
	}
	// Otherwise just return LBRACKET for array subscript
	l.pos++
	return l.makeItem(token.LBRACKET, "[")
}

func (l *Lexer) scanBracketIdentifier() token.Item {
	l.pos++ // skip opening [
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ']' {
			// Check for escaped bracket ]]
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == ']' {
				l.pos += 2
				continue
			}
			l.pos++
			// Extract the identifier without brackets
			val := l.input[l.start+1 : l.pos-1]
			return l.makeItem(token.IDENT, val)
		}
		if ch == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanMinus() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '-':
			// Line comment
			return l.scanLineComment()
		case '>':
			l.pos++
			if l.pos < len(l.input) && l.input[l.pos] == '>' {
				l.pos++
				return l.makeItem(token.DARROW, "->>")
			}
			return l.makeItem(token.ARROW, "->")
		}
	}
	return l.makeItem(token.MINUS, "-")
}

func (l *Lexer) scanSlash() token.Item {
	l.pos++
	if l.pos < len(l.input) && l.input[l.pos] == '*' {
		return l.scanBlockComment()
	}
	return l.makeItem(token.SLASH, "/")
}

func (l *Lexer) scanLineComment() token.Item {
	l.pos++ // skip second -
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	return l.makeItem(token.COMMENT, l.input[l.start:l.pos])
}

func (l *Lexer) scanBlockComment() token.Item {
	l.pos++ // skip *
	for l.pos < len(l.input) {
		if l.input[l.pos] == '*' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			l.pos += 2
			return l.makeItem(token.COMMENT, l.input[l.start:l.pos])
		}
		if l.input[l.pos] == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanLessThan() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '=':
			l.pos++
			return l.makeItem(token.LTE, "<=")
		case '>':
			l.pos++
			return l.makeItem(token.NEQ, "<>")
		case '<':
			l.pos++
			return l.makeItem(token.LSHIFT, "<<")
		}
	}
	return l.makeItem(token.LT, "<")
}

func (l *Lexer) scanGreaterThan() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '=':
			l.pos++
			return l.makeItem(token.GTE, ">=")
		case '>':
			l.pos++
			return l.makeItem(token.RSHIFT, ">>")
		}
	}
	return l.makeItem(token.GT, ">")
}

func (l *Lexer) scanBang() token.Item {
	l.pos++
	if l.pos < len(l.input) && l.input[l.pos] == '=' {
		l.pos++
		return l.makeItem(token.NEQ, "!=")
	}
	return l.makeItem(token.ILLEGAL, "!")
}

func (l *Lexer) scanPipe() token.Item {
	l.pos++
	if l.pos < len(l.input) && l.input[l.pos] == '|' {
		l.pos++
		return l.makeItem(token.CONCAT, "||")
	}
	return l.makeItem(token.BITOR, "|")
}

func (l *Lexer) scanQuestion() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '|':
			l.pos++
			return l.makeItem(token.QUESTIONOR, "?|")
		case '&':
			l.pos++
			return l.makeItem(token.QUESTIONAND, "?&")
		}
	}
	return l.makeItem(token.PARAM, "?")
}

func (l *Lexer) scanDollar() token.Item {
	l.pos++
	// Check for positional parameter $1, $2, etc.
	if l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
		return l.makeItem(token.PARAM, l.input[l.start:l.pos])
	}
	// Check for dollar-quoted string $$...$$ or $tag$...$tag$
	if l.pos < len(l.input) {
		tag := ""
		if l.input[l.pos] == '$' {
			// $$...$$ form
			l.pos++ // skip second $
		} else if isIdentStart(l.input[l.pos]) {
			// $tag$...$tag$ form - tag cannot contain $
			tagStart := l.pos
			for l.pos < len(l.input) && isTagChar(l.input[l.pos]) {
				l.pos++
			}
			if l.pos < len(l.input) && l.input[l.pos] == '$' {
				tag = l.input[tagStart:l.pos]
				l.pos++ // skip closing $ of opening delimiter
			} else {
				// Not a dollar-quoted string
				l.pos = l.start + 1
				return l.makeItem(token.ILLEGAL, "$")
			}
		} else {
			return l.makeItem(token.ILLEGAL, "$")
		}
		return l.scanDollarQuotedStringContent(tag)
	}
	return l.makeItem(token.ILLEGAL, "$")
}

func (l *Lexer) scanDollarQuotedStringContent(tag string) token.Item {
	contentStart := l.pos
	endDelim := "$" + tag + "$"

	for l.pos < len(l.input) {
		if l.input[l.pos] == '$' {
			// Check for closing delimiter
			if l.pos+len(endDelim) <= len(l.input) &&
				l.input[l.pos:l.pos+len(endDelim)] == endDelim {
				content := l.input[contentStart:l.pos]
				l.pos += len(endDelim)
				return l.makeItem(token.STRING, content)
			}
		}
		if l.input[l.pos] == '\n' {
			l.line++
			l.linePos = l.pos + 1
		}
		l.pos++
	}
	return l.makeItem(token.ILLEGAL, l.input[l.start:l.pos])
}

func (l *Lexer) scanColon() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case ':':
			l.pos++
			return l.makeItem(token.DCOLON, "::")
		default:
			// Named parameter :name
			if isIdentStart(l.input[l.pos]) {
				for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
					l.pos++
				}
				return l.makeItem(token.PARAM, l.input[l.start:l.pos])
			}
		}
	}
	return l.makeItem(token.COLON, ":")
}

func (l *Lexer) scanHash() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '>':
			l.pos++
			if l.pos < len(l.input) && l.input[l.pos] == '>' {
				l.pos++
				return l.makeItem(token.HASHDGT, "#>>")
			}
			return l.makeItem(token.HASHGT, "#>")
		case '#':
			// ##global_temp_table (SQL Server global temp table)
			l.pos++
			if l.pos < len(l.input) && isIdentStart(l.input[l.pos]) {
				for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
					l.pos++
				}
				return l.makeItem(token.IDENT, l.input[l.start:l.pos])
			}
			// Just ## without identifier - treat as comment
			l.pos -= 2
		default:
			// SQL Server temp table: #identifier
			if isIdentStart(l.input[l.pos]) {
				for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
					l.pos++
				}
				return l.makeItem(token.IDENT, l.input[l.start:l.pos])
			}
		}
	}
	// MySQL-style comment or just hash
	// For now, treat single # as line comment start (MySQL style)
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	return l.makeItem(token.COMMENT, l.input[l.start:l.pos])
}

func (l *Lexer) scanAt() token.Item {
	l.pos++
	if l.pos < len(l.input) {
		switch l.input[l.pos] {
		case '@':
			l.pos++
			return l.makeItem(token.ATAT, "@@")
		default:
			// MySQL user variable @name
			if isIdentStart(l.input[l.pos]) {
				for l.pos < len(l.input) && isIdentChar(l.input[l.pos]) {
					l.pos++
				}
				return l.makeItem(token.PARAM, l.input[l.start:l.pos])
			}
		}
	}
	return l.makeItem(token.AT, "@")
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '$'
}

func isTagChar(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}
