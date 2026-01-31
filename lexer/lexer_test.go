package lexer

import (
	"testing"

	"github.com/freeeve/machparse/token"
)

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Item
	}{
		{
			input: "SELECT * FROM users",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.ASTERISK, Value: "*"},
				{Type: token.FROM, Value: "FROM"},
				{Type: token.IDENT, Value: "users"},
				{Type: token.EOF, Value: ""},
			},
		},
		{
			input: "SELECT id, name FROM users WHERE id = 1",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.IDENT, Value: "id"},
				{Type: token.COMMA, Value: ","},
				{Type: token.IDENT, Value: "name"},
				{Type: token.FROM, Value: "FROM"},
				{Type: token.IDENT, Value: "users"},
				{Type: token.WHERE, Value: "WHERE"},
				{Type: token.IDENT, Value: "id"},
				{Type: token.EQ, Value: "="},
				{Type: token.INT, Value: "1"},
				{Type: token.EOF, Value: ""},
			},
		},
		{
			input: "a >= b AND c <= d",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.GTE, Value: ">="},
				{Type: token.IDENT, Value: "b"},
				{Type: token.AND, Value: "AND"},
				{Type: token.IDENT, Value: "c"},
				{Type: token.LTE, Value: "<="},
				{Type: token.IDENT, Value: "d"},
				{Type: token.EOF, Value: ""},
			},
		},
		{
			input: "a <> b OR a != c",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.NEQ, Value: "<>"},
				{Type: token.IDENT, Value: "b"},
				{Type: token.OR, Value: "OR"},
				{Type: token.IDENT, Value: "a"},
				{Type: token.NEQ, Value: "!="},
				{Type: token.IDENT, Value: "c"},
				{Type: token.EOF, Value: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				got := l.Next()
				if got.Type != exp.Type {
					t.Errorf("token %d: expected type %v, got %v", i, exp.Type, got.Type)
				}
				if got.Value != exp.Value {
					t.Errorf("token %d: expected value %q, got %q", i, exp.Value, got.Value)
				}
			}
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Item
	}{
		{"123", token.Item{Type: token.INT, Value: "123"}},
		{"123.456", token.Item{Type: token.FLOAT, Value: "123.456"}},
		{".456", token.Item{Type: token.FLOAT, Value: ".456"}},
		{"1e10", token.Item{Type: token.FLOAT, Value: "1e10"}},
		{"1E10", token.Item{Type: token.FLOAT, Value: "1E10"}},
		{"1.5e+10", token.Item{Type: token.FLOAT, Value: "1.5e+10"}},
		{"1.5e-10", token.Item{Type: token.FLOAT, Value: "1.5e-10"}},
		{"0x1A2B", token.Item{Type: token.INT, Value: "0x1A2B"}},
		{"0X1a2b", token.Item{Type: token.INT, Value: "0X1a2b"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			got := l.Next()
			if got.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, got.Type)
			}
			if got.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, got.Value)
			}
		})
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Item
	}{
		{"'hello'", token.Item{Type: token.STRING, Value: "hello"}},
		{"'hello world'", token.Item{Type: token.STRING, Value: "hello world"}},
		{"'it''s'", token.Item{Type: token.STRING, Value: "it's"}},
		{"'line1\nline2'", token.Item{Type: token.STRING, Value: "line1\nline2"}},
		{"'escaped\\nchar'", token.Item{Type: token.STRING, Value: "escaped\nchar"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			got := l.Next()
			if got.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, got.Type)
			}
			if got.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, got.Value)
			}
		})
	}
}

func TestLexerQuotedIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Item
	}{
		{`"column"`, token.Item{Type: token.IDENT, Value: "column"}},
		{`"Column Name"`, token.Item{Type: token.IDENT, Value: "Column Name"}},
		{`"escaped""quote"`, token.Item{Type: token.IDENT, Value: `escaped"quote`}},
		{"`column`", token.Item{Type: token.IDENT, Value: "column"}},
		{"`Column Name`", token.Item{Type: token.IDENT, Value: "Column Name"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			got := l.Next()
			if got.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, got.Type)
			}
			if got.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, got.Value)
			}
		})
	}
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Item
	}{
		{
			input: "a || b",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.CONCAT, Value: "||"},
				{Type: token.IDENT, Value: "b"},
			},
		},
		{
			input: "a | b & c",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.BITOR, Value: "|"},
				{Type: token.IDENT, Value: "b"},
				{Type: token.BITAND, Value: "&"},
				{Type: token.IDENT, Value: "c"},
			},
		},
		{
			input: "a << 2 >> 1",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.LSHIFT, Value: "<<"},
				{Type: token.INT, Value: "2"},
				{Type: token.RSHIFT, Value: ">>"},
				{Type: token.INT, Value: "1"},
			},
		},
		{
			input: "jsondata->>'key'",
			expected: []token.Item{
				{Type: token.IDENT, Value: "jsondata"},
				{Type: token.DARROW, Value: "->>"},
				{Type: token.STRING, Value: "key"},
			},
		},
		{
			input: "jsondata->'key'",
			expected: []token.Item{
				{Type: token.IDENT, Value: "jsondata"},
				{Type: token.ARROW, Value: "->"},
				{Type: token.STRING, Value: "key"},
			},
		},
		{
			input: "a::int",
			expected: []token.Item{
				{Type: token.IDENT, Value: "a"},
				{Type: token.DCOLON, Value: "::"},
				{Type: token.INT_TYPE, Value: "int"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				got := l.Next()
				if got.Type != exp.Type {
					t.Errorf("token %d: expected type %v, got %v", i, exp.Type, got.Type)
				}
				if got.Value != exp.Value {
					t.Errorf("token %d: expected value %q, got %q", i, exp.Value, got.Value)
				}
			}
		})
	}
}

func TestLexerParameters(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Item
	}{
		{"?", token.Item{Type: token.PARAM, Value: "?"}},
		{"$1", token.Item{Type: token.PARAM, Value: "$1"}},
		{"$123", token.Item{Type: token.PARAM, Value: "$123"}},
		{":name", token.Item{Type: token.PARAM, Value: ":name"}},
		{":user_id", token.Item{Type: token.PARAM, Value: ":user_id"}},
		{"@var", token.Item{Type: token.PARAM, Value: "@var"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			got := l.Next()
			if got.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, got.Type)
			}
			if got.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, got.Value)
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Item
	}{
		{
			input: "SELECT -- comment\n1",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.COMMENT, Value: "-- comment"},
				{Type: token.INT, Value: "1"},
			},
		},
		{
			input: "SELECT /* comment */ 1",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.COMMENT, Value: "/* comment */"},
				{Type: token.INT, Value: "1"},
			},
		},
		{
			input: "SELECT /* multi\nline\ncomment */ 1",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.COMMENT, Value: "/* multi\nline\ncomment */"},
				{Type: token.INT, Value: "1"},
			},
		},
		{
			input: "SELECT # mysql comment\n1",
			expected: []token.Item{
				{Type: token.SELECT, Value: "SELECT"},
				{Type: token.COMMENT, Value: "# mysql comment"},
				{Type: token.INT, Value: "1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				got := l.Next()
				if got.Type != exp.Type {
					t.Errorf("token %d: expected type %v, got %v", i, exp.Type, got.Type)
				}
				if got.Value != exp.Value {
					t.Errorf("token %d: expected value %q, got %q", i, exp.Value, got.Value)
				}
			}
		})
	}
}

func TestLexerPositions(t *testing.T) {
	input := "SELECT\n  id\nFROM t"
	l := New(input)

	expected := []struct {
		tok  token.Token
		line int
		col  int
	}{
		{token.SELECT, 1, 1},
		{token.IDENT, 2, 3},
		{token.FROM, 3, 1},
		{token.IDENT, 3, 6},
	}

	for _, exp := range expected {
		got := l.Next()
		if got.Type != exp.tok {
			t.Errorf("expected token %v, got %v", exp.tok, got.Type)
		}
		if got.Pos.Line != exp.line {
			t.Errorf("token %v: expected line %d, got %d", got.Type, exp.line, got.Pos.Line)
		}
		if got.Pos.Column != exp.col {
			t.Errorf("token %v: expected column %d, got %d", got.Type, exp.col, got.Pos.Column)
		}
	}
}

func TestLexerPeek(t *testing.T) {
	l := New("SELECT FROM")

	// Peek should return SELECT
	peek1 := l.Peek()
	if peek1.Type != token.SELECT {
		t.Errorf("expected SELECT, got %v", peek1.Type)
	}

	// Peek again should return the same token
	peek2 := l.Peek()
	if peek2.Type != token.SELECT {
		t.Errorf("expected SELECT, got %v", peek2.Type)
	}

	// Next should return SELECT
	next1 := l.Next()
	if next1.Type != token.SELECT {
		t.Errorf("expected SELECT, got %v", next1.Type)
	}

	// Next should return FROM
	next2 := l.Next()
	if next2.Type != token.FROM {
		t.Errorf("expected FROM, got %v", next2.Type)
	}
}

func TestLexerDollarQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Item
	}{
		{"$$hello$$", token.Item{Type: token.STRING, Value: "hello"}},
		{"$$hello world$$", token.Item{Type: token.STRING, Value: "hello world"}},
		{"$tag$content$tag$", token.Item{Type: token.STRING, Value: "content"}},
		{"$$multi\nline$$", token.Item{Type: token.STRING, Value: "multi\nline"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			got := l.Next()
			if got.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, got.Type)
			}
			if got.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, got.Value)
			}
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "NOT", "IN", "LIKE", "BETWEEN",
		"IS", "NULL", "TRUE", "FALSE", "AS", "JOIN", "INNER", "LEFT", "RIGHT",
		"FULL", "OUTER", "CROSS", "ON", "ORDER", "BY", "ASC", "DESC", "GROUP",
		"HAVING", "LIMIT", "OFFSET", "UNION", "INTERSECT", "EXCEPT", "INSERT",
		"INTO", "VALUES", "UPDATE", "SET", "DELETE", "CREATE", "ALTER", "DROP",
		"TABLE", "INDEX", "IF", "EXISTS", "PRIMARY", "KEY", "FOREIGN", "REFERENCES",
		"UNIQUE", "CONSTRAINT", "CHECK", "CASCADE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "CAST", "DISTINCT", "ALL",
	}

	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			l := New(kw)
			got := l.Next()
			if !got.Type.IsKeyword() {
				t.Errorf("%s should be a keyword, got %v", kw, got.Type)
			}
		})
	}
}

func BenchmarkLexer(b *testing.B) {
	input := `SELECT u.id, u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
  AND u.created_at BETWEEN '2024-01-01' AND '2024-12-31'
GROUP BY u.id, u.name
HAVING COUNT(o.id) > 5
ORDER BY order_count DESC
LIMIT 100`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.Next()
			if tok.Type == token.EOF {
				break
			}
		}
	}
}
