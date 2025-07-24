package main

import (
	"greenbone-case-study/pkg/notifications"
	"log"
	"os"
)

func main() {
	port := getEnv("PORT", "9090")
	log.Printf("Starting mock notification server on port %s", port)
	notifications.StartMockNotificationServer(port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
