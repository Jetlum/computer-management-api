package db

import (
	"log"

	"greenbone-case-study/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabase initializes the database connection
func InitDatabase(databaseURL string, dbType string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch dbType {
	case "postgres":
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	default:
		log.Fatal("Unsupported database type")
	}

	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Computer{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
