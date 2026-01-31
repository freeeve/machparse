// Package token defines SQL token types and position tracking.
package token

// Token represents a SQL token type.
type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	literalBeg
	IDENT  // table_name, column_name
	INT    // 12345
	FLOAT  // 123.45
	STRING // 'string literal'
	BLOB   // X'...' or 0x...
	PARAM  // ? or $1 or :name
	literalEnd

	operatorBeg
	PLUS        // +
	MINUS       // -
	ASTERISK    // *
	SLASH       // /
	PERCENT     // %
	EQ          // =
	NEQ         // != or <>
	LT          // <
	GT          // >
	LTE         // <=
	GTE         // >=
	LPAREN      // (
	RPAREN      // )
	LBRACKET    // [
	RBRACKET    // ]
	COMMA       // ,
	SEMICOLON   // ;
	DOT         // .
	COLON       // :
	DCOLON      // :: (PostgreSQL cast)
	CONCAT      // ||
	BITAND      // &
	BITOR       // |
	BITXOR      // ^
	BITNOT      // ~
	LSHIFT      // <<
	RSHIFT      // >>
	ARROW       // -> (JSON)
	DARROW      // ->> (JSON)
	HASHGT      // #> (PostgreSQL JSON)
	HASHDGT     // #>> (PostgreSQL JSON)
	QUESTION    // ? (PostgreSQL JSON/HSTORE)
	QUESTIONOR  // ?| (PostgreSQL HSTORE)
	QUESTIONAND // ?& (PostgreSQL HSTORE)
	AT          // @
	ATAT        // @@ (PostgreSQL text search)
	operatorEnd

	keywordBeg
	// DML keywords
	SELECT
	FROM
	WHERE
	AND
	OR
	XOR
	NOT
	IN
	LIKE
	ILIKE
	SIMILAR
	BETWEEN
	IS
	ISNULL
	NOTNULL
	NULL
	TRUE
	FALSE
	UNKNOWN
	AS
	ALL
	DISTINCT
	UNIQUE

	// JOIN keywords
	JOIN
	INNER
	LEFT
	RIGHT
	FULL
	OUTER
	CROSS
	NATURAL
	ON
	USING

	// ORDER/GROUP keywords
	ORDER
	BY
	ASC
	DESC
	NULLS
	FIRST
	LAST
	GROUP
	HAVING

	// LIMIT keywords
	LIMIT
	OFFSET
	FETCH
	NEXT
	ROW
	ROWS
	ONLY
	PERCENT_KW
	WITH
	TIES

	// Set operations
	UNION
	INTERSECT
	EXCEPT

	// INSERT keywords
	INSERT
	INTO
	VALUES
	DEFAULT
	RETURNING
	REPLACE
	IGNORE
	DUPLICATE
	KEY
	UPDATE

	// UPDATE keywords
	SET

	// DELETE keywords
	DELETE

	// DDL keywords
	CREATE
	ALTER
	DROP
	TABLE
	INDEX
	VIEW
	DATABASE
	SCHEMA
	IF
	EXISTS
	TEMPORARY
	TEMP
	UNLOGGED
	PRIMARY
	FOREIGN
	REFERENCES
	CONSTRAINT
	CHECK
	CASCADE
	RESTRICT
	NO
	ACTION
	DEFERRABLE
	INITIALLY
	DEFERRED
	IMMEDIATE

	// Column modifiers
	COLUMN
	ADD
	RENAME
	TO
	MODIFY
	CHANGE

	// Data types
	INT_TYPE
	INTEGER
	SMALLINT
	BIGINT
	TINYINT
	MEDIUMINT
	REAL
	DOUBLE
	PRECISION
	FLOAT_TYPE
	DECIMAL
	NUMERIC
	CHAR
	VARCHAR
	TEXT
	BLOB_TYPE
	BINARY
	VARBINARY
	DATE
	TIME
	DATETIME
	TIMESTAMP
	YEAR
	BOOLEAN
	BOOL
	JSON
	JSONB
	UUID
	SERIAL
	BIGSERIAL
	SMALLSERIAL
	ARRAY
	UNSIGNED
	SIGNED
	ZEROFILL
	VARYING
	ZONE

	// Functions and expressions
	CASE
	WHEN
	THEN
	ELSE
	END
	CAST
	CONVERT
	COLLATE
	OVER
	PARTITION
	WINDOW
	FILTER
	WITHIN
	RESPECT
	NULLS_KW
	CURRENT
	UNBOUNDED
	PRECEDING
	FOLLOWING
	RANGE
	GROUPS

	// Aggregate functions (as keywords for special handling)
	COUNT
	SUM
	AVG
	MIN
	MAX
	COALESCE
	NULLIF
	GREATEST
	LEAST
	ANY
	SOME
	EVERY

	// Subquery keywords
	LATERAL
	RECURSIVE
	MATERIALIZED

	// Locking
	FOR
	SHARE
	NOWAIT
	SKIP
	LOCKED

	// Transaction
	BEGIN
	COMMIT
	ROLLBACK
	SAVEPOINT
	RELEASE
	TRANSACTION
	WORK
	ISOLATION
	LEVEL
	READ
	WRITE
	COMMITTED
	UNCOMMITTED
	REPEATABLE
	SERIALIZABLE
	SNAPSHOT

	// CTE
	ORDINALITY

	// Misc
	ANALYZE
	EXPLAIN
	VERBOSE
	FORMAT
	COSTS
	BUFFERS
	TIMING
	TRUNCATE
	VACUUM
	GRANT
	REVOKE
	PRIVILEGES
	PUBLIC
	ROLE
	USER
	ADMIN
	OPTION
	GRANTED

	// Special values
	INTERVAL
	EXTRACT
	EPOCH
	CENTURY
	DECADE
	MILLENNIUM
	QUARTER
	MONTH
	WEEK
	DAY
	HOUR
	MINUTE
	SECOND
	MICROSECOND
	TIMEZONE
	TIMEZONE_HOUR
	TIMEZONE_MINUTE

	// String functions
	SUBSTRING
	TRIM
	LEADING
	TRAILING
	BOTH
	POSITION
	OVERLAY
	PLACING

	// Boolean tests
	SYMMETRIC
	ASYMMETRIC
	ESCAPE

	// LIKE variants
	GLOB
	REGEXP
	RLIKE
	MATCH
	AGAINST
	SOUNDS

	// SQLite specific
	AUTOINCREMENT
	ROWID
	WITHOUT

	// MySQL specific
	AUTO_INCREMENT
	ENGINE
	CHARSET
	CHARACTER
	COMMENT_KW
	STORAGE
	MEMORY
	DISK
	TABLESPACE
	DATA
	DIRECTORY
	CONNECTION
	PARTITION_KW
	PARTITIONS
	SUBPARTITION
	SUBPARTITIONS
	HASH
	LINEAR
	LIST
	LESS
	THAN
	MAXVALUE
	ALGORITHM
	INPLACE
	COPY
	LOCK_KW
	NONE
	SHARED
	EXCLUSIVE
	FORCE
	USE
	STRAIGHT_JOIN
	SQL_CALC_FOUND_ROWS
	SQL_SMALL_RESULT
	SQL_BIG_RESULT
	SQL_BUFFER_RESULT
	HIGH_PRIORITY
	LOW_PRIORITY
	DELAYED
	QUICK
	CONCURRENT
	LOCAL
	INFILE
	LOAD
	OUTFILE
	TERMINATED
	ENCLOSED
	ESCAPED
	LINES
	STARTING
	OPTIONALLY
	FIELDS

	// PostgreSQL specific
	ILIKE_KW
	SIMILAR_KW
	RETURNING_KW
	CONFLICT
	DO
	NOTHING
	OVERRIDING
	SYSTEM
	VALUE
	GENERATED
	ALWAYS
	IDENTITY
	STORED
	VIRTUAL
	INCLUDE
	USING_KW
	BTREE
	GIN
	GIST
	SPGIST
	BRIN
	CONCURRENTLY
	ONLY_KW
	INHERIT
	INHERITS
	OF
	OIDS
	TABLESPACE_KW
	OWNER
	OWNED
	OWNED_KW
	DEPENDS
	EXTENSION
	SEQUENCE
	CYCLE
	INCREMENT
	MINVALUE
	START
	CACHE
	RESTART
	CONTINUE
	OWNED_BY
	TEMP_KW
	PRESERVE
	DISPOSE

	// SQL Server specific
	TOP
	NOLOCK
	READUNCOMMITTED
	READCOMMITTED
	REPEATABLEREAD
	ROWLOCK
	PAGLOCK
	TABLOCK
	TABLOCKX
	UPDLOCK
	XLOCK
	HOLDLOCK
	NOWAIT_KW
	PIVOT
	UNPIVOT
	APPLY
	OUTER_APPLY
	CROSS_APPLY
	MERGE
	OUTPUT_KW
	INSERTED
	DELETED_KW

	// Oracle specific
	ROWNUM
	ROWID_KW
	SYSDATE
	SYSTIMESTAMP
	DUAL
	CONNECT_KW
	START_WITH
	PRIOR
	LEVEL_KW
	NOCYCLE
	SIBLINGS
	MINUS_KW
	SAMPLE
	SEED
	FLASHBACK
	SCN
	VERSIONS
	KEEP
	DENSE_RANK
	FIRST_KW
	LAST_KW
	MODEL
	DIMENSION
	MEASURES
	RULES
	ITERATE
	UNTIL
	RETURN_KW
	RETURNING_BULK
	BULK
	FORALL
	COLLECT
	PIPELINED

	keywordEnd
)

// Pos represents a position in the source.
type Pos struct {
	Offset int // byte offset from start
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

// IsValid returns true if the position is valid.
func (p Pos) IsValid() bool {
	return p.Line > 0
}

// Item represents a lexed token with position and value.
type Item struct {
	Type  Token
	Value string
	Pos   Pos
}

// String returns the token type as a string.
func (t Token) String() string {
	if int(t) < len(tokenNames) {
		return tokenNames[t]
	}
	return "UNKNOWN"
}

// IsLiteral returns true if the token is a literal value.
func (t Token) IsLiteral() bool {
	return t > literalBeg && t < literalEnd
}

// IsOperator returns true if the token is an operator.
func (t Token) IsOperator() bool {
	return t > operatorBeg && t < operatorEnd
}

// IsKeyword returns true if the token is a keyword.
func (t Token) IsKeyword() bool {
	return t > keywordBeg && t < keywordEnd
}

var tokenNames = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	COMMENT:    "COMMENT",
	IDENT:      "IDENT",
	INT:        "INT",
	FLOAT:      "FLOAT",
	STRING:     "STRING",
	BLOB:       "BLOB",
	PARAM:      "PARAM",
	PLUS:       "+",
	MINUS:      "-",
	ASTERISK:   "*",
	SLASH:      "/",
	PERCENT:    "%",
	EQ:         "=",
	NEQ:        "!=",
	LT:         "<",
	GT:         ">",
	LTE:        "<=",
	GTE:        ">=",
	LPAREN:     "(",
	RPAREN:     ")",
	LBRACKET:   "[",
	RBRACKET:   "]",
	COMMA:      ",",
	SEMICOLON:  ";",
	DOT:        ".",
	COLON:      ":",
	DCOLON:     "::",
	CONCAT:     "||",
	BITAND:     "&",
	BITOR:      "|",
	BITXOR:     "^",
	BITNOT:     "~",
	LSHIFT:     "<<",
	RSHIFT:     ">>",
	ARROW:      "->",
	DARROW:     "->>",
	SELECT:     "SELECT",
	FROM:       "FROM",
	WHERE:      "WHERE",
	AND:        "AND",
	OR:         "OR",
	NOT:        "NOT",
	IN:         "IN",
	LIKE:       "LIKE",
	BETWEEN:    "BETWEEN",
	IS:         "IS",
	NULL:       "NULL",
	TRUE:       "TRUE",
	FALSE:      "FALSE",
	AS:         "AS",
	ALL:        "ALL",
	DISTINCT:   "DISTINCT",
	JOIN:       "JOIN",
	INNER:      "INNER",
	LEFT:       "LEFT",
	RIGHT:      "RIGHT",
	FULL:       "FULL",
	OUTER:      "OUTER",
	CROSS:      "CROSS",
	NATURAL:    "NATURAL",
	ON:         "ON",
	USING:      "USING",
	ORDER:      "ORDER",
	BY:         "BY",
	ASC:        "ASC",
	DESC:       "DESC",
	NULLS:      "NULLS",
	FIRST:      "FIRST",
	LAST:       "LAST",
	GROUP:      "GROUP",
	HAVING:     "HAVING",
	LIMIT:      "LIMIT",
	OFFSET:     "OFFSET",
	UNION:      "UNION",
	INTERSECT:  "INTERSECT",
	EXCEPT:     "EXCEPT",
	INSERT:     "INSERT",
	INTO:       "INTO",
	VALUES:     "VALUES",
	DEFAULT:    "DEFAULT",
	RETURNING:  "RETURNING",
	UPDATE:     "UPDATE",
	SET:        "SET",
	DELETE:     "DELETE",
	CREATE:     "CREATE",
	ALTER:      "ALTER",
	DROP:       "DROP",
	TABLE:      "TABLE",
	INDEX:      "INDEX",
	IF:         "IF",
	EXISTS:     "EXISTS",
	PRIMARY:    "PRIMARY",
	KEY:        "KEY",
	FOREIGN:    "FOREIGN",
	REFERENCES: "REFERENCES",
	UNIQUE:     "UNIQUE",
	CONSTRAINT: "CONSTRAINT",
	CHECK:      "CHECK",
	CASCADE:    "CASCADE",
	RESTRICT:   "RESTRICT",
	CASE:       "CASE",
	WHEN:       "WHEN",
	THEN:       "THEN",
	ELSE:       "ELSE",
	END:        "END",
	CAST:       "CAST",
	OVER:       "OVER",
	PARTITION:  "PARTITION",
	WINDOW:     "WINDOW",
	FILTER:     "FILTER",
	FOR:        "FOR",
	WITH:       "WITH",
}
