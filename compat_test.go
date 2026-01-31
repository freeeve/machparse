package machparse

import (
	"testing"
)

// TestVitessCompatibility tests SQL queries from vitess-sqlparser to ensure compatibility.
func TestVitessCompatibility(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Basic SELECT variations
		{"simple select", "select 1 from t"},
		{"select list", "select 1, 2 from t"},
		{"select star", "select * from t"},
		{"select qualified star", "select a.* from t"},
		{"select qualified star 2 levels", "select a.b.* from t"},
		{"select distinct", "select distinct 1 from t"},
		{"column alias", "select a as b from t"},
		{"column alias without as", "select a b from t"},

		// WHERE clause variations
		{"where equals", "select * from t where a = 1"},
		{"where and", "select * from t where a = 1 and b = 2"},
		{"where or", "select * from t where a = 1 or b = 2"},
		{"where in", "select * from t where a in (1, 2, 3)"},
		{"where not in", "select * from t where a not in (1, 2, 3)"},
		{"where between", "select * from t where a between 1 and 10"},
		{"where like", "select * from t where a like '%test%'"},
		{"where is null", "select * from t where a is null"},
		{"where is not null", "select * from t where a is not null"},

		// JOIN variations
		{"join", "select * from t1 join t2 on t1.id = t2.id"},
		{"left join", "select * from t1 left join t2 on t1.id = t2.id"},
		{"right join", "select * from t1 right join t2 on t1.id = t2.id"},
		{"cross join", "select * from t1 cross join t2"},
		{"multiple joins", "select * from t1 join t2 on a = b join t3 on c = d"},
		{"join using", "select * from t1 join t2 using (id)"},
		{"natural join", "select * from t1 natural join t2"},
		{"table list", "select 1 from t1, t2"},
		{"join with subquery", "select * from t1 join (select * from t2 union select * from t3) as t on t1.id = t.id"},

		// UNION variations
		{"union", "select 1 from t union select 2 from t"},
		{"union all", "select 1 from t union all select 2 from t"},
		{"double union", "select 1 from t union select 2 from t union select 3 from t"},
		{"union with order by", "select 1 from t union select 2 from t order by 1"},
		{"union with limit", "select 1 from t union select 2 from t limit 10"},
		{"parenthesized union", "(select 1 from t) union (select 2 from t)"},
		{"complex union", "select 1 from t union select 2 from t union all select 3 from t"},

		// Subqueries
		{"subquery in from", "select * from (select 1 from t) as sub"},
		{"subquery in where", "select * from t where id in (select id from t2)"},
		{"correlated subquery", "select * from t where exists (select 1 from t2 where t2.id = t.id)"},
		{"subquery with union", "select * from t where col in (select 1 from t2 union select 2 from t3)"},
		{"subquery exists", "select * from t1 where exists (select a from t2 union select b from t3)"},

		// CTE (WITH clause)
		{"simple cte", "with cte as (select 1 from t) select * from cte"},
		{"cte with columns", "with cte (a, b) as (select 1, 2 from t) select * from cte"},
		{"multiple ctes", "with cte1 as (select 1 from t), cte2 as (select 2 from t) select * from cte1, cte2"},
		{"recursive cte", "with recursive cte (id, n) as (select 1, 1 from t union all select id+1, n+2 from cte where id < 5) select * from cte"},
		{"complex cte", "with topsales as (select rep_id, sum(amount) as sales from orders group by rep_id order by sales desc limit 5) select * from employees join topsales using (rep_id)"},

		// GROUP BY / HAVING
		{"group by", "select a, count(*) from t group by a"},
		{"group by multiple", "select a, b, count(*) from t group by a, b"},
		{"having", "select a, count(*) from t group by a having count(*) > 5"},

		// ORDER BY / LIMIT
		{"order by", "select * from t order by a"},
		{"order by desc", "select * from t order by a desc"},
		{"order by multiple", "select * from t order by a, b desc"},
		{"limit", "select * from t limit 10"},
		{"limit offset", "select * from t limit 10 offset 20"},
		{"limit comma syntax", "select * from t limit 20, 10"},

		// CASE expressions
		{"case when", "select case when a = 1 then 'one' end from t"},
		{"case when else", "select case when a = 1 then 'one' else 'other' end from t"},
		{"case when multiple", "select case when a = 1 then 'one' when a = 2 then 'two' else 'other' end from t"},
		{"case value", "select case a when 1 then 'one' when 2 then 'two' end from t"},

		// Functions
		{"count star", "select count(*) from t"},
		{"count column", "select count(a) from t"},
		{"count distinct", "select count(distinct a) from t"},
		{"sum", "select sum(a) from t"},
		{"avg", "select avg(a) from t"},
		{"min max", "select min(a), max(a) from t"},
		{"coalesce", "select coalesce(a, b, c) from t"},
		{"nullif", "select nullif(a, b) from t"},
		{"concat", "select concat(a, b) from t"},

		// EXTRACT
		{"extract year", "select extract(year from created_at) from t"},
		{"extract month", "select extract(month from created_at) from t"},

		// CAST
		{"cast", "select cast(a as int) from t"},
		{"cast varchar", "select cast(a as varchar(255)) from t"},
		{"pg cast", "select a::int from t"},

		// String operations
		{"like escape", "select * from t where a like '%test%' escape '#'"},
		{"concat operator", "select a || b from t"},

		// Arithmetic
		{"add", "select a + b from t"},
		{"subtract", "select a - b from t"},
		{"multiply", "select a * b from t"},
		{"divide", "select a / b from t"},
		{"modulo", "select a % b from t"},
		{"unary minus", "select -a from t"},
		{"complex arithmetic", "select (a + b) * c / d from t"},

		// Comparison operators
		{"not equals", "select * from t where a != b"},
		{"not equals 2", "select * from t where a <> b"},
		{"less than", "select * from t where a < b"},
		{"greater than", "select * from t where a > b"},
		{"less than or equal", "select * from t where a <= b"},
		{"greater than or equal", "select * from t where a >= b"},

		// Parentheses
		{"parenthesized expr", "select (a + b) from t"},
		{"nested parentheses", "select ((a + b) * c) from t"},
		{"parenthesis in table", "select 1 from (t)"},
		{"parenthesis multi-table", "select 1 from (t1, t2)"},

		// INSERT variations
		{"insert values", "insert into t (a, b) values (1, 2)"},
		{"insert multiple rows", "insert into t (a, b) values (1, 2), (3, 4)"},
		{"insert select", "insert into t select * from t2"},
		{"replace", "replace into t (a, b) values (1, 2)"},
		{"insert on duplicate", "insert into t (a) values (1) on duplicate key update a = 2"},
		{"insert returning", "insert into t (a) values (1) returning id"},
		{"insert on conflict do nothing", "insert into t (a) values (1) on conflict (a) do nothing"},
		{"insert on conflict do update", "insert into t (a) values (1) on conflict (a) do update set b = 2"},

		// UPDATE variations
		{"update", "update t set a = 1"},
		{"update where", "update t set a = 1 where b = 2"},
		{"update multiple", "update t set a = 1, b = 2 where c = 3"},
		{"update with from", "update t set a = t2.a from t2 where t.id = t2.id"},
		{"update returning", "update t set a = 1 returning *"},

		// DELETE variations
		{"delete", "delete from t"},
		{"delete where", "delete from t where a = 1"},
		{"delete using", "delete from t using t2 where t.id = t2.id"},
		{"delete returning", "delete from t where a = 1 returning *"},

		// CREATE TABLE variations
		{"create table simple", "create table t (id int)"},
		{"create table multiple columns", "create table t (id int, name varchar(255))"},
		{"create table primary key", "create table t (id int primary key)"},
		{"create table not null", "create table t (id int not null)"},
		{"create table default", "create table t (id int default 0)"},
		{"create table if not exists", "create table if not exists t (id int)"},
		{"create table foreign key", "create table t (id int, foreign key (id) references t2(id))"},
		{"create table unique", "create table t (id int unique)"},

		// Index operations
		{"create index", "create index idx on t (a)"},
		{"create unique index", "create unique index idx on t (a)"},
		{"create index multiple columns", "create index idx on t (a, b)"},
		{"drop index", "drop index idx on t"},

		// ALTER TABLE
		{"alter table add column", "alter table t add column a int"},
		{"alter table drop column", "alter table t drop column a"},
		{"alter table modify column", "alter table t modify column a varchar(255)"},

		// DROP TABLE
		{"drop table", "drop table t"},
		{"drop table if exists", "drop table if exists t"},
		{"drop table cascade", "drop table t cascade"},

		// Window functions
		{"row_number", "select row_number() over () from t"},
		{"row_number order by", "select row_number() over (order by id) from t"},
		{"row_number partition by", "select row_number() over (partition by type order by id) from t"},
		{"sum over", "select sum(a) over (partition by b) from t"},
		{"avg over window", "select avg(a) over (order by b rows between 1 preceding and 1 following) from t"},

		// Locking
		{"for update", "select * from t for update"},
		{"for share", "select * from t for share"},

		// Comments (should be ignored)
		{"line comment", "select 1 from t -- comment"},
		{"block comment", "select /* comment */ 1 from t"},

		// PostgreSQL specific
		{"pg array", "select array[1, 2, 3]"},
		{"pg any", "select * from t where a = any(array[1,2,3])"},

		// Boolean literals
		{"true", "select * from t where a = true"},
		{"false", "select * from t where a = false"},

		// NULL handling
		{"null", "select null from t"},
		{"is null", "select * from t where a is null"},
		{"is not null", "select * from t where a is not null"},
		{"coalesce null", "select coalesce(a, null, b) from t"},

		// Qualified identifiers (multi-level)
		{"qualified column", "select t.a from t"},
		{"schema qualified table", "select * from schema1.t"},
		{"schema qualified column", "select schema1.t.a from schema1.t"},
		{"catalog schema table", "select * from catalog1.schema1.t"},
		{"full qualified column", "select catalog1.schema1.t.a from catalog1.schema1.t"},

		// SQL Server specific
		{"bracket identifier", "select [column name] from [table name]"},
		{"bracket with spaces", "select [my column] from [my table]"},
		{"temp table", "select * from #temp"},
		{"global temp table", "select * from ##global_temp"},
		{"bracket and temp", "select [col] from #temp_table"},

		// SQL Server TOP clause (parses as function, works with identifiers)
		{"top clause", "select top(10) * from t"},
		// Note: WITH (NOLOCK) table hints not yet supported
		// {"nolock hint", "select * from t with (nolock)"},

		// Oracle specific
		{"rownum", "select * from t where rownum <= 10"},
		{"sysdate", "select sysdate from dual"},
		{"dual table", "select 1 from dual"},
		// Note: CONNECT BY / START WITH hierarchical queries not yet supported
		// {"connect by", "select * from t connect by prior id = parent_id"},
		// {"start with", "select * from t start with parent_id is null connect by prior id = parent_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v\nInput: %s", err, tt.input)
			}
			if stmt == nil {
				t.Fatalf("Parse returned nil statement\nInput: %s", tt.input)
			}

			// Round-trip test
			formatted := String(stmt)
			if formatted == "" {
				t.Fatalf("Format returned empty string\nInput: %s", tt.input)
			}

			stmt2, err := Parse(formatted)
			if err != nil {
				t.Fatalf("Re-parse error: %v\nOriginal: %s\nFormatted: %s", err, tt.input, formatted)
			}

			formatted2 := String(stmt2)
			if formatted != formatted2 {
				t.Errorf("Round-trip mismatch:\nOriginal:  %s\nFirst:     %s\nSecond:    %s", tt.input, formatted, formatted2)
			}
		})
	}
}
