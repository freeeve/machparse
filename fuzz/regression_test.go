package fuzz

import (
	"testing"

	"github.com/freeeve/machparse"
)

// TestFuzzRegressions contains edge cases discovered by fuzzing.
// Each test documents a specific edge case that previously caused issues.
// When fuzzing finds a new crash, add a test here with a comment explaining the issue.
func TestFuzzRegressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		note  string
	}{
		// Incomplete function calls with operators
		{
			name:  "function with IN keyword inside",
			input: "SELECT A(*IN",
			note:  "Parser must not panic on incomplete function with keyword",
		},
		{
			name:  "function with IS keyword inside",
			input: "SELECT A(*IS",
			note:  "Parser must not panic on incomplete function with keyword",
		},
		{
			name:  "function with BETWEEN keyword inside",
			input: "SELECT A(*BETWEEN",
			note:  "Parser must not panic on incomplete function with keyword",
		},
		{
			name:  "function with LIKE keyword inside",
			input: "SELECT A(*LIKE",
			note:  "Parser must not panic on incomplete function with keyword",
		},
		{
			name:  "function with SIMILAR keyword inside",
			input: "SELECT A(*SIMILAR",
			note:  "Parser must not panic on incomplete function with keyword",
		},

		// Bracket edge cases
		{
			name:  "number followed by brackets",
			input: "SELECT 0[[",
			note:  "Lexer must handle [ after number",
		},
		{
			name:  "number followed by brackets and number",
			input: "SELECT 0[[0",
			note:  "Lexer must handle [[ sequence",
		},

		// Cast operator edge cases
		{
			name:  "cast with empty backtick identifier",
			input: "SELECT 0::``",
			note:  "Cast to empty identifier",
		},
		{
			name:  "incomplete cast in function",
			input: "SELECT A(::",
			note:  "Cast operator at unexpected position",
		},

		// Unary operator edge cases
		{
			name:  "double unary minus",
			input: "SELECT - -0",
			note:  "Multiple unary operators",
		},
		{
			name:  "double unary minus no space",
			input: "SELECT --0",
			note:  "Could be comment or double minus",
		},

		// Dollar quoting edge cases
		{
			name:  "dollar quote with single quotes",
			input: "SELECT $$'''$$",
			note:  "Single quotes inside dollar quotes",
		},
		{
			name:  "dollar quote with backslash",
			input: "SELECT $$\\$$0",
			note:  "Backslash in dollar quote",
		},

		// Empty identifier edge cases
		{
			name:  "empty double-quoted identifier star",
			input: `SELECT"".*%0`,
			note:  "Empty quoted identifier with qualified star",
		},
		{
			name:  "empty double-quoted identifier function",
			input: `SELECT""""(0)`,
			note:  "Empty quoted identifier as function name",
		},

		// EXTRACT edge cases
		{
			name:  "extract with empty quoted field",
			input: `SELECT EXTRACT(""FROM 0)`,
			note:  "Empty field name in EXTRACT",
		},

		// Malformed expressions
		{
			name:  "exists with empty parens",
			input: "SELECT 00WHERE EXISTS()0000000",
			note:  "Number tokens adjacent to keywords",
		},

		// Additional edge cases from typical fuzzing
		{
			name:  "empty input",
			input: "",
			note:  "Empty input should not panic",
		},
		{
			name:  "only whitespace",
			input: "   \t\n\r  ",
			note:  "Whitespace only should not panic",
		},
		{
			name:  "only semicolons",
			input: ";;;",
			note:  "Multiple empty statements",
		},
		{
			name:  "unclosed string",
			input: "SELECT 'unclosed",
			note:  "Unclosed string literal",
		},
		{
			name:  "unclosed parenthesis",
			input: "SELECT (1 + 2",
			note:  "Missing closing paren",
		},
		{
			name:  "too many close parens",
			input: "SELECT (1))",
			note:  "Extra closing paren",
		},
		{
			name:  "null bytes",
			input: "SELECT\x00*",
			note:  "Null byte in input",
		},
		{
			name:  "deeply nested",
			input: "SELECT ((((((((((1))))))))))",
			note:  "Deeply nested parentheses",
		},
		{
			name:  "very long identifier",
			input: "SELECT aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa FROM t",
			note:  "Long identifier (100 chars)",
		},

		// Incomplete statements (should error, not panic)
		{
			name:  "select with parenthesized WITH",
			input: "SELECT(WITH)",
			note:  "Incomplete WITH in subquery should not panic",
		},
		{
			name:  "insert with incomplete SET",
			input: "INSERT INTO A SET",
			note:  "SET without assignments should error",
		},
		{
			name:  "trailing operators",
			input: "SELECT * % 0",
			note:  "Trailing operators after statement should error",
		},
		{
			name:  "incomplete qualified table with paren",
			input: "SELECT*FROM A.(use",
			note:  "Qualified name followed by paren should not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The parser should never panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Parser panicked: %v\nInput: %q\nNote: %s", r, tt.input, tt.note)
				}
			}()

			stmt, err := machparse.Parse(tt.input)

			// Parse errors are acceptable - we're testing for panics
			if err != nil {
				return
			}

			if stmt == nil {
				return
			}

			// If parsing succeeded, formatting should not panic
			formatted := machparse.String(stmt)
			if formatted == "" {
				t.Logf("Warning: valid parse but empty format for: %q", tt.input)
			}
		})
	}
}

// TestFuzzRoundTrip tests that valid SQL round-trips correctly.
// Add cases here when fuzzing finds formatting issues.
func TestFuzzRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple select", "SELECT * FROM t"},
		{"select with alias", "SELECT a AS b FROM t"},
		{"join", "SELECT * FROM t1 JOIN t2 ON t1.id = t2.id"},
		{"subquery", "SELECT * FROM (SELECT 1) AS sub"},
		{"cte", "WITH cte AS (SELECT 1) SELECT * FROM cte"},
		{"union", "SELECT 1 UNION SELECT 2"},
		{"parenthesized union", "(SELECT 1) UNION (SELECT 2)"},
		{"case expression", "SELECT CASE WHEN a = 1 THEN 'x' ELSE 'y' END FROM t"},
		{"window function", "SELECT ROW_NUMBER() OVER (ORDER BY id) FROM t"},
		{"multi-level identifier", "SELECT a.b.c.d FROM a.b.c"},
		{"bracket identifier", "SELECT [col] FROM [table]"},
		{"temp table", "SELECT * FROM #temp"},
		{"postgresql cast", "SELECT a::int FROM t"},
		{"array literal", "SELECT ARRAY[1, 2, 3]"},
		{"array with identifier", "SELECT ARRAY[ A]"},
		{"subscript with space", "SELECT arr[ idx ]"},
		{"create index no name", "CREATE INDEX ON A(col)"},
		{"function with keyword name", "SELECT ANY(x) FROM t"},
		{"large type precision", "SELECT 0::A(10000000000000000000)"},
		{"empty backtick in index", "CREATE INDEX ON A(``)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := machparse.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			formatted1 := machparse.String(stmt)
			if formatted1 == "" {
				t.Fatal("Format returned empty string")
			}

			stmt2, err := machparse.Parse(formatted1)
			if err != nil {
				t.Fatalf("Re-parse failed: %v\nFormatted: %s", err, formatted1)
			}

			formatted2 := machparse.String(stmt2)
			if formatted1 != formatted2 {
				t.Errorf("Round-trip mismatch:\nInput:     %s\nFormatted: %s\nRe-format: %s",
					tt.input, formatted1, formatted2)
			}
		})
	}
}
