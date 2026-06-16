package database

import (
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const migrationVersionCount = 6

func TestMigrationsReversible(t *testing.T) {
	t.Parallel()

	postgresURL := os.Getenv("DATABASE_URL")
	if postgresURL == "" {
		t.Skip("DATABASE_URL が未設定のためスキップ（CI では postgres サービスと併用）")
	}

	m := newTestMigrator(t, postgresURL)
	t.Cleanup(func() { _, _ = m.Close() })

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrate up: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("version after up: %v", err)
	}
	if dirty {
		t.Fatal("migration dirty after up")
	}
	if version != migrationVersionCount {
		t.Fatalf("version = %d, want %d", version, migrationVersionCount)
	}

	for step := 0; step < migrationVersionCount; step++ {
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("migrate down step %d: %v", step+1, err)
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrate up again: %v", err)
	}

	version, dirty, err = m.Version()
	if err != nil {
		t.Fatalf("version after re-up: %v", err)
	}
	if dirty {
		t.Fatal("migration dirty after re-up")
	}
	if version != migrationVersionCount {
		t.Fatalf("version after re-up = %d, want %d", version, migrationVersionCount)
	}
}

func newTestMigrator(t *testing.T, postgresURL string) *migrate.Migrate {
	t.Helper()

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("migrations source: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, postgresURL)
	if err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	return m
}
