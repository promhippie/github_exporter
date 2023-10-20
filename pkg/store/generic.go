package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/go-github/v56/github"
	"github.com/promhippie/github_exporter/pkg/config"

	// Import SQLite driver for database/sql
	_ "modernc.org/sqlite"

	// Import MySQL driver for database/sql
	_ "github.com/go-sql-driver/mysql"

	// Import PostgreSQL driver for database/sql
	_ "github.com/lib/pq"
)

//go:embed sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed mysql/*.sql
var mysqlMigrations embed.FS

//go:embed postgres/*.sql
var postgresMigrations embed.FS

// genericStore implements the Store interface for Sqlite.
type genericStore struct {
	logger          log.Logger
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
	handle          *sql.DB
}

// StoreWorkflowRunEvent handles workflow_run events from GitHub.
func (s *genericStore) StoreWorkflowRunEvent(event *github.WorkflowRunEvent) error {
	createdAt := event.GetWorkflowRun().GetCreatedAt().Time
	updatedAt := event.GetWorkflowRun().GetUpdatedAt().Time
	startedAt := event.GetWorkflowRun().GetRunStartedAt().Time

	record := &WorkflowRun{
		Owner:      event.GetRepo().GetOwner().GetLogin(),
		Repo:       event.GetRepo().GetName(),
		WorkflowID: event.GetWorkflowRun().GetWorkflowID(),
		Number:     event.GetWorkflowRun().GetRunNumber(),
		Attempt:    event.GetWorkflowRun().GetRunAttempt(),
		Event:      event.GetWorkflowRun().GetEvent(),
		Name:       event.GetWorkflowRun().GetName(),
		Title:      event.GetWorkflowRun().GetDisplayTitle(),
		Status:     event.GetWorkflowRun().GetConclusion(),
		Branch:     event.GetWorkflowRun().GetHeadBranch(),
		SHA:        event.GetWorkflowRun().GetHeadSHA(),
		Identifier: event.GetWorkflowRun().GetID(),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		StartedAt:  startedAt,
	}

	if record.Status == "" {
		record.Status = event.GetWorkflowRun().GetStatus()
	}

	return s.CreateOrUpdateWorkflowRun(record)
}

// CreateOrUpdateWorkflowRun creates or updates the record.
func (s *genericStore) CreateOrUpdateWorkflowRun(record *WorkflowRun) error {
	exists := WorkflowRun{}

	if err := s.handle.QueryRow(
		findWorkflowRunQuery,
		record.Owner,
		record.Repo,
		record.WorkflowID,
		record.Number,
	).Scan(&exists); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to find record: %w", err)
	}

	if exists.WorkflowID == 0 {
		if _, err := s.handle.Exec(
			createWorkflowRunQuery,
			record.Owner,
			record.Repo,
			record.WorkflowID,
			record.Number,
			record.Attempt,
			record.Event,
			record.Name,
			record.Title,
			record.Status,
			record.Branch,
			record.SHA,
			record.Identifier,
			record.CreatedAt,
			record.UpdatedAt,
			record.StartedAt,
		); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
	} else {
		if _, err := s.handle.Exec(
			updateWorkflowRunQuery,
			record.Owner,
			record.Repo,
			record.WorkflowID,
			record.Number,
			record.Attempt,
			record.Event,
			record.Name,
			record.Title,
			record.Status,
			record.Branch,
			record.SHA,
			record.Identifier,
			record.CreatedAt,
			record.UpdatedAt,
			record.StartedAt,
		); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	return nil
}

// GetWorkflowRuns retrieves the workflow runs from the database.
func (s *genericStore) GetWorkflowRuns() ([]*WorkflowRun, error) {
	records := make([]*WorkflowRun, 0)

	rows, err := s.handle.Query(
		selectWorkflowRunsQuery,
	)

	if err != nil {
		return records, err
	}

	defer rows.Close()

	for rows.Next() {
		record := &WorkflowRun{}

		if err := rows.Scan(
			record.Owner,
			record.Repo,
			record.WorkflowID,
			record.Number,
			record.Attempt,
			record.Event,
			record.Name,
			record.Title,
			record.Status,
			record.Branch,
			record.SHA,
			record.Identifier,
			record.CreatedAt,
			record.UpdatedAt,
			record.StartedAt,
		); err != nil {
			return records, err
		}

		records = append(
			records,
			record,
		)
	}

	if err := rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

// PruneWorkflowRuns prunes older workflow run records.
func (s *genericStore) PruneWorkflowRuns(timeframe time.Duration) error {
	if _, err := s.handle.Exec(
		purgeWorkflowRunsQuery,
		time.Now().Add(-timeframe),
	); err != nil {
		return fmt.Errorf("failed to prune workflow runs: %w", err)
	}

	return nil
}

// Open simply opens the database connection.
func (s *genericStore) Open() (err error) {
	s.handle, err = sql.Open(
		s.driver,
		s.dsn(),
	)

	if err != nil {
		return err
	}

	switch s.driver {
	case "mysql":
		s.handle.SetMaxOpenConns(s.maxOpenConns)
		s.handle.SetMaxIdleConns(s.maxIdleConns)
		s.handle.SetConnMaxLifetime(s.connMaxLifetime)
	case "postgres":
		s.handle.SetMaxOpenConns(s.maxOpenConns)
		s.handle.SetMaxIdleConns(s.maxIdleConns)
		s.handle.SetConnMaxLifetime(s.connMaxLifetime)
	}

	return nil
}

// Close simply closes the database connection.
func (s *genericStore) Close() error {
	return s.handle.Close()
}

// Ping just tests the database connection.
func (s *genericStore) Ping() error {
	return s.handle.Ping()
}

// Migrate executes required db migrations.
func (s *genericStore) Migrate() error {
	var (
		driver database.Driver
		source source.Driver
	)

	switch s.driver {
	case "sqlite":
		drv, err := sqlite.WithInstance(
			s.handle,
			&sqlite.Config{},
		)

		if err != nil {
			return fmt.Errorf("failed to prepare driver: %w", err)
		}

		src, err := iofs.New(
			sqliteMigrations,
			"sqlite",
		)

		if err != nil {
			return fmt.Errorf("failed to load migrations: %w", err)
		}

		driver = drv
		source = src
	case "mysql":
		drv, err := mysql.WithInstance(
			s.handle,
			&mysql.Config{},
		)

		if err != nil {
			return fmt.Errorf("failed to prepare driver: %w", err)
		}

		src, err := iofs.New(
			mysqlMigrations,
			"mysql",
		)

		if err != nil {
			return fmt.Errorf("failed to load migrations: %w", err)
		}

		driver = drv
		source = src
	case "postgres":
		drv, err := postgres.WithInstance(
			s.handle,
			&postgres.Config{},
		)

		if err != nil {
			return fmt.Errorf("failed to prepare driver: %w", err)
		}

		src, err := iofs.New(
			postgresMigrations,
			"postgres",
		)

		if err != nil {
			return fmt.Errorf("failed to load migrations: %w", err)
		}

		driver = drv
		source = src
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		s.database,
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to init migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to exec migrations: %w", err)
	}

	return nil
}

func (s *genericStore) dsn() string {
	switch s.driver {
	case "sqlite":
		return fmt.Sprintf(
			"%s?%s",
			s.database,
			s.meta.Encode(),
		)
	case "mysql":
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
	case "postgres":
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

	return ""
}

// NewGenericStore initializes a new generic store.
func NewGenericStore(cfg config.Database, logger log.Logger) (Store, error) {
	parsed, err := url.Parse(cfg.DSN)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	client := &genericStore{
		logger:   logger,
		driver:   parsed.Scheme,
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

	switch client.driver {
	case "sqlite", "sqlite3":
		client.driver = "sqlite"
		client.database = path.Join(parsed.Host, parsed.Path)

		client.meta.Add("_pragma", "journal_mode(WAL)")
		client.meta.Add("_pragma", "busy_timeout(5000)")
		client.meta.Add("_pragma", "foreign_keys(1)")
	case "mysql", "mariadb":
		client.driver = "mysql"
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
	case "postgres", "postgresql":
		client.driver = "postgres"
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
	}

	return client, nil
}

var selectWorkflowRunsQuery = `
SELECT
	*
FROM
	workflow_runs
ORDER BY
	updated_at ASC;`

var findWorkflowRunQuery = `
SELECT
	*
FROM
	workflow_runs
WHERE
	owner=$1 AND repo=$2 AND workflow_id=$3 AND number=$4;`

var createWorkflowRunQuery = `
INSERT INTO workflow_runs (
	owner,
	repo,
	workflow_id,
	number,
	attempt,
	event,
	name,
	title,
	status,
	branch,
	sha,
	identifier,
	created_at,
	updated_at,
	started_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15
);`

var updateWorkflowRunQuery = `
UPDATE
	workflow_runs
SET
	attempt=$5,
	event=$6,
	name=$7,
	title=$8,
	status=$9,
	branch=$10,
	sha=$11,
	identifier=$12,
	created_at=$13,
	updated_at=$14,
	started_at=$15
WHERE
	owner=$1 AND repo=$2 AND workflow_id=$3 AND number=$4;`

var purgeWorkflowRunsQuery = `
DELETE FROM
	workflow_runs
WHERE
	updated_at < $1;`
