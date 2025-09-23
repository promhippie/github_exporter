package store

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GuiaBolso/darwin"
	"github.com/google/go-github/v75/github"
	"github.com/jmoiron/sqlx"
	"github.com/promhippie/github_exporter/pkg/migration/dialect"

	// Import PostgreSQL driver for database/sql
	_ "github.com/lib/pq"
)

var (
	postgresMigrations = []darwin.Migration{
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
		{
			Version:     2,
			Description: "Adding actor column to workflow_runs table",
			Script:      `ALTER TABLE workflow_runs ADD COLUMN actor TEXT;`,
		},
		{
			Version:     3,
			Description: "Creating table workflow_jobs",
			Script: `CREATE TABLE workflow_jobs (
				owner TEXT NOT NULL,
				repo TEXT NOT NULL,
				name TEXT,
				status TEXT,
				branch TEXT,
				sha TEXT,
				conclusion TEXT,
				labels TEXT,
				identifier INTEGER,
				run_id INTEGER NOT NULL,
				run_attempt INTEGER NOT NULL,
				created_at BIGINT,
				started_at BIGINT,
				completed_at BIGINT,
				runner_id INTEGER,
				runner_name TEXT,
				runner_group_id INTEGER,
				runner_group_name TEXT,
				workflow_name TEXT,
				PRIMARY KEY(owner, repo, identifier)
			);`,
		},
		{
			Version:     4,
			Description: "Fix identifier be BIGINT",
			Script:      `ALTER TABLE workflow_jobs ALTER COLUMN identifier TYPE BIGINT USING identifier::BIGINT;`,
		},
		{
			Version:     5,
			Description: "Fix run_id be BIGINT",
			Script:      `ALTER TABLE workflow_jobs ALTER COLUMN run_id TYPE BIGINT USING run_id::BIGINT;`,
		},
	}
)

func init() {
	register("postgres", NewPostgresStore)
	register("postgresql", NewPostgresStore)
}

// postgresStore implements the Store interface for PostgreSQL.
type postgresStore struct {
	logger          *slog.Logger
	driver          string
	host            string
	port            string
	username        string
	password        string
	database        string
	meta            url.Values
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	handle          *sqlx.DB
}

// Open simply opens the database connection.
func (s *postgresStore) Open() (res bool, err error) {
	s.handle, err = sqlx.Open(
		s.driver,
		s.dsn(),
	)

	if err != nil {
		return false, err
	}

	s.handle.SetMaxOpenConns(s.maxOpenConns)
	s.handle.SetMaxIdleConns(s.maxIdleConns)
	s.handle.SetConnMaxLifetime(s.connMaxLifetime)

	return true, nil
}

// Close simply closes the database connection.
func (s *postgresStore) Close() error {
	return s.handle.Close()
}

// Ping just tests the database connection.
func (s *postgresStore) Ping() (bool, error) {
	if err := s.handle.Ping(); err != nil {
		return false, err
	}

	return true, nil
}

// Migrate executes required db migrations.
func (s *postgresStore) Migrate() error {
	driver := darwin.New(
		darwin.NewGenericDriver(
			s.handle.DB,
			dialect.PostgresDialect{},
		),
		postgresMigrations,
		nil,
	)

	if err := driver.Migrate(); err != nil {
		return fmt.Errorf("failed to exec migrations: %w", err)
	}

	return nil
}

// StoreWorkflowRunEvent implements the Store interface.
func (s *postgresStore) StoreWorkflowRunEvent(event *github.WorkflowRunEvent) error {
	return storeWorkflowRunEvent(s.handle, event)
}

// GetWorkflowRuns implements the Store interface.
func (s *postgresStore) GetWorkflowRuns(window time.Duration) ([]*WorkflowRun, error) {
	return getWorkflowRuns(s.handle, window)
}

// PruneWorkflowRuns implements the Store interface.
func (s *postgresStore) PruneWorkflowRuns(timeframe time.Duration) error {
	return pruneWorkflowRuns(s.handle, timeframe)
}

// StoreWorkflowJobEvent implements the Store interface.
func (s *postgresStore) StoreWorkflowJobEvent(event *github.WorkflowJobEvent) error {
	return storeWorkflowJobEvent(s.handle, event)
}

// GetWorkflowJobs implements the Store interface.
func (s *postgresStore) GetWorkflowJobs(window time.Duration) ([]*WorkflowJob, error) {
	return getWorkflowJobs(s.handle, window)
}

// PruneWorkflowJobs implements the Store interface.
func (s *postgresStore) PruneWorkflowJobs(timeframe time.Duration) error {
	return pruneWorkflowJobs(s.handle, timeframe)
}

func (s *postgresStore) dsn() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s",
		s.host,
		s.port,
		s.database,
		s.username,
	)

	if s.password != "" {
		dsn = fmt.Sprintf(
			"%s password=%s",
			dsn,
			s.password,
		)
	}

	for key, val := range s.meta {
		dsn = fmt.Sprintf("%s %s=%s", dsn, key, strings.Join(val, ""))
	}

	return dsn
}

// NewPostgresStore initializes a new PostgreSQL store.
func NewPostgresStore(dsn string, logger *slog.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	client := &postgresStore{
		logger:   logger,
		driver:   "postgres",
		username: parsed.User.Username(),
		meta:     parsed.Query(),
	}

	if password, ok := parsed.User.Password(); ok {
		client.password = password
	}

	if client.meta.Has("maxOpenConns") {
		val, err := strconv.Atoi(
			client.meta.Get("maxOpenConns"),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse maxOpenConns: %w", err)
		}

		client.maxOpenConns = val
		client.meta.Del("maxOpenConns")
	} else {
		client.maxOpenConns = 25
	}

	if client.meta.Has("maxIdleConns") {
		val, err := strconv.Atoi(
			client.meta.Get("maxIdleConns"),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse maxIdleConns: %w", err)
		}

		client.maxIdleConns = val
		client.meta.Del("maxIdleConns")
	} else {
		client.maxIdleConns = 25
	}

	if client.meta.Has("connMaxLifetime") {
		val, err := time.ParseDuration(
			client.meta.Get("connMaxLifetime"),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse connMaxLifetime: %w", err)
		}

		client.connMaxLifetime = val
		client.meta.Del("connMaxLifetime")
	} else {
		client.connMaxLifetime = 5 * time.Minute
	}

	client.database = strings.TrimPrefix(parsed.Path, "/")

	host, port, err := net.SplitHostPort(parsed.Host)

	if err != nil && strings.Contains(err.Error(), "missing port in address") {
		client.host = parsed.Host
		client.port = "5432"
	} else if err != nil {
		return nil, err
	} else {
		client.host = host
		client.port = port
	}

	if val := client.meta.Get("sslmode"); val == "" {
		client.meta.Set("sslmode", "disable")
	}

	return client, nil
}
