//go:build genji

package store

import (
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/GuiaBolso/darwin"
	"github.com/go-kit/log"
	"github.com/google/go-github/v56/github"
	"github.com/jmoiron/sqlx"
	"github.com/promhippie/github_exporter/pkg/migration/dialect"

	// Import Genji driver for database/sql
	_ "github.com/genjidb/genji/driver"
)

var (
	genjiMigrations = []darwin.Migration{
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
				identifier INTEGER,
				created_at INTEGER,
				updated_at INTEGER,
				started_at INTEGER,
				PRIMARY KEY(owner, repo, workflow_id, number)
			);`,
		},
	}
)

func init() {
	register("genji", NewGenjiStore)
}

// genjiStore implements the Store interface for Genji.
type genjiStore struct {
	logger   log.Logger
	driver   string
	database string
	meta     url.Values
	handle   *sqlx.DB
}

// Open simply opens the database connection.
func (s *genjiStore) Open() (err error) {
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
func (s *genjiStore) Close() error {
	return s.handle.Close()
}

// Ping just tests the database connection.
func (s *genjiStore) Ping() error {
	return s.handle.Ping()
}

// Migrate executes required db migrations.
func (s *genjiStore) Migrate() error {
	driver := darwin.New(
		darwin.NewGenericDriver(
			s.handle.DB,
			dialect.GenjiDialect{},
		),
		genjiMigrations,
		nil,
	)

	if err := driver.Migrate(); err != nil {
		return fmt.Errorf("failed to exec migrations: %w", err)
	}

	return nil
}

// StoreWorkflowRunEvent implements the Store interface.
func (s *genjiStore) StoreWorkflowRunEvent(event *github.WorkflowRunEvent) error {
	return storeWorkflowRunEvent(s.handle, event)
}

// GetWorkflowRuns implements the Store interface.
func (s *genjiStore) GetWorkflowRuns() ([]*WorkflowRun, error) {
	return getWorkflowRuns(s.handle)
}

// PruneWorkflowRuns implements the Store interface.
func (s *genjiStore) PruneWorkflowRuns(timeframe time.Duration) error {
	return pruneWorkflowRuns(s.handle, timeframe)
}

func (s *genjiStore) dsn() string {
	if len(s.meta) > 0 {
		return fmt.Sprintf(
			"%s?%s",
			s.database,
			s.meta.Encode(),
		)
	}

	return s.database
}

// NewGenjiStore initializes a new MySQL store.
func NewGenjiStore(dsn string, logger log.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	client := &genjiStore{
		logger:   logger,
		driver:   "genji",
		database: path.Join(parsed.Host, parsed.Path),
		meta:     parsed.Query(),
	}

	return client, nil
}
