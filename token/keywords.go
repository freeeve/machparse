package token

// keywords maps lowercase keyword strings to token types.
var keywords map[string]Token

func init() {
	keywords = map[string]Token{
		// DML
		"select":   SELECT,
		"from":     FROM,
		"where":    WHERE,
		"and":      AND,
		"or":       OR,
		"xor":      XOR,
		"not":      NOT,
		"in":       IN,
		"like":     LIKE,
		"ilike":    ILIKE,
		"similar":  SIMILAR,
		"between":  BETWEEN,
		"is":       IS,
		"isnull":   ISNULL,
		"notnull":  NOTNULL,
		"null":     NULL,
		"true":     TRUE,
		"false":    FALSE,
		"unknown":  UNKNOWN,
		"as":       AS,
		"all":      ALL,
		"distinct": DISTINCT,
		"unique":   UNIQUE,

		// JOIN
		"join":    JOIN,
		"inner":   INNER,
		"left":    LEFT,
		"right":   RIGHT,
		"full":    FULL,
		"outer":   OUTER,
		"cross":   CROSS,
		"natural": NATURAL,
		"on":      ON,
		"using":   USING,

		// ORDER/GROUP
		"order":  ORDER,
		"by":     BY,
		"asc":    ASC,
		"desc":   DESC,
		"nulls":  NULLS,
		"first":  FIRST,
		"last":   LAST,
		"group":  GROUP,
		"having": HAVING,

		// LIMIT
		"limit":   LIMIT,
		"offset":  OFFSET,
		"fetch":   FETCH,
		"next":    NEXT,
		"row":     ROW,
		"rows":    ROWS,
		"only":    ONLY,
		"percent": PERCENT_KW,
		"with":    WITH,
		"ties":    TIES,

		// Set operations
		"union":     UNION,
		"intersect": INTERSECT,
		"except":    EXCEPT,

		// INSERT
		"insert":    INSERT,
		"into":      INTO,
		"values":    VALUES,
		"default":   DEFAULT,
		"returning": RETURNING,
		"replace":   REPLACE,
		"ignore":    IGNORE,
		"duplicate": DUPLICATE,
		"key":       KEY,

		// UPDATE
		"update": UPDATE,
		"set":    SET,

		// DELETE
		"delete": DELETE,

		// DDL
		"create":     CREATE,
		"alter":      ALTER,
		"drop":       DROP,
		"table":      TABLE,
		"index":      INDEX,
		"view":       VIEW,
		"database":   DATABASE,
		"schema":     SCHEMA,
		"if":         IF,
		"exists":     EXISTS,
		"temporary":  TEMPORARY,
		"temp":       TEMP,
		"unlogged":   UNLOGGED,
		"primary":    PRIMARY,
		"foreign":    FOREIGN,
		"references": REFERENCES,
		"constraint": CONSTRAINT,
		"check":      CHECK,
		"cascade":    CASCADE,
		"restrict":   RESTRICT,
		"no":         NO,
		"action":     ACTION,
		"deferrable": DEFERRABLE,
		"initially":  INITIALLY,
		"deferred":   DEFERRED,
		"immediate":  IMMEDIATE,

		// Column modifiers
		"column": COLUMN,
		"add":    ADD,
		"rename": RENAME,
		"to":     TO,
		"modify": MODIFY,
		"change": CHANGE,

		// Data types
		"int":         INT_TYPE,
		"integer":     INTEGER,
		"smallint":    SMALLINT,
		"bigint":      BIGINT,
		"tinyint":     TINYINT,
		"mediumint":   MEDIUMINT,
		"real":        REAL,
		"double":      DOUBLE,
		"precision":   PRECISION,
		"float":       FLOAT_TYPE,
		"decimal":     DECIMAL,
		"numeric":     NUMERIC,
		"char":        CHAR,
		"varchar":     VARCHAR,
		"text":        TEXT,
		"blob":        BLOB_TYPE,
		"binary":      BINARY,
		"varbinary":   VARBINARY,
		"date":        DATE,
		"time":        TIME,
		"datetime":    DATETIME,
		"timestamp":   TIMESTAMP,
		"year":        YEAR,
		"boolean":     BOOLEAN,
		"bool":        BOOL,
		"json":        JSON,
		"jsonb":       JSONB,
		"uuid":        UUID,
		"serial":      SERIAL,
		"bigserial":   BIGSERIAL,
		"smallserial": SMALLSERIAL,
		"array":       ARRAY,
		"unsigned":    UNSIGNED,
		"signed":      SIGNED,
		"zerofill":    ZEROFILL,
		"varying":     VARYING,
		"zone":        ZONE,

		// Expressions
		"case":      CASE,
		"when":      WHEN,
		"then":      THEN,
		"else":      ELSE,
		"end":       END,
		"cast":      CAST,
		"convert":   CONVERT,
		"collate":   COLLATE,
		"over":      OVER,
		"partition": PARTITION,
		"window":    WINDOW,
		"filter":    FILTER,
		"within":    WITHIN,
		"respect":   RESPECT,
		"current":   CURRENT,
		"unbounded": UNBOUNDED,
		"preceding": PRECEDING,
		"following": FOLLOWING,
		"range":     RANGE,
		"groups":    GROUPS,

		// Aggregates
		"count":    COUNT,
		"sum":      SUM,
		"avg":      AVG,
		"min":      MIN,
		"max":      MAX,
		"coalesce": COALESCE,
		"nullif":   NULLIF,
		"greatest": GREATEST,
		"least":    LEAST,
		"any":      ANY,
		"some":     SOME,
		"every":    EVERY,

		// Subqueries
		"lateral":      LATERAL,
		"recursive":    RECURSIVE,
		"materialized": MATERIALIZED,

		// Locking
		"for":    FOR,
		"share":  SHARE,
		"nowait": NOWAIT,
		"skip":   SKIP,
		"locked": LOCKED,

		// Transaction
		"begin":        BEGIN,
		"commit":       COMMIT,
		"rollback":     ROLLBACK,
		"savepoint":    SAVEPOINT,
		"release":      RELEASE,
		"transaction":  TRANSACTION,
		"work":         WORK,
		"isolation":    ISOLATION,
		"level":        LEVEL,
		"read":         READ,
		"write":        WRITE,
		"committed":    COMMITTED,
		"uncommitted":  UNCOMMITTED,
		"repeatable":   REPEATABLE,
		"serializable": SERIALIZABLE,
		"snapshot":     SNAPSHOT,

		// CTE
		"ordinality": ORDINALITY,

		// Misc
		"analyze":    ANALYZE,
		"explain":    EXPLAIN,
		"verbose":    VERBOSE,
		"format":     FORMAT,
		"costs":      COSTS,
		"buffers":    BUFFERS,
		"timing":     TIMING,
		"truncate":   TRUNCATE,
		"vacuum":     VACUUM,
		"grant":      GRANT,
		"revoke":     REVOKE,
		"privileges": PRIVILEGES,
		"public":     PUBLIC,
		"role":       ROLE,
		"user":       USER,
		"admin":      ADMIN,
		"option":     OPTION,
		"granted":    GRANTED,

		// Date/time
		"interval":        INTERVAL,
		"extract":         EXTRACT,
		"epoch":           EPOCH,
		"century":         CENTURY,
		"decade":          DECADE,
		"millennium":      MILLENNIUM,
		"quarter":         QUARTER,
		"month":           MONTH,
		"week":            WEEK,
		"day":             DAY,
		"hour":            HOUR,
		"minute":          MINUTE,
		"second":          SECOND,
		"microsecond":     MICROSECOND,
		"timezone":        TIMEZONE,
		"timezone_hour":   TIMEZONE_HOUR,
		"timezone_minute": TIMEZONE_MINUTE,

		// String functions
		"substring": SUBSTRING,
		"trim":      TRIM,
		"leading":   LEADING,
		"trailing":  TRAILING,
		"both":      BOTH,
		"position":  POSITION,
		"overlay":   OVERLAY,
		"placing":   PLACING,

		// Boolean
		"symmetric":  SYMMETRIC,
		"asymmetric": ASYMMETRIC,
		"escape":     ESCAPE,

		// Pattern matching
		"glob":    GLOB,
		"regexp":  REGEXP,
		"rlike":   RLIKE,
		"match":   MATCH,
		"against": AGAINST,
		"sounds":  SOUNDS,

		// SQLite
		"autoincrement": AUTOINCREMENT,
		"rowid":         ROWID,
		"without":       WITHOUT,

		// MySQL
		"auto_increment":      AUTO_INCREMENT,
		"engine":              ENGINE,
		"charset":             CHARSET,
		"character":           CHARACTER,
		"comment":             COMMENT_KW,
		"storage":             STORAGE,
		"memory":              MEMORY,
		"disk":                DISK,
		"tablespace":          TABLESPACE,
		"data":                DATA,
		"directory":           DIRECTORY,
		"connection":          CONNECTION,
		"partitions":          PARTITIONS,
		"subpartition":        SUBPARTITION,
		"subpartitions":       SUBPARTITIONS,
		"hash":                HASH,
		"linear":              LINEAR,
		"list":                LIST,
		"less":                LESS,
		"than":                THAN,
		"maxvalue":            MAXVALUE,
		"algorithm":           ALGORITHM,
		"inplace":             INPLACE,
		"copy":                COPY,
		"lock":                LOCK_KW,
		"none":                NONE,
		"shared":              SHARED,
		"exclusive":           EXCLUSIVE,
		"force":               FORCE,
		"use":                 USE,
		"straight_join":       STRAIGHT_JOIN,
		"sql_calc_found_rows": SQL_CALC_FOUND_ROWS,
		"sql_small_result":    SQL_SMALL_RESULT,
		"sql_big_result":      SQL_BIG_RESULT,
		"sql_buffer_result":   SQL_BUFFER_RESULT,
		"high_priority":       HIGH_PRIORITY,
		"low_priority":        LOW_PRIORITY,
		"delayed":             DELAYED,
		"quick":               QUICK,
		"concurrent":          CONCURRENT,
		"local":               LOCAL,
		"infile":              INFILE,
		"load":                LOAD,
		"outfile":             OUTFILE,
		"terminated":          TERMINATED,
		"enclosed":            ENCLOSED,
		"escaped":             ESCAPED,
		"lines":               LINES,
		"starting":            STARTING,
		"optionally":          OPTIONALLY,
		"fields":              FIELDS,

		// PostgreSQL
		"conflict":     CONFLICT,
		"do":           DO,
		"nothing":      NOTHING,
		"overriding":   OVERRIDING,
		"system":       SYSTEM,
		"value":        VALUE,
		"generated":    GENERATED,
		"always":       ALWAYS,
		"identity":     IDENTITY,
		"stored":       STORED,
		"virtual":      VIRTUAL,
		"include":      INCLUDE,
		"btree":        BTREE,
		"gin":          GIN,
		"gist":         GIST,
		"spgist":       SPGIST,
		"brin":         BRIN,
		"concurrently": CONCURRENTLY,
		"inherit":      INHERIT,
		"inherits":     INHERITS,
		"of":           OF,
		"oids":         OIDS,
		"owner":        OWNER,
		"owned":        OWNED,
		"depends":      DEPENDS,
		"extension":    EXTENSION,
		"sequence":     SEQUENCE,
		"cycle":        CYCLE,
		"increment":    INCREMENT,
		"minvalue":     MINVALUE,
		"start":        START,
		"cache":        CACHE,
		"restart":      RESTART,
		"continue":     CONTINUE,
		"preserve":     PRESERVE,
		"dispose":      DISPOSE,

		// SQL Server specific
		"top":             TOP,
		"nolock":          NOLOCK,
		"readuncommitted": READUNCOMMITTED,
		"readcommitted":   READCOMMITTED,
		"repeatableread":  REPEATABLEREAD,
		"rowlock":         ROWLOCK,
		"paglock":         PAGLOCK,
		"tablock":         TABLOCK,
		"tablockx":        TABLOCKX,
		"updlock":         UPDLOCK,
		"xlock":           XLOCK,
		"holdlock":        HOLDLOCK,
		"pivot":           PIVOT,
		"unpivot":         UNPIVOT,
		"apply":           APPLY,
		"merge":           MERGE,
		"inserted":        INSERTED,

		// Oracle specific
		"rownum":       ROWNUM,
		"sysdate":      SYSDATE,
		"systimestamp": SYSTIMESTAMP,
		"dual":         DUAL,
		"prior":        PRIOR,
		"nocycle":      NOCYCLE,
		"siblings":     SIBLINGS,
		"sample":       SAMPLE,
		"seed":         SEED,
		"flashback":    FLASHBACK,
		"scn":          SCN,
		"versions":     VERSIONS,
		"keep":         KEEP,
		"dense_rank":   DENSE_RANK,
		"model":        MODEL,
		"dimension":    DIMENSION,
		"measures":     MEASURES,
		"rules":        RULES,
		"iterate":      ITERATE,
		"until":        UNTIL,
		"bulk":         BULK,
		"forall":       FORALL,
		"collect":      COLLECT,
		"pipelined":    PIPELINED,
	}
}

// LookupIdent returns the token type for an identifier.
// If the identifier is a keyword, returns the keyword token.
// Otherwise returns IDENT.
// This implementation avoids allocation by checking if the string
// is already lowercase before doing a conversion.
func LookupIdent(ident string) Token {
	// Fast path: check if already lowercase (common case)
	if isLowercase(ident) {
		if tok, ok := keywords[ident]; ok {
			return tok
		}
		return IDENT
	}

	// Slow path: need to lowercase
	// Use stack-allocated buffer for short strings (covers all keywords)
	if len(ident) <= 32 {
		var buf [32]byte
		for i := 0; i < len(ident); i++ {
			c := ident[i]
			if c >= 'A' && c <= 'Z' {
				buf[i] = c + 32
			} else {
				buf[i] = c
			}
		}
		// Convert to string for map lookup - this still allocates
		// but only for mixed-case identifiers (rare for SQL)
		lower := string(buf[:len(ident)])
		if tok, ok := keywords[lower]; ok {
			return tok
		}
		return IDENT
	}

	// Very long identifiers - can't be keywords anyway (max keyword is ~20 chars)
	return IDENT
}

// isLowercase checks if a string contains only lowercase ASCII letters,
// digits, and underscores (valid SQL identifier chars).
func isLowercase(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

// IsKeyword returns true if the identifier is a SQL keyword.
func IsKeyword(ident string) bool {
	return LookupIdent(ident) != IDENT
}
