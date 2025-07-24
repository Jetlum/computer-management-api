package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNotificationClient_SendNotification(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/notify" {
			t.Errorf("Expected /api/notify path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "received"}`))
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	err := client.SendNotification(notification)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNotificationClient_SendNotificationWithContext(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "received"}`))
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.SendNotificationWithContext(ctx, notification)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNotificationClient_SendNotificationRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "received"}`))
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	err := client.SendNotification(notification)
	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestNotificationClient_SendNotificationMaxRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	err := client.SendNotification(notification)
	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if !strings.Contains(err.Error(), "notification failed after 3 attempts") {
		t.Errorf("Expected max retries error message, got: %v", err)
	}
}

func TestNotificationClient_SendNotificationContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow server
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.SendNotificationWithContext(ctx, notification)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}

	if !strings.Contains(err.Error(), "notification cancelled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestNotificationClient_AddTimestamp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that timestamp was added to the notification
		decoder := json.NewDecoder(r.Body)
		var received Notification
		if err := decoder.Decode(&received); err != nil {
			t.Errorf("Failed to decode notification: %v", err)
		}

		if received.Timestamp == "" {
			t.Error("Expected timestamp to be set")
		}

		// Verify timestamp format
		if _, err := time.Parse(time.RFC3339, received.Timestamp); err != nil {
			t.Errorf("Invalid timestamp format: %s", received.Timestamp)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
		// No timestamp provided - should be added automatically
	}

	err := client.SendNotification(notification)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNotificationClient_PreserveExistingTimestamp(t *testing.T) {
	existingTimestamp := "2024-01-15T10:30:00Z"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var received Notification
		if err := decoder.Decode(&received); err != nil {
			t.Errorf("Failed to decode notification: %v", err)
		}

		if received.Timestamp != existingTimestamp {
			t.Errorf("Expected timestamp %s, got %s", existingTimestamp, received.Timestamp)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewNotificationClient(server.URL)
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
		Timestamp:            existingTimestamp,
	}

	err := client.SendNotification(notification)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNotificationClient_InvalidURL(t *testing.T) {
	client := NewNotificationClient("invalid-url")
	notification := Notification{
		Level:                "warning",
		EmployeeAbbreviation: "abc",
		Message:              "Test notification",
	}

	err := client.SendNotification(notification)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
