package notifications

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// StartMockNotificationServer starts a mock notification server for testing
func StartMockNotificationServer(port string) {
	http.HandleFunc("/api/notify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var notification Notification
		if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Mock notification received: %+v", notification)

		// Simulate some processing time
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "received",
			"id":     time.Now().Format("20060102150405"),
		})
	})

	log.Printf("Starting mock notification server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Mock notification server failed:", err)
	}
}
