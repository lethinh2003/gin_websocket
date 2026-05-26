package config

import (
	"log"
	"os"

	"go_server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Require DATABASE_URL in .env like: host=localhost user=postgres password=root dbname=learning_go port=5432 sslmode=disable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not found, using default postgres connection string")
		dsn = "host=localhost user=postgres password=root dbname=learning_go port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(
		&models.User{},
		&models.Keno{},
		&models.LichSuDatCuocKeno{},
		&models.ChiTietDatCuocKeno{},
		&models.SystemSetting{},
	)

	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	// Seed Initial Values
	if err := models.SeedSystemSettings(db); err != nil {
		log.Fatalf("Failed to seed system settings: %v", err)
	}

	DB = db
	log.Println("Database connection established")
}
