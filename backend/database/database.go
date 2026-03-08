package database

import (
	"fmt"
	"log"
	"github.com/KENTA0326/run-sync-pro/model" 
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=db user=user password=password dbname=runsync_db port=5432 sslmode=disable TimeZone=Asia/Tokyo"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自動でテーブルを作成する
	err = db.AutoMigrate(&model.User{}, &model.Shoe{}, &model.TrainingLog{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("Database connection and migration successful!")
	DB = db
}