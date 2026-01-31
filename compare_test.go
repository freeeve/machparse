//go:build ignore

// To run vitess comparison benchmarks:
// 1. Uncomment the vitess require in go.mod
// 2. Change build tag above to: //go:build compare_vitess
// 3. Run: go test -tags=compare_vitess -bench=Compare

package machparse

import (
	"testing"

	vitess "github.com/blastrain/vitess-sqlparser/sqlparser"
)

// Comparative benchmarks between machparse and vitess-sqlparser

var compareQueries = map[string]string{
	"simple":  "SELECT 1",
	"columns": "SELECT id, name, email, created_at FROM users",
	"where":   "SELECT * FROM users WHERE status = 'active' AND age > 18",
	"join":    "SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id",
	"complex": `SELECT u.id, u.name, COUNT(o.id) as order_count, SUM(o.total) as total_spent
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active' AND u.created_at > '2024-01-01'
		GROUP BY u.id, u.name
		HAVING COUNT(o.id) > 5
		ORDER BY total_spent DESC
		LIMIT 100`,
	"subquery":  "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
	"aggregate": "SELECT status, COUNT(*), AVG(age) FROM users GROUP BY status HAVING COUNT(*) > 10",
	"insert":    "INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com')",
	"update":    "UPDATE users SET name = 'Jane', updated_at = NOW() WHERE id = 1",
	"delete":    "DELETE FROM users WHERE status = 'deleted' AND updated_at < '2024-01-01'",
}

// BenchmarkCompareParse compares parsing performance
func BenchmarkCompareParse(b *testing.B) {
	for name, query := range compareQueries {
		b.Run("machparse/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Parse(query)
			}
		})

		b.Run("vitess/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = vitess.Parse(query)
			}
		})
	}
}

// BenchmarkCompareFormat compares formatting performance
func BenchmarkCompareFormat(b *testing.B) {
	for name, query := range compareQueries {
		// Parse with machparse
		machStmt, err := Parse(query)
		if err != nil {
			b.Skipf("machparse failed to parse %s: %v", name, err)
			continue
		}

		// Parse with vitess
		vitessStmt, err := vitess.Parse(query)
		if err != nil {
			b.Skipf("vitess failed to parse %s: %v", name, err)
			continue
		}

		b.Run("machparse/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = String(machStmt)
			}
		})

		b.Run("vitess/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = vitess.String(vitessStmt)
			}
		})
	}
}

// BenchmarkCompareRoundTrip compares full parse + format cycle
func BenchmarkCompareRoundTrip(b *testing.B) {
	for name, query := range compareQueries {
		b.Run("machparse/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				stmt, _ := Parse(query)
				_ = String(stmt)
			}
		})

		b.Run("vitess/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				stmt, _ := vitess.Parse(query)
				_ = vitess.String(stmt)
			}
		})
	}
}

// BenchmarkCompareComplexQuery focuses on the complex query used in our main benchmark
func BenchmarkCompareComplexQuery(b *testing.B) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
		FROM users u LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
		GROUP BY u.id, u.name
		ORDER BY order_count DESC`

	b.Run("machparse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Parse(query)
		}
	})

	b.Run("machparse_pooled", func(b *testing.B) {
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
	})

	b.Run("vitess", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = vitess.Parse(query)
		}
	})
}

// BenchmarkCompareAllQueries shows summary comparison across all query types
func BenchmarkCompareAllQueries(b *testing.B) {
	// Combine all queries for aggregate comparison
	allQueries := []string{
		"SELECT 1",
		"SELECT id, name, email, created_at FROM users",
		"SELECT * FROM users WHERE status = 'active' AND age > 18",
		"SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id",
		"SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
		"INSERT INTO users (id, name, email) VALUES (1, 'John', 'john@example.com')",
		"UPDATE users SET name = 'Jane', updated_at = NOW() WHERE id = 1",
		"DELETE FROM users WHERE status = 'deleted' AND updated_at < '2024-01-01'",
	}

	b.Run("machparse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, q := range allQueries {
				_, _ = Parse(q)
			}
		}
	})

	b.Run("machparse_pooled", func(b *testing.B) {
		// Warm up
		for i := 0; i < 50; i++ {
			for _, q := range allQueries {
				stmt, _ := Parse(q)
				Repool(stmt)
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, q := range allQueries {
				stmt, _ := Parse(q)
				Repool(stmt)
			}
		}
	})

	b.Run("vitess", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, q := range allQueries {
				_, _ = vitess.Parse(q)
			}
		}
	})
}
