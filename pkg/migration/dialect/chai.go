package dialect

// ChaiDialect implements the Dialect interface from darwin for SQLite.
type ChaiDialect struct{}

// CreateTableSQL returns the query to create the schema table.
func (g ChaiDialect) CreateTableSQL() string {
	return `CREATE TABLE IF NOT EXISTS darwin_migrations (
		version        DOUBLE   PRIMARY KEY,
		description    TEXT     NOT NULL,
		checksum       TEXT     NOT NULL,
		applied_at     INTEGER  NOT NULL,
		execution_time DOUBLE   NOT NULL
	);`
}

// InsertSQL returns the query to insert a new migration in the schema table.
func (g ChaiDialect) InsertSQL() string {
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
func (g ChaiDialect) AllSQL() string {
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
