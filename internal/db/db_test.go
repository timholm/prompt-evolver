package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "registry.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE builds (
			id TEXT PRIMARY KEY,
			repo_name TEXT NOT NULL,
			status TEXT NOT NULL,
			error_log TEXT,
			created_at TEXT NOT NULL,
			duration_sec INTEGER,
			has_tests INTEGER,
			has_readme INTEGER,
			mod_path TEXT
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return dbPath
}

func insertBuild(t *testing.T, dbPath string, id, repoName, status, errorLog, createdAt string, durationSec int, hasTests, hasReadme int, modPath string) {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db for insert: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO builds (id, repo_name, status, error_log, created_at, duration_sec, has_tests, has_readme, mod_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, repoName, status, errorLog, createdAt, durationSec, hasTests, hasReadme, modPath)
	if err != nil {
		t.Fatalf("insert build: %v", err)
	}
}

func TestOpenInvalidPath(t *testing.T) {
	_, err := Open("/nonexistent/path/registry.db")
	if err == nil {
		t.Error("expected error opening nonexistent database")
	}
}

func TestOpenAndClose(t *testing.T) {
	dbPath := createTestDB(t)
	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestAllBuildsEmpty(t *testing.T) {
	dbPath := createTestDB(t)
	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer reg.Close()

	builds, err := reg.AllBuilds()
	if err != nil {
		t.Fatalf("AllBuilds: %v", err)
	}
	if len(builds) != 0 {
		t.Errorf("expected 0 builds, got %d", len(builds))
	}
}

func TestAllBuildsWithData(t *testing.T) {
	dbPath := createTestDB(t)
	insertBuild(t, dbPath, "b1", "url-shortener", "shipped", "", "2025-01-15T10:00:00Z", 120, 1, 1, "github.com/test/url-shortener")
	insertBuild(t, dbPath, "b2", "log-parser", "failed", "build failed: syntax error", "2025-01-16T10:00:00Z", 30, 0, 0, "")
	insertBuild(t, dbPath, "b3", "config-merger", "shipped", "", "2025-01-17T10:00:00Z", 90, 1, 1, "github.com/test/config-merger")

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer reg.Close()

	builds, err := reg.AllBuilds()
	if err != nil {
		t.Fatalf("AllBuilds: %v", err)
	}
	if len(builds) != 3 {
		t.Fatalf("expected 3 builds, got %d", len(builds))
	}

	// Results ordered by created_at DESC.
	if builds[0].ID != "b3" {
		t.Errorf("first build ID = %q, want %q", builds[0].ID, "b3")
	}
	if builds[0].Status != "shipped" {
		t.Errorf("first build status = %q, want %q", builds[0].Status, "shipped")
	}
	if !builds[0].HasTests {
		t.Error("first build should have tests")
	}
	if !builds[0].HasReadme {
		t.Error("first build should have readme")
	}

	failedBuild := builds[1]
	if failedBuild.Status != "failed" {
		t.Errorf("b2 status = %q, want %q", failedBuild.Status, "failed")
	}
	if failedBuild.ErrorLog != "build failed: syntax error" {
		t.Errorf("b2 error_log = %q", failedBuild.ErrorLog)
	}
}

func TestShippedBuilds(t *testing.T) {
	dbPath := createTestDB(t)
	insertBuild(t, dbPath, "b1", "r1", "shipped", "", "2025-01-15T10:00:00Z", 60, 1, 1, "m")
	insertBuild(t, dbPath, "b2", "r2", "failed", "err", "2025-01-16T10:00:00Z", 30, 0, 0, "")
	insertBuild(t, dbPath, "b3", "r3", "shipped", "", "2025-01-17T10:00:00Z", 90, 1, 1, "m")

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer reg.Close()

	shipped, err := reg.ShippedBuilds()
	if err != nil {
		t.Fatalf("ShippedBuilds: %v", err)
	}
	if len(shipped) != 2 {
		t.Errorf("expected 2 shipped, got %d", len(shipped))
	}
	for _, b := range shipped {
		if b.Status != "shipped" {
			t.Errorf("non-shipped build in ShippedBuilds: %s", b.Status)
		}
	}
}

func TestFailedBuilds(t *testing.T) {
	dbPath := createTestDB(t)
	insertBuild(t, dbPath, "b1", "r1", "shipped", "", "2025-01-15T10:00:00Z", 60, 1, 1, "m")
	insertBuild(t, dbPath, "b2", "r2", "failed", "err", "2025-01-16T10:00:00Z", 30, 0, 0, "")

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer reg.Close()

	failed, err := reg.FailedBuilds()
	if err != nil {
		t.Fatalf("FailedBuilds: %v", err)
	}
	if len(failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(failed))
	}
	if failed[0].Status != "failed" {
		t.Errorf("expected failed status, got %q", failed[0].Status)
	}
}

func TestNullableFields(t *testing.T) {
	dbPath := createTestDB(t)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO builds (id, repo_name, status, created_at)
		VALUES ('b1', 'test-repo', 'failed', '2025-01-15T10:00:00Z')
	`)
	db.Close()
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	reg, errOpen := Open(dbPath)
	if errOpen != nil {
		t.Fatalf("Open: %v", errOpen)
	}
	defer reg.Close()

	builds, err := reg.AllBuilds()
	if err != nil {
		t.Fatalf("AllBuilds: %v", err)
	}
	if len(builds) != 1 {
		t.Fatalf("expected 1 build, got %d", len(builds))
	}

	b := builds[0]
	if b.ErrorLog != "" {
		t.Errorf("ErrorLog = %q, want empty", b.ErrorLog)
	}
	if b.HasTests {
		t.Error("HasTests should be false for NULL")
	}
	if b.HasReadme {
		t.Error("HasReadme should be false for NULL")
	}
	if b.ModPath != "" {
		t.Errorf("ModPath = %q, want empty", b.ModPath)
	}
}

func TestOpenReadOnly(t *testing.T) {
	dbPath := createTestDB(t)

	if err := os.Chmod(dbPath, 0o444); err != nil {
		t.Skipf("cannot set read-only permissions: %v", err)
	}
	defer os.Chmod(dbPath, 0o644)

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open read-only: %v", err)
	}
	reg.Close()
}
