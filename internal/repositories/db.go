package repositories

import (
	"fmt"
	"log"
	"time"

	"github.com/devops-360-online/go-with-me/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBUser,
		cfg.DBPort,
	)

	var db *gorm.DB
	var err error

	// Retry logic
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Optional: Ping the database to ensure connection is established
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				return db, nil
			}
		}
		log.Printf("Failed to connect to database (attempt %d/5): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to the database after several attempts: %v", err)
}
