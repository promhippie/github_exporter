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
	"github.com/google/go-github/v72/github"
	"github.com/jmoiron/sqlx"
	"github.com/promhippie/github_exporter/pkg/migration/dialect"

	// Import MySQL driver for database/sql
	_ "github.com/go-sql-driver/mysql"
)

var (
	mysqlMigrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table workflow_runs",
			Script: `CREATE TABLE workflow_runs (
				owner VARCHAR(255) NOT NULL,
				repo VARCHAR(255) NOT NULL,
				workflow_id INTEGER NOT NULL,
				number INTEGER NOT NULL,
				attempt INTEGER,
				event VARCHAR(255),
				name VARCHAR(255),
				title VARCHAR(255),
				status VARCHAR(255),
				branch VARCHAR(255),
				sha VARCHAR(255),
				identifier BIGINT,
				created_at BIGINT,
				updated_at BIGINT,
				started_at BIGINT,
				PRIMARY KEY(owner, repo, workflow_id, number)
			) ENGINE=InnoDB CHARACTER SET=utf8;`,
		},
		{
			Version:     2,
			Description: "Altering table workflow_runs to add actor column",
			Script:      `ALTER TABLE workflow_runs ADD COLUMN actor VARCHAR(255);`,
		},
		{
			Version:     3,
			Description: "Creating table workflow_jobs",
			Script: `CREATE TABLE workflow_jobs (
				owner VARCHAR(255) NOT NULL,
				repo VARCHAR(255) NOT NULL,
				name VARCHAR(255),
				status VARCHAR(255),
				branch VARCHAR(255),
				sha VARCHAR(255),
				conclusion VARCHAR(255),
				labels VARCHAR(255),
				identifier INTEGER,
				run_id INTEGER NOT NULL,
				run_attempt INTEGER NOT NULL,
				created_at BIGINT,
				started_at BIGINT,
				completed_at BIGINT,
				runner_id INTEGER,
				runner_name VARCHAR(255),
				runner_group_id INTEGER,
				runner_group_name VARCHAR(255),
				workflow_name VARCHAR(255),
				PRIMARY KEY(owner, repo, identifier)
			);`,
		},
	}
)

func init() {
	register("mysql", NewMysqlStore)
	register("mariadb", NewMysqlStore)
}

// mysqlStore implements the Store interface for MySQL.
type mysqlStore struct {
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
func (s *mysqlStore) Open() (res bool, err error) {
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
func (s *mysqlStore) Close() error {
	return s.handle.Close()
}

// Ping just tests the database connection.
func (s *mysqlStore) Ping() (bool, error) {
	if err := s.handle.Ping(); err != nil {
		return false, err
	}

	return true, nil
}

// Migrate executes required db migrations.
func (s *mysqlStore) Migrate() error {
	driver := darwin.New(
		darwin.NewGenericDriver(
			s.handle.DB,
			dialect.MySQLDialect{},
		),
		mysqlMigrations,
		nil,
	)

	if err := driver.Migrate(); err != nil {
		return fmt.Errorf("failed to exec migrations: %w", err)
	}

	return nil
}

// StoreWorkflowRunEvent implements the Store interface.
func (s *mysqlStore) StoreWorkflowRunEvent(event *github.WorkflowRunEvent) error {
	return storeWorkflowRunEvent(s.handle, event)
}

// GetWorkflowRuns implements the Store interface.
func (s *mysqlStore) GetWorkflowRuns(window time.Duration) ([]*WorkflowRun, error) {
	return getWorkflowRuns(s.handle, window)
}

// PruneWorkflowRuns implements the Store interface.
func (s *mysqlStore) PruneWorkflowRuns(timeframe time.Duration) error {
	return pruneWorkflowRuns(s.handle, timeframe)
}

// StoreWorkflowJobEvent implements the Store interface.
func (s *mysqlStore) StoreWorkflowJobEvent(event *github.WorkflowJobEvent) error {
	return storeWorkflowJobEvent(s.handle, event)
}

// GetWorkflowJobs implements the Store interface.
func (s *mysqlStore) GetWorkflowJobs(window time.Duration) ([]*WorkflowJob, error) {
	return getWorkflowJobs(s.handle, window)
}

// PruneWorkflowJobs implements the Store interface.
func (s *mysqlStore) PruneWorkflowJobs(timeframe time.Duration) error {
	return pruneWorkflowJobs(s.handle, timeframe)
}

func (s *mysqlStore) dsn() string {
	if s.password != "" {
		return fmt.Sprintf(
			"%s:%s@(%s:%s)/%s?%s",
			s.username,
			s.password,
			s.host,
			s.port,
			s.database,
			s.meta.Encode(),
		)
	}

	return fmt.Sprintf(
		"%s@(%s:%s)/%s?%s",
		s.username,
		s.host,
		s.port,
		s.database,
		s.meta.Encode(),
	)
}

// NewMysqlStore initializes a new MySQL store.
func NewMysqlStore(dsn string, logger *slog.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	client := &mysqlStore{
		logger:   logger,
		driver:   "mysql",
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
		client.port = "3306"
	} else if err != nil {
		return nil, err
	} else {
		client.host = host
		client.port = port
	}

	if val := client.meta.Get("charset"); val == "" {
		client.meta.Set("charset", "utf8")
	}

	if val := client.meta.Get("parseTime"); val == "" {
		client.meta.Set("parseTime", "True")
	}

	if val := client.meta.Get("loc"); val == "" {
		client.meta.Set("loc", "Local")
	}

	return client, nil
}
