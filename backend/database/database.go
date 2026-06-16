package database

import (
	"embed"
	"errors"
	"log/slog"
	"os"

	"github.com/KENTA0326/run-sync-pro/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var DB *gorm.DB

// getDSN は環境変数 DATABASE_URL があればそれを使い、なければ Docker 用デフォルトを返す
func getDSN() string {
	return config.ResolveString("", "DATABASE_URL", "host=db user=user password=password dbname=runsync_db port=5432 sslmode=disable TimeZone=Asia/Tokyo")
}

// getPostgresURL は GORM 用 DSN を golang-migrate 用の URL 形式に変換する（簡易版: DSN のままでは動かないので URL を返す）
func getPostgresURL() string {
	return config.ResolveString("", "DATABASE_URL", "postgres://user:password@db:5432/runsync_db?sslmode=disable&TimeZone=Asia/Tokyo")
}

func Connect() {
	postgresURL := getPostgresURL()

	// 1. バージョン管理されたマイグレーションを実行
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		slog.Error("migrations_source_failed", slog.Any("err", err))
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, postgresURL)
	if err != nil {
		slog.Error("migrations_instance_failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("migrations_up_failed", slog.Any("err", err))
		os.Exit(1)
	}
	if err == nil {
		slog.Info("migrations_applied")
	} else {
		slog.Info("migrations_no_change")
	}

	// 2. GORM で DB 接続（アプリ用）
	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("database_connect_failed", slog.Any("err", err))
		os.Exit(1)
	}

	RegisterCallbacks(db)

	slog.Info("database_connected")
	DB = db
}
