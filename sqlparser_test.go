package machparse

import (
	"testing"

	"github.com/freeeve/machparse/ast"
)

func TestParseAndFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // empty means input == expected
	}{
		{
			name:  "simple select",
			input: "SELECT * FROM users",
		},
		{
			name:  "select with where",
			input: "SELECT id, name FROM users WHERE status = 'active'",
		},
		{
			name:  "select with join",
			input: "SELECT a.id, b.name FROM a JOIN b ON a.id = b.a_id",
		},
		{
			name:  "select with multiple joins",
			input: "SELECT * FROM a LEFT JOIN b ON a.id = b.a_id RIGHT JOIN c ON b.id = c.b_id",
		},
		{
			name:  "select with subquery",
			input: "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders)",
		},
		{
			name:  "insert",
			input: "INSERT INTO users (id, name) VALUES (1, 'test')",
		},
		{
			name:  "update",
			input: "UPDATE users SET name = 'new' WHERE id = 1",
		},
		{
			name:  "delete",
			input: "DELETE FROM users WHERE id = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			formatted := String(stmt)
			if formatted == "" {
				t.Fatal("Formatted output is empty")
			}

			// Re-parse formatted output
			stmt2, err := Parse(formatted)
			if err != nil {
				t.Fatalf("Re-parse error: %v\nFormatted: %s", err, formatted)
			}

			// Format again should be identical
			formatted2 := String(stmt2)
			if formatted != formatted2 {
				t.Errorf("Round-trip mismatch:\nFirst:  %s\nSecond: %s", formatted, formatted2)
			}
		})
	}
}

func TestWalk(t *testing.T) {
	stmt, err := Parse("SELECT a.id, b.name FROM users a JOIN orders b ON a.id = b.user_id WHERE a.status = 'active'")
	if err != nil {
		t.Fatal(err)
	}

	var columns []string
	Walk(stmt, func(node Node) bool {
		if col, ok := node.(*ColName); ok {
			columns = append(columns, col.Name())
		}
		return true
	})

	expected := []string{"id", "name", "id", "user_id", "status"}
	if len(columns) != len(expected) {
		t.Errorf("Expected %d columns, got %d: %v", len(expected), len(columns), columns)
	}
}

func TestRewrite(t *testing.T) {
	stmt, err := Parse("SELECT id, name FROM users WHERE status = 'active'")
	if err != nil {
		t.Fatal(err)
	}

	// Rewrite to add table qualifier to all columns
	rewritten := Rewrite(stmt, func(node Node) Node {
		if col, ok := node.(*ColName); ok && len(col.Parts) == 1 {
			// Add table qualifier "u" to unqualified columns
			return &ColName{
				Parts: []string{"u", col.Name()},
			}
		}
		return node
	})

	formatted := String(rewritten)
	if formatted == "" {
		t.Fatal("Rewritten output is empty")
	}

	// Should contain qualified column names
	t.Logf("Rewritten: %s", formatted)
}

func TestExtractTables(t *testing.T) {
	stmt, err := Parse("SELECT * FROM users u JOIN orders o ON u.id = o.user_id WHERE EXISTS (SELECT 1 FROM items)")
	if err != nil {
		t.Fatal(err)
	}

	tables := ExtractTables(stmt)
	if len(tables) != 3 {
		t.Errorf("Expected 3 tables, got %d: %v", len(tables), tables)
	}
}

func ExtractTables(stmt Statement) []string {
	var tables []string
	seen := make(map[string]bool)
	Walk(stmt, func(node Node) bool {
		// Skip walking into ColName qualifiers by not recursing into ColName
		if _, ok := node.(*ColName); ok {
			return false // Don't recurse into ColName
		}
		if tn, ok := node.(*TableName); ok {
			name := tn.Name()
			if !seen[name] {
				tables = append(tables, name)
				seen[name] = true
			}
		}
		return true
	})
	return tables
}

func TestComplexQueries(t *testing.T) {
	queries := []string{
		`WITH active AS (SELECT id FROM users WHERE status = 'active')
		 SELECT * FROM active`,
		`SELECT id, COUNT(*) as cnt FROM orders GROUP BY id HAVING COUNT(*) > 5`,
		`SELECT ROW_NUMBER() OVER (PARTITION BY type ORDER BY created_at DESC) FROM items`,
		`SELECT CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END FROM users`,
		`SELECT * FROM users WHERE name LIKE '%test%' ESCAPE '\\'`,
		`SELECT * FROM users WHERE created_at BETWEEN '2024-01-01' AND '2024-12-31'`,
		`SELECT COALESCE(name, 'unknown') FROM users`,
		`SELECT CAST(price AS INT) FROM products`,
		`SELECT a || ' ' || b FROM names`,
		`SELECT * FROM users FOR UPDATE`,
		`SELECT * FROM users LIMIT 10 OFFSET 20`,
	}

	for _, q := range queries {
		t.Run(q[:30], func(t *testing.T) {
			stmt, err := Parse(q)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			formatted := String(stmt)
			if formatted == "" {
				t.Error("Empty formatted output")
			}
		})
	}
}

func TestDDL(t *testing.T) {
	queries := []string{
		`CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255) NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS items (id INT, price DECIMAL(10,2))`,
		`ALTER TABLE users ADD COLUMN email VARCHAR(255)`,
		`ALTER TABLE users DROP COLUMN IF EXISTS temp`,
		`DROP TABLE IF EXISTS old_users CASCADE`,
		`CREATE UNIQUE INDEX idx_email ON users (email)`,
		`DROP INDEX idx_old ON users`,
		`TRUNCATE TABLE logs`,
	}

	for _, q := range queries {
		name := q
		if len(name) > 20 {
			name = name[:20]
		}
		t.Run(name, func(t *testing.T) {
			stmt, err := Parse(q)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			formatted := String(stmt)
			if formatted == "" {
				t.Error("Empty formatted output")
			}
		})
	}
}

func TestMultiDialect(t *testing.T) {
	queries := []struct {
		name  string
		query string
	}{
		// MySQL features
		{"mysql replace", "REPLACE INTO users (id, name) VALUES (1, 'test')"},
		{"mysql on duplicate", "INSERT INTO users (id, name) VALUES (1, 'test') ON DUPLICATE KEY UPDATE name = 'new'"},
		{"mysql limit offset", "SELECT * FROM users LIMIT 10, 20"},

		// PostgreSQL features
		{"pg cast", "SELECT a::int FROM t"},
		{"pg returning", "INSERT INTO users (name) VALUES ('test') RETURNING id"},
		{"pg on conflict", "INSERT INTO users (id, name) VALUES (1, 'test') ON CONFLICT (id) DO NOTHING"},
		{"pg array", "SELECT ARRAY[1, 2, 3]"},

		// Common features
		{"cte", "WITH t AS (SELECT 1) SELECT * FROM t"},
		{"window", "SELECT SUM(x) OVER (PARTITION BY y) FROM t"},
		{"exists", "SELECT * FROM t WHERE EXISTS (SELECT 1 FROM u)"},
	}

	for _, tc := range queries {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := Parse(tc.query)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			formatted := String(stmt)
			if formatted == "" {
				t.Error("Empty formatted output")
			}
		})
	}
}

func TestMultiLevelIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCols int
	}{
		{
			name:     "simple column",
			input:    "SELECT a FROM t",
			wantCols: 1,
		},
		{
			name:     "two-level column",
			input:    "SELECT t.a FROM t",
			wantCols: 1,
		},
		{
			name:     "three-level column",
			input:    "SELECT schema.table.column FROM schema.table",
			wantCols: 1,
		},
		{
			name:     "four-level column (catalog.schema.table.column)",
			input:    "SELECT catalog.schema.table.column FROM catalog.schema.table",
			wantCols: 1,
		},
		{
			name:     "mixed levels",
			input:    "SELECT a, t.b, s.t.c, cat.s.t.d FROM t",
			wantCols: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			sel, ok := stmt.(*SelectStmt)
			if !ok {
				t.Fatalf("Expected SelectStmt, got %T", stmt)
			}
			if len(sel.Columns) != tt.wantCols {
				t.Errorf("Expected %d columns, got %d", tt.wantCols, len(sel.Columns))
			}

			// Round-trip test
			formatted := String(stmt)
			stmt2, err := Parse(formatted)
			if err != nil {
				t.Fatalf("Re-parse error: %v\nFormatted: %s", err, formatted)
			}
			formatted2 := String(stmt2)
			if formatted != formatted2 {
				t.Errorf("Round-trip mismatch:\nFirst:  %s\nSecond: %s", formatted, formatted2)
			}
		})
	}
}

func TestMultiLevelIdentifierParts(t *testing.T) {
	stmt, err := Parse("SELECT catalog.schema.table.column FROM db")
	if err != nil {
		t.Fatal(err)
	}

	sel := stmt.(*SelectStmt)
	ae := sel.Columns[0].(*AliasedExpr)
	col := ae.Expr.(*ColName)

	if len(col.Parts) != 4 {
		t.Fatalf("Expected 4 parts, got %d: %v", len(col.Parts), col.Parts)
	}

	// Test helper methods
	if col.Name() != "column" {
		t.Errorf("Name() = %q, want %q", col.Name(), "column")
	}
	if col.Table() != "table" {
		t.Errorf("Table() = %q, want %q", col.Table(), "table")
	}
	if col.Schema() != "schema" {
		t.Errorf("Schema() = %q, want %q", col.Schema(), "schema")
	}
	if col.Catalog() != "catalog" {
		t.Errorf("Catalog() = %q, want %q", col.Catalog(), "catalog")
	}
}

func TestMultiLevelTableName(t *testing.T) {
	stmt, err := Parse("SELECT * FROM catalog.schema.table")
	if err != nil {
		t.Fatal(err)
	}

	sel := stmt.(*SelectStmt)
	// From can be either *TableName directly or *AliasedTableExpr wrapping it
	var tn *TableName
	switch from := sel.From.(type) {
	case *TableName:
		tn = from
	case *AliasedTableExpr:
		tn = from.Expr.(*TableName)
	default:
		t.Fatalf("unexpected From type: %T", sel.From)
	}

	if len(tn.Parts) != 3 {
		t.Fatalf("Expected 3 parts, got %d: %v", len(tn.Parts), tn.Parts)
	}

	if tn.Name() != "table" {
		t.Errorf("Name() = %q, want %q", tn.Name(), "table")
	}
	if tn.Schema() != "schema" {
		t.Errorf("Schema() = %q, want %q", tn.Schema(), "schema")
	}
	if tn.Catalog() != "catalog" {
		t.Errorf("Catalog() = %q, want %q", tn.Catalog(), "catalog")
	}
}

func BenchmarkParseFormat(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
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
		stmt, _ := Parse(query)
		_ = String(stmt)
	}
}

func BenchmarkWalk(b *testing.B) {
	stmt, _ := Parse(`SELECT u.id, u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
GROUP BY u.id, u.name
ORDER BY order_count DESC`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Walk(stmt, func(node ast.Node) bool {
			return true
		})
	}
}
