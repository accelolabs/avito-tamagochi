package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthMigrationHasGooseSections(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "0001_auth.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read auth migration: %v", err)
	}

	sql := string(contents)
	for _, marker := range []string{"-- +goose Up", "-- +goose Down", "CREATE TABLE users", "CREATE TABLE sessions"} {
		if !strings.Contains(sql, marker) {
			t.Fatalf("auth migration does not contain %q", marker)
		}
	}
}

func TestGamePetMigrationHasCoreTablesAndConstraints(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "0002_game_pet.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read game migration: %v", err)
	}

	sql := string(contents)
	for _, marker := range []string{
		"-- +goose Up",
		"-- +goose Down",
		"CREATE TABLE pets",
		"CREATE TABLE xp_events",
		"UNIQUE (user_id, source_key)",
		"CREATE INDEX xp_events_user_date_idx",
	} {
		if !strings.Contains(sql, marker) {
			t.Fatalf("game migration does not contain %q", marker)
		}
	}
}
