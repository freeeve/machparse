package parser

import (
	"testing"

	"github.com/freeeve/machparse/ast"
)

func TestParseSelect(t *testing.T) {
	tests := []struct {
		input    string
		wantCols int
	}{
		{"SELECT * FROM users", 1},
		{"SELECT id, name FROM users", 2},
		{"SELECT id, name, email FROM users WHERE id = 1", 3},
		{"SELECT a.id, b.name FROM a JOIN b ON a.id = b.a_id", 2},
		{"SELECT COUNT(*) FROM users", 1},
		{"SELECT DISTINCT name FROM users", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			sel, ok := stmt.(*ast.SelectStmt)
			if !ok {
				t.Fatalf("Expected SelectStmt, got %T", stmt)
			}
			if len(sel.Columns) != tt.wantCols {
				t.Errorf("Expected %d columns, got %d", tt.wantCols, len(sel.Columns))
			}
		})
	}
}

func TestParseInsert(t *testing.T) {
	tests := []struct {
		input string
		want  int // expected number of value rows
	}{
		{"INSERT INTO users (id, name) VALUES (1, 'test')", 1},
		{"INSERT INTO users VALUES (1, 'test'), (2, 'test2')", 2},
		{"REPLACE INTO users (id) VALUES (1)", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			ins, ok := stmt.(*ast.InsertStmt)
			if !ok {
				t.Fatalf("Expected InsertStmt, got %T", stmt)
			}
			if len(ins.Values) != tt.want {
				t.Errorf("Expected %d value rows, got %d", tt.want, len(ins.Values))
			}
		})
	}
}

func TestParseUpdate(t *testing.T) {
	tests := []struct {
		input    string
		wantSets int
	}{
		{"UPDATE users SET name = 'test' WHERE id = 1", 1},
		{"UPDATE users SET name = 'test', email = 'a@b.com'", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			upd, ok := stmt.(*ast.UpdateStmt)
			if !ok {
				t.Fatalf("Expected UpdateStmt, got %T", stmt)
			}
			if len(upd.Set) != tt.wantSets {
				t.Errorf("Expected %d SET expressions, got %d", tt.wantSets, len(upd.Set))
			}
		})
	}
}

func TestParseDelete(t *testing.T) {
	tests := []struct {
		input    string
		hasWhere bool
	}{
		{"DELETE FROM users WHERE id = 1", true},
		{"DELETE FROM users", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			del, ok := stmt.(*ast.DeleteStmt)
			if !ok {
				t.Fatalf("Expected DeleteStmt, got %T", stmt)
			}
			if (del.Where != nil) != tt.hasWhere {
				t.Errorf("Expected hasWhere=%v, got %v", tt.hasWhere, del.Where != nil)
			}
		})
	}
}

func TestParseCreateTable(t *testing.T) {
	input := `CREATE TABLE users (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	p := New(input)
	stmt, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	create, ok := stmt.(*ast.CreateTableStmt)
	if !ok {
		t.Fatalf("Expected CreateTableStmt, got %T", stmt)
	}

	if create.Table.Name() != "users" {
		t.Errorf("Expected table name 'users', got %s", create.Table.Name())
	}

	if len(create.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(create.Columns))
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"SELECT 1 + 2"},
		{"SELECT a AND b OR c"},
		{"SELECT a = 1 AND b = 2"},
		{"SELECT a BETWEEN 1 AND 10"},
		{"SELECT a IN (1, 2, 3)"},
		{"SELECT a LIKE '%test%'"},
		{"SELECT a IS NULL"},
		{"SELECT a IS NOT NULL"},
		{"SELECT CASE WHEN a = 1 THEN 'one' ELSE 'other' END"},
		{"SELECT CAST(a AS INT)"},
		{"SELECT COUNT(*)"},
		{"SELECT SUM(amount)"},
		{"SELECT a::int"},
		{"SELECT a || b"},
		{"SELECT COALESCE(a, b, c)"},
		{"SELECT NULLIF(a, b)"},
		{"SELECT EXISTS (SELECT 1 FROM t)"},
		{"SELECT * FROM t WHERE a IN (SELECT id FROM t2)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if stmt == nil {
				t.Fatal("Expected statement, got nil")
			}
		})
	}
}

func TestParseJoins(t *testing.T) {
	tests := []string{
		"SELECT * FROM a JOIN b ON a.id = b.a_id",
		"SELECT * FROM a INNER JOIN b ON a.id = b.a_id",
		"SELECT * FROM a LEFT JOIN b ON a.id = b.a_id",
		"SELECT * FROM a LEFT OUTER JOIN b ON a.id = b.a_id",
		"SELECT * FROM a RIGHT JOIN b ON a.id = b.a_id",
		"SELECT * FROM a FULL OUTER JOIN b ON a.id = b.a_id",
		"SELECT * FROM a CROSS JOIN b",
		"SELECT * FROM a NATURAL JOIN b",
		"SELECT * FROM a JOIN b USING (id)",
		"SELECT * FROM a, b WHERE a.id = b.a_id",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			p := New(input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if stmt == nil {
				t.Fatal("Expected statement, got nil")
			}
		})
	}
}

func TestParseWithCTE(t *testing.T) {
	input := `WITH active_users AS (
		SELECT id, name FROM users WHERE status = 'active'
	)
	SELECT * FROM active_users WHERE name LIKE 'A%'`

	p := New(input)
	stmt, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	sel, ok := stmt.(*ast.SelectStmt)
	if !ok {
		t.Fatalf("Expected SelectStmt, got %T", stmt)
	}

	if sel.With == nil {
		t.Fatal("Expected WITH clause")
	}

	if len(sel.With.CTEs) != 1 {
		t.Errorf("Expected 1 CTE, got %d", len(sel.With.CTEs))
	}
}

func TestParseWindowFunctions(t *testing.T) {
	tests := []string{
		"SELECT ROW_NUMBER() OVER () FROM t",
		"SELECT ROW_NUMBER() OVER (ORDER BY id) FROM t",
		"SELECT ROW_NUMBER() OVER (PARTITION BY type ORDER BY id) FROM t",
		"SELECT SUM(amount) OVER (PARTITION BY user_id) FROM orders",
		"SELECT AVG(price) OVER (ORDER BY date ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING) FROM prices",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			p := New(input)
			stmt, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if stmt == nil {
				t.Fatal("Expected statement, got nil")
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
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
		p := New(input)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseSimple(b *testing.B) {
	input := "SELECT * FROM users WHERE id = 1"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := New(input)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}
