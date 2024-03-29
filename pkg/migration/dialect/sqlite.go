package dialect

// SqliteDialect implements the Dialect interface from darwin for SQLite.
type SqliteDialect struct{}

// CreateTableSQL returns the query to create the schema table.
func (s SqliteDialect) CreateTableSQL() string {
	return `CREATE TABLE IF NOT EXISTS darwin_migrations (
		id             INTEGER  PRIMARY KEY,
		version        FLOAT    NOT NULL,
		description    TEXT     NOT NULL,
		checksum       TEXT     NOT NULL,
		applied_at     DATETIME NOT NULL,
		execution_time FLOAT    NOT NULL,
		UNIQUE         (version)
	);`
}

// InsertSQL returns the query to insert a new migration in the schema table.
func (s SqliteDialect) InsertSQL() string {
	return `INSERT INTO darwin_migrations (
		version,
		description,
		checksum,
		applied_at,
		execution_time
	) VALUES (
		?, ?, ?, ?, ?
	);`
}

// AllSQL returns a query to get all entries in the table.
func (s SqliteDialect) AllSQL() string {
	return `SELECT
		version,
		description,
		checksum,
		applied_at,
		execution_time
	FROM
		darwin_migrations
	ORDER BY
		version ASC;`
}
