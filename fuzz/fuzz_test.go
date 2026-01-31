package fuzz

import (
	"testing"

	"github.com/freeeve/machparse"
	"github.com/freeeve/machparse/lexer"
	"github.com/freeeve/machparse/token"
)

// FuzzParse tests that the parser doesn't panic on arbitrary input.
func FuzzParse(f *testing.F) {
	// Add seed corpus of valid SQL
	seeds := []string{
		// Basic SELECT
		"SELECT * FROM users",
		"SELECT id, name FROM users WHERE status = 'active'",
		"SELECT a.id, b.name FROM a JOIN b ON a.id = b.a_id",
		"SELECT DISTINCT a, b FROM t",
		"SELECT ALL * FROM t",

		// DML
		"INSERT INTO users (id, name) VALUES (1, 'test')",
		"INSERT INTO t (a, b) VALUES (1, 2), (3, 4), (5, 6)",
		"UPDATE users SET name = 'new' WHERE id = 1",
		"UPDATE t SET a = 1, b = 2, c = 3 WHERE x > 0",
		"DELETE FROM users WHERE id = 1",
		"DELETE FROM t USING t2 WHERE t.id = t2.id",

		// Subqueries
		"SELECT * FROM users WHERE id IN (SELECT user_id FROM orders)",
		"SELECT * FROM (SELECT 1 FROM t) AS sub",
		"SELECT (SELECT MAX(id) FROM t2) FROM t",
		"SELECT * FROM t WHERE EXISTS (SELECT 1 FROM u WHERE u.id = t.id)",

		// CTE
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
		"WITH cte (a, b) AS (SELECT 1, 2) SELECT * FROM cte",
		"WITH RECURSIVE cte AS (SELECT 1 UNION ALL SELECT n+1 FROM cte WHERE n < 10) SELECT * FROM cte",
		"WITH cte1 AS (SELECT 1), cte2 AS (SELECT 2) SELECT * FROM cte1, cte2",

		// Window functions
		"SELECT COUNT(*) OVER (PARTITION BY type ORDER BY id) FROM items",
		"SELECT ROW_NUMBER() OVER () FROM t",
		"SELECT SUM(x) OVER (ORDER BY y ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING) FROM t",
		"SELECT AVG(x) OVER (PARTITION BY a ORDER BY b RANGE UNBOUNDED PRECEDING) FROM t",
		"SELECT RANK() OVER (ORDER BY score DESC) FROM t",
		"SELECT SUM(x) FILTER (WHERE y > 0) FROM t",

		// CASE expressions
		"SELECT CASE WHEN x = 1 THEN 'a' ELSE 'b' END FROM t",
		"SELECT CASE x WHEN 1 THEN 'one' WHEN 2 THEN 'two' END FROM t",
		"SELECT CASE WHEN a THEN 1 WHEN b THEN 2 WHEN c THEN 3 ELSE 0 END FROM t",

		// DDL
		"CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255))",
		"CREATE TABLE t (id INT NOT NULL, name TEXT DEFAULT 'x', UNIQUE(id))",
		"CREATE TABLE IF NOT EXISTS t (id INT)",
		"CREATE TEMPORARY TABLE tmp (a INT)",
		"ALTER TABLE users ADD COLUMN email VARCHAR(255)",
		"ALTER TABLE t DROP COLUMN a",
		"ALTER TABLE t MODIFY COLUMN a VARCHAR(100)",
		"DROP TABLE IF EXISTS users CASCADE",
		"DROP TABLE t RESTRICT",

		// Indexes
		"CREATE INDEX idx ON t (a, b)",
		"CREATE UNIQUE INDEX idx ON t (a)",
		"DROP INDEX idx ON t",

		// Clauses
		"SELECT * FROM users LIMIT 10 OFFSET 20",
		"SELECT * FROM t LIMIT 10, 20",
		"SELECT * FROM t ORDER BY a ASC, b DESC",
		"SELECT * FROM t ORDER BY a NULLS FIRST",
		"SELECT * FROM t GROUP BY a HAVING COUNT(*) > 1",
		"SELECT * FROM t GROUP BY a, b, c",

		// Functions
		"SELECT COALESCE(a, b, c) FROM t",
		"SELECT NULLIF(a, b) FROM t",
		"SELECT GREATEST(a, b, c) FROM t",
		"SELECT LEAST(a, b, c) FROM t",
		"SELECT CAST(x AS INT) FROM t",
		"SELECT EXTRACT(YEAR FROM date_col) FROM t",
		"SELECT EXTRACT(MONTH FROM ts) FROM t",
		"SELECT TRIM(BOTH ' ' FROM name) FROM t",
		"SELECT TRIM(LEADING 'x' FROM name) FROM t",
		"SELECT SUBSTRING(name FROM 1 FOR 10) FROM t",
		"SELECT POSITION('x' IN name) FROM t",
		"SELECT OVERLAY(a PLACING b FROM 1 FOR 2) FROM t",

		// Operators
		"SELECT * FROM t WHERE a BETWEEN 1 AND 10",
		"SELECT * FROM t WHERE a NOT BETWEEN 1 AND 10",
		"SELECT * FROM t WHERE name LIKE '%test%'",
		"SELECT * FROM t WHERE name LIKE '%x%' ESCAPE '#'",
		"SELECT * FROM t WHERE name ILIKE '%TEST%'",
		"SELECT * FROM t WHERE a IN (1, 2, 3)",
		"SELECT * FROM t WHERE a NOT IN (1, 2, 3)",
		"SELECT * FROM t WHERE a IS NULL",
		"SELECT * FROM t WHERE a IS NOT NULL",
		"SELECT * FROM t WHERE a IS TRUE",
		"SELECT * FROM t WHERE a IS NOT FALSE",
		"SELECT * FROM t WHERE a IS DISTINCT FROM b",

		// Arithmetic and boolean
		"SELECT 1 + 2 * 3 - 4 / 5",
		"SELECT a % b FROM t",
		"SELECT NOT a AND b OR c FROM t",
		"SELECT a AND b AND c OR d OR e FROM t",
		"SELECT -1, +2, ~3 FROM t",
		"SELECT a || b FROM t",
		"SELECT a & b, a | b, a ^ b FROM t",

		// JOINs
		"SELECT * FROM t1 NATURAL JOIN t2",
		"SELECT * FROM t1 LEFT OUTER JOIN t2 ON t1.id = t2.id",
		"SELECT * FROM t1 RIGHT JOIN t2 ON a = b",
		"SELECT * FROM t1 FULL OUTER JOIN t2 USING (id)",
		"SELECT * FROM t1 CROSS JOIN t2",
		"SELECT * FROM t1 JOIN t2 ON a = b JOIN t3 ON c = d",
		"SELECT * FROM t1, t2, t3",
		"SELECT * FROM (t1, t2) JOIN t3 ON a = b",

		// Set operations
		"SELECT 1 UNION SELECT 2",
		"SELECT 1 UNION ALL SELECT 2",
		"SELECT 1 INTERSECT SELECT 2",
		"SELECT 1 EXCEPT SELECT 2",
		"(SELECT 1) UNION (SELECT 2)",
		"SELECT 1 UNION SELECT 2 UNION ALL SELECT 3",
		"(SELECT 1 FROM t) UNION (SELECT 2 FROM t) ORDER BY 1",

		// Locking
		"SELECT * FROM t FOR UPDATE",
		"SELECT * FROM t FOR SHARE NOWAIT",
		"SELECT * FROM t FOR UPDATE SKIP LOCKED",

		// PostgreSQL specific
		"SELECT ARRAY[1, 2, 3]",
		"SELECT a::int FROM t",
		"SELECT a::varchar(100) FROM t",
		"SELECT a->>'key' FROM t",
		"SELECT a->'nested'->'key' FROM t",
		"SELECT a#>'{a,b}' FROM t",
		"INSERT INTO t (a) VALUES (1) RETURNING id",
		"INSERT INTO t (a) VALUES (1) RETURNING *",
		"INSERT INTO t (a) VALUES (1) ON CONFLICT (a) DO NOTHING",
		"INSERT INTO t (a) VALUES (1) ON CONFLICT (a) DO UPDATE SET b = 2",
		"UPDATE t SET a = 1 RETURNING *",
		"DELETE FROM t WHERE a = 1 RETURNING id",
		"SELECT * FROM t WHERE a = ANY(ARRAY[1,2,3])",

		// MySQL specific
		"REPLACE INTO t (a, b) VALUES (1, 2)",
		"INSERT INTO t (a) VALUES (1) ON DUPLICATE KEY UPDATE a = 2",
		"SELECT `column` FROM `table`",
		"SELECT * FROM t LIMIT 20, 10",

		// SQL Server specific - bracket identifiers
		"SELECT [column name] FROM [table name]",
		"SELECT [my column] FROM [my table]",
		"SELECT [a], [b], [c] FROM [t]",
		"SELECT [schema].[table].[column] FROM [schema].[table]",

		// SQL Server specific - temp tables
		"SELECT * FROM #temp",
		"SELECT * FROM ##global_temp",
		"SELECT a, b FROM #temp_table WHERE x > 0",
		"INSERT INTO #temp (a) VALUES (1)",
		"SELECT [col] FROM #temp_table",

		// SQL Server TOP
		"SELECT top(10) * FROM t",
		"SELECT * FROM t WITH (nolock)",

		// Oracle specific
		"SELECT * FROM t WHERE rownum <= 10",
		"SELECT sysdate FROM dual",
		"SELECT systimestamp FROM dual",
		"SELECT 1 FROM dual",
		"SELECT * FROM t CONNECT BY prior id = parent_id",
		"SELECT * FROM t START WITH parent_id IS NULL CONNECT BY prior id = parent_id",

		// Multi-level identifiers
		"SELECT * FROM schema.table",
		"SELECT schema.table.column FROM schema.table",
		"SELECT catalog.schema.table.column FROM catalog.schema.table",
		"SELECT a.b.c.d.e FROM a.b.c.d",

		// Qualified stars
		"SELECT t.* FROM t",
		"SELECT a.b.* FROM a.b",

		// Parameters and variables
		"SELECT @var := 1",
		"SELECT @@global_var",
		"SELECT $1, $2 FROM t",
		"SELECT :name FROM t",
		"SELECT ? FROM t WHERE a = ?",

		// Comments
		"SELECT /* comment */ * FROM t",
		"SELECT * FROM t -- line comment",
		"SELECT /* multi\nline\ncomment */ 1",
		"SELECT 1 /* inline */ + 2 FROM t",

		// Literals
		"SELECT 1e10, 1.5e-3, .5 FROM t",
		"SELECT 0x1A, 0b1010 FROM t",
		"SELECT TRUE, FALSE, NULL FROM t",
		"SELECT 'string', 'with''escape' FROM t",
		`SELECT "double quoted" FROM t`,

		// Edge cases
		"",
		" ",
		";;;",
		"SELECT 1",
		"(SELECT 1)",
		"((SELECT 1))",
		"SELECT ((a + b) * (c - d)) FROM t",

		// Misc
		"TRUNCATE TABLE t",
		"EXPLAIN SELECT * FROM t",
		"EXPLAIN ANALYZE SELECT * FROM t",

		// Regression tests from fuzz findings
		"SELECT A(*IN",
		"SELECT A(*IS",
		"SELECT A(*BETWEEN",
		"SELECT A(*LIKE",
		"SELECT A(*SIMILAR",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		// The parser should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parse panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		stmt, err := machparse.Parse(sql)
		if err != nil {
			// Parse errors are acceptable
			return
		}

		if stmt == nil {
			return
		}

		// Test formatting
		formatted := machparse.String(stmt)
		if formatted == "" {
			t.Errorf("Formatted output is empty for valid parse of: %q", sql)
			return
		}

		// Round-trip test: re-parse the formatted output
		stmt2, err := machparse.Parse(formatted)
		if err != nil {
			t.Errorf("Re-parse failed:\nOriginal: %q\nFormatted: %q\nError: %v", sql, formatted, err)
			return
		}

		if stmt2 == nil {
			t.Errorf("Re-parse returned nil for: %q", formatted)
			return
		}

		// Format again should be identical (idempotent)
		formatted2 := machparse.String(stmt2)
		if formatted != formatted2 {
			t.Errorf("Round-trip mismatch:\nOriginal: %q\nFirst:    %q\nSecond:   %q", sql, formatted, formatted2)
		}
	})
}

// FuzzLexer tests that the lexer doesn't panic on arbitrary input.
func FuzzLexer(f *testing.F) {
	// Add seed corpus
	seeds := []string{
		// Basic SQL
		"SELECT * FROM users",
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET a = 1",
		"DELETE FROM t",

		// String literals
		"'string with ''escapes'''",
		"'multi\nline\nstring'",
		`"double quoted"`,
		`"with ""escape"""`,
		"`backtick quoted`",
		"`with ``escape```",

		// Dollar quoting (PostgreSQL)
		"$$dollar$$",
		"$tag$content$tag$",
		"$abc$nested$$dollar$$inside$abc$",

		// Comments
		"-- line comment\nSELECT 1",
		"/* block comment */ SELECT 1",
		"/* nested /* comment */ */",
		"# mysql line comment\nSELECT 1",

		// Numbers
		"1.5e-10",
		"1.5E+10",
		".5",
		"5.",
		"0x1A2B",
		"0X1a2b",
		"0b1010",
		"0B1010",
		"123456789",
		"123.456.789",

		// Parameters
		":named_param",
		"$1",
		"$123",
		"@variable",
		"@@global",
		"?",

		// Operators
		"a->>'b'",
		"a->>b",
		"a->b",
		"a#>'{a,b}'",
		"a#>>'{a,b}'",
		"a @@ b",
		"a <-> b",
		"a::int",
		"a::varchar(100)",
		"a <> b",
		"a != b",
		"a <= b",
		"a >= b",
		"a << b",
		"a >> b",
		"a || b",
		"a && b",

		// SQL Server brackets
		"[identifier]",
		"[with spaces]",
		"[with]]bracket]",
		"[schema].[table].[column]",

		// SQL Server temp tables
		"#temp",
		"##global_temp",
		"#temp_table",

		// Unicode and special chars
		"",
		"\x00\x01\x02",
		"SELECT\t\n\r *",
		"SELECT\u00A0*",
		"идентификатор",
		"表名",

		// Edge cases
		"...",
		"::::",
		";;;;",
		"((()))",
		"[[[",
		"]]]",
		"/**/",
		"--\n",
		"''",
		`""`,
		"``",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Lexer panicked on input: %q\npanic: %v", input, r)
			}
		}()

		l := lexer.New(input)
		for {
			tok := l.Next()
			if tok.Type == token.EOF {
				break
			}
			if tok.Type == token.ILLEGAL {
				// Illegal tokens are acceptable
				continue
			}
		}
	})
}

// FuzzParseAll tests parsing multiple statements.
func FuzzParseAll(f *testing.F) {
	seeds := []string{
		"SELECT 1; SELECT 2",
		"SELECT 1; SELECT 2; SELECT 3",
		"INSERT INTO t VALUES (1); UPDATE t SET a = 1",
		";;;",
		"SELECT 1;;SELECT 2",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ParseAll panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		stmts, err := machparse.ParseAll(sql)
		if err != nil {
			return
		}

		for _, stmt := range stmts {
			if stmt != nil {
				_ = machparse.String(stmt)
			}
		}
	})
}

// FuzzWalk tests walking the AST.
func FuzzWalk(f *testing.F) {
	seeds := []string{
		"SELECT a, b FROM t WHERE c = 1",
		"SELECT * FROM a JOIN b ON a.id = b.id",
		"SELECT (SELECT 1) FROM t",
		"SELECT CASE WHEN a THEN b ELSE c END FROM t",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Walk panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		stmt, err := machparse.Parse(sql)
		if err != nil {
			return
		}
		if stmt == nil {
			return
		}

		// Walk the entire tree
		count := 0
		machparse.Walk(stmt, func(n machparse.Node) bool {
			count++
			return true
		})

		// Walk with early termination
		machparse.Walk(stmt, func(n machparse.Node) bool {
			return count < 5
		})
	})
}

// FuzzRewrite tests rewriting the AST.
func FuzzRewrite(f *testing.F) {
	seeds := []string{
		"SELECT a FROM t",
		"SELECT a, b, c FROM t WHERE d = 1",
		"UPDATE t SET a = 1 WHERE b = 2",
		"DELETE FROM t WHERE x IN (1, 2, 3)",
		"INSERT INTO t (a, b) VALUES (1, 2)",
		"SELECT * FROM t1 JOIN t2 ON t1.id = t2.id",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Rewrite panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		stmt, err := machparse.Parse(sql)
		if err != nil {
			return
		}
		if stmt == nil {
			return
		}

		// Identity rewrite
		rewritten := machparse.Rewrite(stmt, func(n machparse.Node) machparse.Node {
			return n
		})

		if rewritten == nil {
			t.Errorf("Rewrite returned nil for valid input: %q", sql)
			return
		}

		// Format the rewritten AST
		_ = machparse.String(rewritten)
	})
}

// FuzzFormat tests that formatted SQL is valid and can be re-parsed identically.
// This is a comprehensive round-trip test focusing on format stability.
func FuzzFormat(f *testing.F) {
	seeds := []string{
		// SELECT variations
		"SELECT * FROM t",
		"SELECT a, b, c FROM t",
		"SELECT DISTINCT a FROM t",
		"SELECT t.* FROM t",
		"SELECT a AS alias FROM t",

		// Complex expressions
		"SELECT 1 + 2 * 3 FROM t",
		"SELECT (a + b) * c FROM t",
		"SELECT -a, +b, NOT c FROM t",
		"SELECT a AND b OR c FROM t",
		"SELECT a BETWEEN 1 AND 10 FROM t",
		"SELECT CASE WHEN a THEN b ELSE c END FROM t",

		// JOINs
		"SELECT * FROM t1 JOIN t2 ON t1.id = t2.id",
		"SELECT * FROM t1 LEFT JOIN t2 ON a = b",
		"SELECT * FROM t1, t2 WHERE t1.id = t2.id",

		// Subqueries
		"SELECT * FROM (SELECT 1) AS sub",
		"SELECT * FROM t WHERE a IN (SELECT b FROM u)",
		"SELECT (SELECT 1) FROM t",

		// CTEs
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
		"WITH cte (a) AS (SELECT 1) SELECT * FROM cte",

		// Set operations
		"SELECT 1 UNION SELECT 2",
		"SELECT 1 UNION ALL SELECT 2",
		"(SELECT 1) UNION (SELECT 2)",

		// Clauses
		"SELECT * FROM t WHERE a = 1",
		"SELECT * FROM t ORDER BY a",
		"SELECT * FROM t GROUP BY a HAVING COUNT(*) > 1",
		"SELECT * FROM t LIMIT 10 OFFSET 5",

		// DML
		"INSERT INTO t (a) VALUES (1)",
		"INSERT INTO t (a, b) VALUES (1, 2), (3, 4)",
		"UPDATE t SET a = 1 WHERE b = 2",
		"DELETE FROM t WHERE a = 1",

		// DDL
		"CREATE TABLE t (id INT PRIMARY KEY)",
		"ALTER TABLE t ADD COLUMN a INT",
		"DROP TABLE t",

		// Multi-level identifiers
		"SELECT schema.table.column FROM schema.table",
		"SELECT a.b.c.d FROM a.b.c",

		// SQL Server brackets
		"SELECT [column] FROM [table]",
		"SELECT [a].[b] FROM [a]",

		// Temp tables
		"SELECT * FROM #temp",
		"SELECT * FROM ##global",

		// PostgreSQL
		"SELECT a::int FROM t",
		"INSERT INTO t (a) VALUES (1) RETURNING id",
		"SELECT ARRAY[1, 2, 3]",

		// Window functions
		"SELECT SUM(a) OVER (PARTITION BY b ORDER BY c) FROM t",
		"SELECT ROW_NUMBER() OVER () FROM t",

		// Functions
		"SELECT COUNT(*) FROM t",
		"SELECT COALESCE(a, b, c) FROM t",
		"SELECT CAST(a AS INT) FROM t",
		"SELECT EXTRACT(YEAR FROM d) FROM t",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Format panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		// First parse
		stmt, err := machparse.Parse(sql)
		if err != nil {
			return
		}
		if stmt == nil {
			return
		}

		// Format to SQL string
		formatted1 := machparse.String(stmt)
		if formatted1 == "" {
			t.Errorf("Format produced empty string for: %q", sql)
			return
		}

		// Re-parse the formatted output
		stmt2, err := machparse.Parse(formatted1)
		if err != nil {
			t.Errorf("Re-parse failed:\nOriginal:  %q\nFormatted: %q\nError: %v", sql, formatted1, err)
			return
		}
		if stmt2 == nil {
			t.Errorf("Re-parse returned nil for: %q", formatted1)
			return
		}

		// Format again - must be identical (idempotent)
		formatted2 := machparse.String(stmt2)
		if formatted1 != formatted2 {
			t.Errorf("Format not idempotent:\nOriginal:   %q\nFormatted1: %q\nFormatted2: %q", sql, formatted1, formatted2)
			return
		}

		// Third round-trip to be extra sure
		stmt3, err := machparse.Parse(formatted2)
		if err != nil {
			t.Errorf("Third parse failed: %v", err)
			return
		}
		formatted3 := machparse.String(stmt3)
		if formatted2 != formatted3 {
			t.Errorf("Format not stable after 3 rounds:\nRound2: %q\nRound3: %q", formatted2, formatted3)
		}
	})
}

// FuzzPooling tests that AST pooling works correctly.
// Nodes can be returned to the pool via Repool() for reuse.
func FuzzPooling(f *testing.F) {
	seeds := []string{
		"SELECT * FROM t",
		"SELECT a, b FROM t WHERE c = 1",
		"INSERT INTO t (a) VALUES (1)",
		"UPDATE t SET a = 1",
		"DELETE FROM t",
		"SELECT * FROM t1 JOIN t2 ON t1.id = t2.id",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
		"SELECT CASE WHEN a THEN b END FROM t",
		"CREATE TABLE t (id INT)",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Pooling panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		// Parse and use the statement
		stmt, err := machparse.Parse(sql)
		if err != nil {
			return
		}
		if stmt == nil {
			return
		}

		// Format while AST is still valid
		formatted := machparse.String(stmt)

		// Walk the AST before returning to pool
		machparse.Walk(stmt, func(n machparse.Node) bool {
			return true
		})

		// Return AST nodes to pool for reuse
		machparse.Repool(stmt)

		// Parse the formatted output (uses recycled nodes)
		stmt2, err := machparse.Parse(formatted)
		if err != nil && formatted != "" {
			t.Errorf("Re-parse after repool failed:\nOriginal:  %q\nFormatted: %q\nError: %v", sql, formatted, err)
			return
		}

		if stmt2 != nil {
			formatted2 := machparse.String(stmt2)
			if formatted != formatted2 {
				t.Errorf("Parse after repool different:\nFirst:  %q\nSecond: %q", formatted, formatted2)
			}
			// Return the second statement to pool as well
			machparse.Repool(stmt2)
		}
	})
}

// FuzzDialects tests dialect-specific SQL features.
func FuzzDialects(f *testing.F) {
	seeds := []string{
		// PostgreSQL
		"SELECT a::int FROM t",
		"SELECT a::varchar(100)::text FROM t",
		"SELECT ARRAY[1, 2, 3]",
		"SELECT a->'b'->>'c' FROM t",
		"SELECT a#>'{a,b}' FROM t",
		"INSERT INTO t VALUES (1) ON CONFLICT DO NOTHING",
		"INSERT INTO t VALUES (1) ON CONFLICT (a) DO UPDATE SET b = 2",
		"SELECT * FROM t RETURNING *",

		// MySQL
		"SELECT `column` FROM `table`",
		"SELECT * FROM t LIMIT 10, 20",
		"REPLACE INTO t VALUES (1)",
		"INSERT INTO t VALUES (1) ON DUPLICATE KEY UPDATE a = 2",

		// SQL Server
		"SELECT [column] FROM [table]",
		"SELECT [a b c] FROM [d e f]",
		"SELECT * FROM #temp",
		"SELECT * FROM ##global_temp",
		"SELECT [col] FROM #temp WHERE [x] > 0",
		"SELECT top(10) * FROM t",

		// Oracle
		"SELECT * FROM dual",
		"SELECT sysdate FROM dual",
		"SELECT * FROM t WHERE rownum <= 10",
		"SELECT * FROM t CONNECT BY prior id = parent_id",
		"SELECT * FROM t START WITH id = 1 CONNECT BY prior id = parent_id",

		// Mixed dialects (parser should handle gracefully)
		"SELECT [a], `b`, \"c\" FROM t",
		"SELECT #temp.a FROM #temp",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sql string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Dialect fuzz panicked on input: %q\npanic: %v", sql, r)
			}
		}()

		stmt, err := machparse.Parse(sql)
		if err != nil {
			return
		}
		if stmt == nil {
			return
		}

		// Format and round-trip
		formatted := machparse.String(stmt)
		if formatted == "" {
			return
		}

		stmt2, err := machparse.Parse(formatted)
		if err != nil {
			t.Errorf("Dialect round-trip failed:\nOriginal:  %q\nFormatted: %q\nError: %v", sql, formatted, err)
			return
		}

		formatted2 := machparse.String(stmt2)
		if formatted != formatted2 {
			t.Errorf("Dialect format not stable:\nFirst:  %q\nSecond: %q", formatted, formatted2)
		}
	})
}
