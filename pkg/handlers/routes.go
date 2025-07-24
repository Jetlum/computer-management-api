package handlers

import (
	"greenbone-case-study/pkg/models"
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(service models.ComputerService) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// Create handler
	computerHandler := NewComputerHandler(service)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Computer routes
	api.HandleFunc("/computers", computerHandler.CreateComputer).Methods("POST")
	api.HandleFunc("/computers", computerHandler.GetAllComputers).Methods("GET")
	api.HandleFunc("/computers/{id}", computerHandler.GetComputerByID).Methods("GET")
	api.HandleFunc("/computers/{id}", computerHandler.UpdateComputer).Methods("PUT")
	api.HandleFunc("/computers/{id}", computerHandler.DeleteComputer).Methods("DELETE")

	// Employee routes
	api.HandleFunc("/employees/{abbr}/computers", computerHandler.GetComputersByEmployee).Methods("GET")

	// Health check endpoint
	api.HandleFunc("/health", healthCheckHandler).Methods("GET")

	return router
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "computer-management-api"}`))
}
