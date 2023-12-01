//go:build sqlite

package store

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/GuiaBolso/darwin"
	"github.com/go-kit/log"
	"github.com/google/go-github/v57/github"
	"github.com/jmoiron/sqlx"
	"github.com/promhippie/github_exporter/pkg/migration/dialect"

	// Import SQLite driver for database/sql
	_ "modernc.org/sqlite"
)

var (
	sqliteMigrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table workflow_runs",
			Script: `CREATE TABLE workflow_runs (
				owner TEXT NOT NULL,
				repo TEXT NOT NULL,
				workflow_id INTEGER NOT NULL,
				number INTEGER NOT NULL,
				attempt INTEGER,
				event TEXT,
				name TEXT,
				title TEXT,
				status TEXT,
				branch TEXT,
				sha TEXT,
				identifier BIGINT,
				created_at BIGINT,
				updated_at BIGINT,
				started_at BIGINT,
				PRIMARY KEY(owner, repo, workflow_id, number)
			);`,
		},
	}
)

// sqliteStore implements the Store interface for SQLite.
type sqliteStore struct {
	logger   log.Logger
	driver   string
	database string
	meta     url.Values
	handle   *sqlx.DB
}

func init() {
	register("sqlite", NewSqliteStore)
	register("sqlite3", NewSqliteStore)
}

// Open simply opens the database connection.
func (s *sqliteStore) Open() (err error) {
	if dir := path.Dir(s.database); dir != "." {
		if err := os.MkdirAll(dir, 0770); err != nil {
			return fmt.Errorf("failed to create database dir: %w", err)
		}
	}

	s.handle, err = sqlx.Open(
		s.driver,
		s.dsn(),
	)

	if err != nil {
		return err
	}

	return nil
}

// Close simply closes the database connection.
func (s *sqliteStore) Close() error {
	return s.handle.Close()
}

// Ping just tests the database connection.
func (s *sqliteStore) Ping() error {
	return s.handle.Ping()
}

// Migrate executes required db migrations.
func (s *sqliteStore) Migrate() error {
	driver := darwin.New(
		darwin.NewGenericDriver(
			s.handle.DB,
			dialect.SqliteDialect{},
		),
		sqliteMigrations,
		nil,
	)

	if err := driver.Migrate(); err != nil {
		fmt.Printf("%+v\n", err)
		return fmt.Errorf("failed to exec migrations: %w", err)
	}

	return nil
}

// StoreWorkflowRunEvent implements the Store interface.
func (s *sqliteStore) StoreWorkflowRunEvent(event *github.WorkflowRunEvent) error {
	return storeWorkflowRunEvent(s.handle, event)
}

// GetWorkflowRuns implements the Store interface.
func (s *sqliteStore) GetWorkflowRuns() ([]*WorkflowRun, error) {
	return getWorkflowRuns(s.handle)
}

// PruneWorkflowRuns implements the Store interface.
func (s *sqliteStore) PruneWorkflowRuns(timeframe time.Duration) error {
	return pruneWorkflowRuns(s.handle, timeframe)
}

func (s *sqliteStore) dsn() string {
	if len(s.meta) > 0 {
		return fmt.Sprintf(
			"%s?%s",
			s.database,
			s.meta.Encode(),
		)
	}

	return s.database
}

// NewSqliteStore initializes a new SQLite store.
func NewSqliteStore(dsn string, logger log.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	client := &sqliteStore{
		logger:   logger,
		driver:   "sqlite",
		database: path.Join(parsed.Host, parsed.Path),
		meta:     parsed.Query(),
	}

	client.meta.Add("_pragma", "journal_mode(WAL)")
	client.meta.Add("_pragma", "busy_timeout(5000)")
	client.meta.Add("_pragma", "foreign_keys(1)")

	return client, nil
}
