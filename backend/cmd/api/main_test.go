package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestConfigureDatabasePoolBoundsOpenConnections(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://unused")
	if err != nil {
		t.Fatalf("open database handle: %v", err)
	}
	defer db.Close()

	configureDatabasePool(db)

	if got := db.Stats().MaxOpenConnections; got != maxDatabaseConnections {
		t.Fatalf("max open connections = %d, want %d", got, maxDatabaseConnections)
	}
}

func TestEnvironmentOrDefault(t *testing.T) {
	const name = "TEST_NOTIFICATION_SETTING"
	t.Cleanup(func() { _ = os.Unsetenv(name) })
	if got := environmentOrDefault(name, "fallback"); got != "fallback" {
		t.Fatalf("unset value = %q, want fallback", got)
	}
	if err := os.Setenv(name, "configured"); err != nil {
		t.Fatalf("set environment: %v", err)
	}
	if got := environmentOrDefault(name, "fallback"); got != "configured" {
		t.Fatalf("configured value = %q, want configured", got)
	}
}
