package storage_util

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	var db *gorm.DB
	var err error

	dialect := os.Getenv("DB_DIALECT")
	if dialect == "" {
		dialect = "sqlite"
	}

	switch dialect {
	case "postgres":
		dsn := os.Getenv("POSTGRES_DSN")
		if dsn == "" {
			dsn = "host=localhost user=postgres dbname=search_logs password=secret sslmode=disable"
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	default:
		log.Fatal("Unsupported DB dialect")
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
