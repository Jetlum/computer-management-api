package main

import (
	"greenbone-case-study/internal/db"
	"greenbone-case-study/pkg/handlers"
	"greenbone-case-study/pkg/models"
	"greenbone-case-study/pkg/notifications"
	"greenbone-case-study/pkg/services"
	"log"
	"net/http"
	"os"
)

func main() {
	// Get environment variables
	dbType := getEnv("DB_TYPE", "sqlite")
	dbURL := getEnv("DATABASE_URL", "computers.db")
	notificationURL := getEnv("NOTIFICATION_URL", "http://localhost:9090")
	port := getEnv("PORT", "8080")

	// Initialize database
	database, err := db.InitDatabase(dbURL, dbType)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize dependencies
	computerRepo := models.NewComputerRepository(database)
	notificationClient := notifications.NewNotificationClient(notificationURL)
	computerService := services.NewComputerService(computerRepo, notificationClient)

	// Setup routes
	router := handlers.SetupRoutes(computerService)

	// Start server
	log.Printf("Starting server on port %s", port)
	log.Printf("Database type: %s", dbType)
	log.Printf("Notification URL: %s", notificationURL)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
