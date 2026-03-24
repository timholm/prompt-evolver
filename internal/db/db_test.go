package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "prompt-evolver-db-test-*")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "registry.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE builds (
			id TEXT PRIMARY KEY,
			repo_name TEXT NOT NULL,
			status TEXT NOT NULL,
			error_log TEXT,
			created_at TEXT,
			duration_sec INTEGER,
			has_tests INTEGER DEFAULT 0,
			has_readme INTEGER DEFAULT 0,
			mod_path TEXT
		)
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO builds (id, repo_name, status, error_log, created_at, duration_sec, has_tests, has_readme, mod_path) VALUES
		('b1', 'url-shortener', 'shipped', '', '2026-03-20T10:00:00Z', 120, 1, 1, 'github.com/timholm/url-shortener'),
		('b2', 'log-parser', 'failed', 'no test files found', '2026-03-20T11:00:00Z', 90, 0, 1, ''),
		('b3', 'config-tool', 'shipped', '', '2026-03-20T12:00:00Z', 150, 1, 1, 'github.com/timholm/config-tool'),
		('b4', 'broken-api', 'failed', 'undefined: HandleRequest, build failed', '2026-03-20T13:00:00Z', 60, 0, 0, ''),
		('b5', 'port-scanner', 'failed', '--- FAIL: TestScan (0.01s) and timed out', '2026-03-20T14:00:00Z', 300, 1, 1, 'github.com/timholm/port-scanner')
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	db.Close()

	return dbPath, func() { os.RemoveAll(tmpDir) }
}

func TestOpenAndClose(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestOpenNonexistent(t *testing.T) {
	_, err := Open("/nonexistent/path/db.sqlite")
	if err == nil {
		t.Fatal("expected error opening nonexistent db")
	}
}

func TestAllBuilds(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close()

	builds, err := reg.AllBuilds()
	if err != nil {
		t.Fatalf("AllBuilds failed: %v", err)
	}
	if len(builds) != 5 {
		t.Errorf("expected 5 builds, got %d", len(builds))
	}
}

func TestShippedBuilds(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close()

	shipped, err := reg.ShippedBuilds()
	if err != nil {
		t.Fatal(err)
	}
	if len(shipped) != 2 {
		t.Errorf("expected 2 shipped builds, got %d", len(shipped))
	}
	for _, b := range shipped {
		if b.Status != "shipped" {
			t.Errorf("expected status 'shipped', got %q", b.Status)
		}
	}
}

func TestFailedBuilds(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close()

	failed, err := reg.FailedBuilds()
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 3 {
		t.Errorf("expected 3 failed builds, got %d", len(failed))
	}
	for _, b := range failed {
		if b.Status != "failed" {
			t.Errorf("expected status 'failed', got %q", b.Status)
		}
		if b.ErrorLog == "" {
			t.Error("failed build should have an error log")
		}
	}
}

func TestBuildFields(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reg, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Close()

	builds, err := reg.AllBuilds()
	if err != nil {
		t.Fatal(err)
	}

	// Find the shipped url-shortener.
	var found bool
	for _, b := range builds {
		if b.RepoName == "url-shortener" {
			found = true
			if !b.HasTests {
				t.Error("url-shortener should have tests")
			}
			if !b.HasReadme {
				t.Error("url-shortener should have readme")
			}
			if b.ModPath != "github.com/timholm/url-shortener" {
				t.Errorf("unexpected mod path: %q", b.ModPath)
			}
			if b.Duration.Seconds() != 120 {
				t.Errorf("expected 120s duration, got %v", b.Duration)
			}
		}
	}
	if !found {
		t.Error("url-shortener build not found")
	}
}
