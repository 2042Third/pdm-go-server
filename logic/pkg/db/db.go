package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type Database struct {
	DB *gorm.DB
}

// NewDatabase initializes and returns a new Database instance.
func NewDatabase() *Database {
	// Retrieve environment variables
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSL_MODE")

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	// Initialize connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Connected to the database successfully!")

	// Run migrations
	//if err := db.AutoMigrate(&models.SessionKey{}); err != nil {
	//	log.Fatalf("Failed to migrate database schema: %v", err)
	//}

	//if err := db.AutoMigrate(&models.RefreshKey{}); err != nil {
	//	log.Fatalf("Failed to migrate database schema: %v", err)
	//}
	//
	//log.Println("Database migration completed!")

	return &Database{DB: db}
}
