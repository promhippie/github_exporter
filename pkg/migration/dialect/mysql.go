package dialect

// MySQLDialect implements the Dialect interface from darwin for MySQL.
type MySQLDialect struct{}

// CreateTableSQL returns the query to create the schema table.
func (m MySQLDialect) CreateTableSQL() string {
	return `CREATE TABLE IF NOT EXISTS darwin_migrations (
		id             INT          AUTO_INCREMENT,
		version        FLOAT        NOT NULL,
		description    VARCHAR(255) NOT NULL,
		checksum       VARCHAR(32)  NOT NULL,
		applied_at     INT          NOT NULL,
		execution_time FLOAT        NOT NULL,
		UNIQUE         (version),
		PRIMARY KEY    (id)
	) ENGINE=InnoDB CHARACTER SET=utf8;`
}

// InsertSQL returns the query to insert a new migration in the schema table.
func (m MySQLDialect) InsertSQL() string {
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
func (m MySQLDialect) AllSQL() string {
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
