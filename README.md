# machparse

[![Go Reference](https://pkg.go.dev/badge/github.com/freeeve/machparse.svg)](https://pkg.go.dev/github.com/freeeve/machparse)
[![Snyk Vulnerabilities](https://snyk.io/test/github/freeeve/machparse/badge.svg)](https://snyk.io/test/github/freeeve/machparse)

A high-performance SQL parser for Go. Parses MySQL, PostgreSQL, and SQLite syntax.

## Features

- **Fast**: 3-6x faster than vitess-sqlparser
- **Low memory**: Up to 300x fewer allocations with pooling
- **Multi-dialect**: Supports MySQL, PostgreSQL, and SQLite syntax
- **Complete**: SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, and more
- **Round-trip safe**: Parse → Format → Parse produces identical AST

## Installation

```bash
go get github.com/freeeve/machparse
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/freeeve/machparse"
)

func main() {
    stmt, err := machparse.Parse("SELECT id, name FROM users WHERE active = true")
    if err != nil {
        panic(err)
    }

    // Format back to SQL
    fmt.Println(machparse.String(stmt))
    // Output: SELECT id, name FROM users WHERE active = true
}
```

## Performance

Benchmarked against [vitess-sqlparser](https://github.com/blastrain/vitess-sqlparser):

| Parser | Time | Memory | Allocations |
|--------|------|--------|-------------|
| **machparse** | 6,019 ns | 3,301 B | 41 |
| **machparse + Repool** | 3,001 ns | 113 B | 6 |
| vitess-sqlparser | 17,877 ns | 34,072 B | 123 |

machparse is **3x faster** out of the box, and **6x faster** with `Repool()`.

### High-Performance Mode

For maximum performance when parsing many queries, call `Repool()` when done with a statement:

```go
stmt, err := machparse.Parse(sql)
if err != nil {
    return err
}
defer machparse.Repool(stmt)

// ... use stmt ...
```

This returns AST nodes to internal pools for reuse, reducing allocations by ~85%.

**Note**: `Repool()` is optional. If not called, nodes are garbage collected normally. Use it in high-throughput scenarios like SQL proxies or query analyzers.

## API

### Parsing

```go
// Parse a single statement
stmt, err := machparse.Parse("SELECT * FROM users")

// Parse multiple statements
stmts, err := machparse.ParseAll("SELECT 1; SELECT 2")
```

### Formatting

```go
// Format AST back to SQL
sql := machparse.String(stmt)
```

### Walking the AST

```go
machparse.Walk(stmt, func(node machparse.Node) bool {
    if col, ok := node.(*machparse.ColName); ok {
        fmt.Printf("Found column: %s\n", col.Name)
    }
    return true // continue walking
})
```

### Rewriting the AST

```go
// Rewrite is called post-order (children first, then parent)
rewritten := machparse.Rewrite(stmt, func(node machparse.Node) machparse.Node {
    // Replace table names
    if tn, ok := node.(*machparse.TableName); ok {
        tn.Name = "new_" + tn.Name
        return tn
    }
    return node
})
```

### Pooling (Optional)

```go
// Return AST nodes to pool when done (optional, for max performance)
machparse.Repool(stmt)
```

## Supported SQL

### Statements
- SELECT (with JOINs, subqueries, CTEs, window functions, UNION/INTERSECT/EXCEPT)
- INSERT (with ON CONFLICT, RETURNING)
- UPDATE
- DELETE
- CREATE TABLE/INDEX/VIEW
- ALTER TABLE
- DROP TABLE/INDEX/VIEW
- TRUNCATE
- EXPLAIN

### Expressions
- Binary operators (+, -, *, /, %, AND, OR, etc.)
- Comparison operators (=, !=, <, >, <=, >=, LIKE, IN, BETWEEN, etc.)
- Functions (COUNT, SUM, AVG, COALESCE, etc.)
- CASE expressions
- CAST/type conversion
- Subqueries
- Window functions (ROW_NUMBER, RANK, LAG, LEAD, etc.)
- Array expressions and subscripts
- JSON operators (PostgreSQL ->, ->>, etc.)

### Dialect Features
- **MySQL**: backtick quotes, AUTO_INCREMENT, ON DUPLICATE KEY
- **PostgreSQL**: double-colon casts, RETURNING, ON CONFLICT, dollar-quoted strings
- **SQLite**: AUTOINCREMENT, WITHOUT ROWID

## Examples

### Extract Table Names

```go
func extractTables(stmt machparse.Statement) []string {
    var tables []string
    machparse.Walk(stmt, func(node machparse.Node) bool {
        if tn, ok := node.(*machparse.TableName); ok {
            tables = append(tables, tn.Name)
        }
        return true
    })
    return tables
}
```

### Rewrite Column References

```go
func prefixColumns(stmt machparse.Statement, prefix string) machparse.Statement {
    return machparse.Rewrite(stmt, func(node machparse.Node) machparse.Node {
        if col, ok := node.(*machparse.ColName); ok {
            col.Name = prefix + col.Name
        }
        return node
    }).(machparse.Statement)
}
```

### High-Throughput Parsing

```go
func parseQueries(queries []string) ([]*machparse.SelectStmt, error) {
    results := make([]*machparse.SelectStmt, 0, len(queries))

    for _, sql := range queries {
        stmt, err := machparse.Parse(sql)
        if err != nil {
            return nil, err
        }

        if sel, ok := stmt.(*machparse.SelectStmt); ok {
            results = append(results, sel)
        }

        // Optional: return nodes to pool if you're done processing
        // and don't need to keep the AST around
        // machparse.Repool(stmt)
    }

    return results, nil
}
```

## License

MIT
