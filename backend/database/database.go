package database

import (
	"embed"
	"fmt"
	"log"

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
		log.Fatal("Failed to open migrations source:", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, postgresURL)
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed to run migrations:", err)
	}
	if err == nil {
		fmt.Println("Migrations applied successfully.")
	}

	// 2. GORM で DB 接続（アプリ用）
	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection successful!")
	DB = db
}
