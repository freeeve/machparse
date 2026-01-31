package machparse

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var benchQueries = map[string]string{
	"simple":    "SELECT 1",
	"columns":   "SELECT id, name, email, created_at FROM users",
	"where":     "SELECT * FROM users WHERE status = 'active' AND age > 18",
	"join":      "SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id",
	"subquery":  "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
	"aggregate": "SELECT status, COUNT(*), AVG(age) FROM users GROUP BY status HAVING COUNT(*) > 10",
	"complex": `SELECT u.id, u.name, COUNT(o.id) as order_count, SUM(o.total) as total_spent
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active' AND u.created_at > '2024-01-01'
		GROUP BY u.id, u.name
		HAVING COUNT(o.id) > 5
		ORDER BY total_spent DESC
		LIMIT 100`,
	"cte": `WITH active_users AS (
		SELECT id, name FROM users WHERE status = 'active'
	), user_orders AS (
		SELECT user_id, COUNT(*) as cnt FROM orders GROUP BY user_id
	)
	SELECT a.id, a.name, COALESCE(o.cnt, 0) as orders
	FROM active_users a
	LEFT JOIN user_orders o ON a.id = o.user_id`,
	"insert":  "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com')",
	"update":  "UPDATE users SET name = 'Jane', updated_at = NOW() WHERE id = 1",
	"delete":  "DELETE FROM users WHERE status = 'deleted' AND updated_at < '2024-01-01'",
	"create":  "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) UNIQUE)",
	"window":  "SELECT id, name, ROW_NUMBER() OVER (PARTITION BY status ORDER BY created_at DESC) as rn FROM users",
	"case":    "SELECT id, CASE WHEN status = 1 THEN 'active' WHEN status = 2 THEN 'pending' ELSE 'unknown' END FROM users",
	"like":    "SELECT * FROM users WHERE name LIKE '%john%' OR email LIKE '%@gmail.com'",
	"between": "SELECT * FROM orders WHERE created_at BETWEEN '2024-01-01' AND '2024-12-31' AND total BETWEEN 100 AND 1000",
	"cast":    "SELECT CAST(id AS VARCHAR), CAST(price AS DECIMAL(10,2)), id::text FROM products",
	"func":    "SELECT COALESCE(name, 'unknown'), UPPER(email), LENGTH(description), SUBSTRING(name FROM 1 FOR 10) FROM users",

	// Complex queries for stress testing
	"nested_subquery": `SELECT * FROM users u
		WHERE u.id IN (
			SELECT o.user_id FROM orders o
			WHERE o.total > (SELECT AVG(total) FROM orders WHERE status = 'completed')
			AND o.created_at > (SELECT MAX(created_at) FROM orders WHERE user_id = u.id AND status = 'cancelled')
		)`,

	"deep_nested": `SELECT * FROM (
		SELECT * FROM (
			SELECT * FROM (
				SELECT id, name FROM users WHERE status = 'active'
			) t1 WHERE id > 100
		) t2 WHERE name LIKE 'A%'
	) t3 LIMIT 10`,

	"union_complex": `SELECT id, name, 'user' as type FROM users WHERE active = true
		UNION ALL
		SELECT id, title, 'product' as type FROM products WHERE in_stock = true
		UNION
		SELECT id, name, 'category' as type FROM categories
		ORDER BY type, name
		LIMIT 100`,

	"multi_join": `SELECT
			u.id, u.name, u.email,
			o.id as order_id, o.total,
			p.name as product_name, p.price,
			c.name as category,
			s.name as supplier
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		INNER JOIN order_items oi ON o.id = oi.order_id
		INNER JOIN products p ON oi.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN suppliers s ON p.supplier_id = s.id
		WHERE o.status = 'completed' AND o.created_at > '2024-01-01'`,

	"recursive_cte": `WITH RECURSIVE subordinates AS (
			SELECT id, name, manager_id, 1 as level
			FROM employees
			WHERE manager_id IS NULL
			UNION ALL
			SELECT e.id, e.name, e.manager_id, s.level + 1
			FROM employees e
			INNER JOIN subordinates s ON e.manager_id = s.id
		)
		SELECT * FROM subordinates ORDER BY level, name`,

	"complex_aggregation": `SELECT
			DATE_TRUNC('month', o.created_at) as month,
			c.name as category,
			COUNT(DISTINCT o.id) as order_count,
			COUNT(DISTINCT o.user_id) as unique_customers,
			SUM(oi.quantity) as total_items,
			SUM(oi.quantity * p.price) as gross_revenue,
			AVG(o.total) as avg_order_value,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY o.total) as median_order
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		JOIN categories c ON p.category_id = c.id
		WHERE o.status = 'completed'
		GROUP BY DATE_TRUNC('month', o.created_at), c.name
		HAVING COUNT(DISTINCT o.id) > 10
		ORDER BY month DESC, gross_revenue DESC`,

	"window_complex": `SELECT
			u.id, u.name,
			o.total,
			SUM(o.total) OVER (PARTITION BY u.id ORDER BY o.created_at) as running_total,
			AVG(o.total) OVER (PARTITION BY u.id) as avg_order,
			ROW_NUMBER() OVER (PARTITION BY u.id ORDER BY o.created_at DESC) as order_rank,
			LAG(o.total) OVER (PARTITION BY u.id ORDER BY o.created_at) as prev_order,
			LEAD(o.total) OVER (PARTITION BY u.id ORDER BY o.created_at) as next_order,
			FIRST_VALUE(o.total) OVER (PARTITION BY u.id ORDER BY o.created_at
				ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) as first_order
		FROM users u
		JOIN orders o ON u.id = o.user_id`,

	"correlated_subquery": `SELECT u.id, u.name,
			(SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as order_count,
			(SELECT MAX(total) FROM orders o WHERE o.user_id = u.id) as max_order,
			(SELECT AVG(total) FROM orders o WHERE o.user_id = u.id AND o.status = 'completed') as avg_completed
		FROM users u
		WHERE EXISTS (SELECT 1 FROM orders WHERE user_id = u.id AND total > 1000)`,

	"lateral_join": `SELECT u.id, u.name, recent.*
		FROM users u
		CROSS JOIN LATERAL (
			SELECT o.id, o.total, o.created_at
			FROM orders o
			WHERE o.user_id = u.id
			ORDER BY o.created_at DESC
			LIMIT 3
		) recent
		WHERE u.status = 'active'`,
}

func BenchmarkParseByQuery(b *testing.B) {
	for name, query := range benchQueries {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(query)
			}
		})
	}
}

func BenchmarkFormatByQuery(b *testing.B) {
	stmts := make(map[string]Statement)
	for name, query := range benchQueries {
		stmt, _ := Parse(query)
		stmts[name] = stmt
	}

	for name, stmt := range stmts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = String(stmt)
			}
		})
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	for name, query := range benchQueries {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				stmt, _ := Parse(query)
				_ = String(stmt)
			}
		})
	}
}

// Benchmark to measure parsing throughput
func BenchmarkParseThroughput(b *testing.B) {
	// Mix of queries representing typical workload
	queries := []string{
		"SELECT * FROM users WHERE id = 1",
		"SELECT id, name FROM users WHERE status = 'active'",
		"INSERT INTO logs (msg) VALUES ('test')",
		"UPDATE users SET last_login = NOW() WHERE id = 1",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			_, _ = Parse(q)
		}
	}
}

// Benchmark with AST release - demonstrates benefit of returning nodes to pool
func BenchmarkParseWithRelease(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
		FROM users u LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
		GROUP BY u.id, u.name
		ORDER BY order_count DESC`

	// Warm up pools
	for i := 0; i < 100; i++ {
		stmt, _ := Parse(query)
		Repool(stmt)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stmt, _ := Parse(query)
		Repool(stmt)
	}
}

// Benchmark without AST release - nodes are garbage collected
func BenchmarkParseWithoutRelease(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
		FROM users u LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
		GROUP BY u.id, u.name
		ORDER BY order_count DESC`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Parse(query)
	}
}

// Benchmark large queries with many clauses
func BenchmarkParseLargeQueries(b *testing.B) {
	// Generate queries with varying sizes
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		// Large IN list
		b.Run(fmt.Sprintf("in_list_%d", size), func(b *testing.B) {
			query := generateInList(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(query)
			}
		})

		// Large WHERE clause chain
		b.Run(fmt.Sprintf("where_chain_%d", size), func(b *testing.B) {
			query := generateWhereChain(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(query)
			}
		})

		// Large SELECT column list
		b.Run(fmt.Sprintf("columns_%d", size), func(b *testing.B) {
			query := generateColumnList(size)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(query)
			}
		})
	}
}

// generateInList creates SELECT * FROM t WHERE id IN (0, 1, 2, ..., n-1)
func generateInList(n int) string {
	var b strings.Builder
	b.WriteString("SELECT * FROM t WHERE id IN (")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(i))
	}
	b.WriteString(")")
	return b.String()
}

// generateWhereChain creates SELECT * FROM t WHERE a0 = 0 AND a1 = 1 AND ... AND an = n
func generateWhereChain(n int) string {
	var b strings.Builder
	b.WriteString("SELECT * FROM t WHERE ")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString("a")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" = ")
		b.WriteString(strconv.Itoa(i))
	}
	return b.String()
}

// generateColumnList creates SELECT col0, col1, ..., coln FROM t
func generateColumnList(n int) string {
	var b strings.Builder
	b.WriteString("SELECT ")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("col")
		b.WriteString(strconv.Itoa(i))
	}
	b.WriteString(" FROM t")
	return b.String()
}

// Benchmark lexer separately
func BenchmarkLexerOnly(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
		FROM users u LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
		GROUP BY u.id, u.name
		ORDER BY order_count DESC`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// We can't easily benchmark lexer alone without importing it
		// This is a proxy - parse includes lexing
		_, _ = Parse(query)
	}
}
