package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("DB_PATH")
	os.Unsetenv("DB_MAX_OPEN_CONNS")
	os.Unsetenv("DB_MAX_IDLE_CONNS")
	os.Unsetenv("DB_JOURNAL_MODE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DBPath != "./data.db" {
		t.Errorf("expected DBPath './data.db', got '%s'", cfg.DBPath)
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Errorf("expected DBMaxOpenConns 25, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 5 {
		t.Errorf("expected DBMaxIdleConns 5, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBJournalMode != "WAL" {
		t.Errorf("expected DBJournalMode 'WAL', got '%s'", cfg.DBJournalMode)
	}
	if cfg.DBBusyTimeoutMs != 5000 {
		t.Errorf("expected DBBusyTimeoutMs 5000, got %d", cfg.DBBusyTimeoutMs)
	}
	if cfg.DBSynchronous != "NORMAL" {
		t.Errorf("expected DBSynchronous 'NORMAL', got '%s'", cfg.DBSynchronous)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("DB_PATH", "/tmp/test.db")
	os.Setenv("DB_MAX_OPEN_CONNS", "10")
	os.Setenv("DB_JOURNAL_MODE", "DELETE")
	defer func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_JOURNAL_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected DBPath '/tmp/test.db', got '%s'", cfg.DBPath)
	}
	if cfg.DBMaxOpenConns != 10 {
		t.Errorf("expected DBMaxOpenConns 10, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBJournalMode != "DELETE" {
		t.Errorf("expected DBJournalMode 'DELETE', got '%s'", cfg.DBJournalMode)
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBPath:          "/tmp/test.db",
		DBJournalMode:   "WAL",
		DBBusyTimeoutMs: 5000,
		DBCacheSize:     -2000,
		DBSynchronous:   "NORMAL",
	}

	dsn := cfg.DSN()
	expected := "/tmp/test.db?_journal_mode=WAL&_busy_timeout=5000&_cache_size=-2000&_synchronous=NORMAL"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestDSN_InMemory(t *testing.T) {
	cfg := &Config{
		DBPath: ":memory:",
	}

	dsn := cfg.DSN()
	if dsn != ":memory:?cache=shared" {
		t.Errorf("expected DSN ':memory:?cache=shared', got '%s'", dsn)
	}
}

func TestOpenDB_InMemory(t *testing.T) {
	cfg := &Config{
		DBPath:          ":memory:",
		DBMaxOpenConns:  5,
		DBMaxIdleConns:  2,
		DBJournalMode:   "WAL",
		DBBusyTimeoutMs: 5000,
		DBCacheSize:     -2000,
		DBSynchronous:   "NORMAL",
	}

	db, err := cfg.OpenDB()
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify we can execute a query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("failed to query database: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestMustLoad_Success(t *testing.T) {
	os.Unsetenv("DB_PATH")

	// Should not panic with valid defaults
	cfg := MustLoad()
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

func TestMustOpenDB_InMemory(t *testing.T) {
	cfg := &Config{
		DBPath:          ":memory:",
		DBMaxOpenConns:  5,
		DBMaxIdleConns:  2,
		DBJournalMode:   "WAL",
		DBBusyTimeoutMs: 5000,
		DBCacheSize:     -2000,
		DBSynchronous:   "NORMAL",
	}

	db := cfg.MustOpenDB()
	defer db.Close()

	if db == nil {
		t.Error("expected non-nil db")
	}
}
