package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Build represents a single factory build record.
type Build struct {
	ID        string
	RepoName  string
	Status    string // "shipped" or "failed"
	ErrorLog  string
	CreatedAt time.Time
	Duration  time.Duration
	HasTests  bool
	HasReadme bool
	ModPath   string
}

// Registry reads build records from the factory's SQLite database.
type Registry struct {
	db *sql.DB
}

// Open connects to the SQLite registry at the given path.
func Open(path string) (*Registry, error) {
	db, err := sql.Open("sqlite3", path+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open registry: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping registry: %w", err)
	}
	return &Registry{db: db}, nil
}

// Close closes the database connection.
func (r *Registry) Close() error {
	return r.db.Close()
}

// AllBuilds returns all builds from the registry.
// Supports both the legacy "builds" table and the current "build_queue" table.
func (r *Registry) AllBuilds() ([]Build, error) {
	// Try the legacy "builds" table first.
	rows, err := r.db.Query(`
		SELECT id, repo_name, status, COALESCE(error_log, ''),
		       created_at, COALESCE(duration_sec, 0),
		       COALESCE(has_tests, 0), COALESCE(has_readme, 0),
		       COALESCE(mod_path, '')
		FROM builds
		ORDER BY created_at DESC
	`)
	if err == nil {
		defer rows.Close()
		return scanLegacyBuilds(rows)
	}

	// Fall back to "build_queue" table (current factory schema).
	rows2, err2 := r.db.Query(`
		SELECT CAST(id AS TEXT), name, status, COALESCE(error_log, ''),
		       COALESCE(queued_at, ''), language
		FROM build_queue
		ORDER BY queued_at DESC
	`)
	if err2 != nil {
		return nil, fmt.Errorf("query builds: %w (also tried legacy: %w)", err2, err)
	}
	defer rows2.Close()

	var builds []Build
	for rows2.Next() {
		var b Build
		var queuedAt, language string
		if err := rows2.Scan(&b.ID, &b.RepoName, &b.Status, &b.ErrorLog,
			&queuedAt, &language); err != nil {
			return nil, fmt.Errorf("scan build_queue: %w", err)
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339, queuedAt)
		if b.CreatedAt.IsZero() {
			b.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", queuedAt)
		}
		b.ModPath = language
		builds = append(builds, b)
	}
	return builds, rows2.Err()
}

func scanLegacyBuilds(rows *sql.Rows) ([]Build, error) {
	var builds []Build
	for rows.Next() {
		var b Build
		var createdAt string
		var durationSec int
		var hasTests, hasReadme int

		if err := rows.Scan(&b.ID, &b.RepoName, &b.Status, &b.ErrorLog,
			&createdAt, &durationSec, &hasTests, &hasReadme, &b.ModPath); err != nil {
			return nil, fmt.Errorf("scan build: %w", err)
		}

		b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		b.Duration = time.Duration(durationSec) * time.Second
		b.HasTests = hasTests == 1
		b.HasReadme = hasReadme == 1
		builds = append(builds, b)
	}
	return builds, rows.Err()
}

// ShippedBuilds returns only builds with status "shipped".
func (r *Registry) ShippedBuilds() ([]Build, error) {
	all, err := r.AllBuilds()
	if err != nil {
		return nil, err
	}
	var shipped []Build
	for _, b := range all {
		if b.Status == "shipped" {
			shipped = append(shipped, b)
		}
	}
	return shipped, nil
}

// FailedBuilds returns only builds with status "failed".
func (r *Registry) FailedBuilds() ([]Build, error) {
	all, err := r.AllBuilds()
	if err != nil {
		return nil, err
	}
	var failed []Build
	for _, b := range all {
		if b.Status == "failed" {
			failed = append(failed, b)
		}
	}
	return failed, nil
}
