package database

import (
	"os"
	"testing"
	"time"
)

func TestConnectSQLite(t *testing.T) {
	dbFile := "test_connect.db"
	defer os.Remove(dbFile)

	db, err := Connect(Config{
		Driver:          "sqlite",
		DSN:             dbFile,
		MaxConns:        5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Verify connection is alive
	if err := db.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestConnectSQLiteWALMode(t *testing.T) {
	dbFile := "test_wal.db"
	defer os.Remove(dbFile)

	db, err := Connect(Config{
		Driver: "sqlite",
		DSN:    dbFile,
	})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Verify WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("expected journal_mode=wal, got %q", journalMode)
	}
}

func TestConnectSQLiteForeignKeys(t *testing.T) {
	dbFile := "test_fk.db"
	defer os.Remove(dbFile)

	db, err := Connect(Config{
		Driver: "sqlite",
		DSN:    dbFile,
	})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		t.Fatalf("failed to query foreign_keys: %v", err)
	}

	if fk != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fk)
	}
}

func TestConnectPoolSettings(t *testing.T) {
	dbFile := "test_pool.db"
	defer os.Remove(dbFile)

	db, err := Connect(Config{
		Driver:          "sqlite",
		DSN:             dbFile,
		MaxConns:        50,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	if stats.MaxOpenConnections != 50 {
		t.Errorf("expected MaxOpenConnections=50, got %d", stats.MaxOpenConnections)
	}
}

func TestConnectInvalidDriver(t *testing.T) {
	_, err := Connect(Config{
		Driver: "nonexistent_driver",
		DSN:    "test.db",
	})
	if err == nil {
		t.Error("expected error for invalid driver")
	}
}

func TestConnectSQLiteExecAndQuery(t *testing.T) {
	dbFile := "test_exec.db"
	defer os.Remove(dbFile)

	db, err := Connect(Config{
		Driver: "sqlite",
		DSN:    dbFile,
	})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec("CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert
	_, err = db.Exec("INSERT INTO test_items (name) VALUES (?)", "item1")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Query
	var name string
	err = db.QueryRow("SELECT name FROM test_items WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if name != "item1" {
		t.Errorf("expected name=%q, got %q", "item1", name)
	}
}
