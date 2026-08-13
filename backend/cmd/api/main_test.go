package main

import (
	"database/sql"
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
