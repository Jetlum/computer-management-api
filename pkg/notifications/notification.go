package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Notification represents a notification message
type Notification struct {
	Level                string `json:"level"`
	EmployeeAbbreviation string `json:"employeeAbbreviation"`
	Message              string `json:"message"`
	Timestamp            string `json:"timestamp,omitempty"`
}

// NotificationClient interface for sending notifications
type NotificationClient interface {
	SendNotification(notification Notification) error
	SendNotificationWithContext(ctx context.Context, notification Notification) error
}

type httpNotificationClient struct {
	client  *http.Client
	baseURL string
	logger  *log.Logger
}

// NewNotificationClient creates a new HTTP notification client
func NewNotificationClient(baseURL string) NotificationClient {
	return &httpNotificationClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		logger:  log.New(log.Writer(), "[NOTIFICATION] ", log.LstdFlags),
	}
}

// SendNotification sends a notification with retry logic
func (c *httpNotificationClient) SendNotification(notification Notification) error {
	return c.SendNotificationWithContext(context.Background(), notification)
}

// SendNotificationWithContext sends a notification with context and retry logic
func (c *httpNotificationClient) SendNotificationWithContext(ctx context.Context, notification Notification) error {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	// Add timestamp if not present
	if notification.Timestamp == "" {
		notification.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	c.logger.Printf("Sending notification for employee %s (level: %s)",
		notification.EmployeeAbbreviation, notification.Level)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("notification cancelled: %w", ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/notify", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Computer-Management-API/1.0")

		resp, err := c.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			c.logger.Printf("Notification sent successfully for employee %s (attempt %d)",
				notification.EmployeeAbbreviation, attempt)
			return nil
		}

		if resp != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: request failed", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt < maxRetries {
			// Exponential backoff with jitter
			delay := baseDelay * time.Duration(1<<(attempt-1))
			jitter := time.Duration(attempt*100) * time.Millisecond
			totalDelay := delay + jitter

			c.logger.Printf("Notification attempt %d failed for employee %s, retrying in %v: %v",
				attempt, notification.EmployeeAbbreviation, totalDelay, lastErr)

			select {
			case <-ctx.Done():
				return fmt.Errorf("notification cancelled during retry: %w", ctx.Err())
			case <-time.After(totalDelay):
			}
		}
	}

	c.logger.Printf("Notification failed after %d attempts for employee %s: %v",
		maxRetries, notification.EmployeeAbbreviation, lastErr)
	return fmt.Errorf("notification failed after %d attempts: %w", maxRetries, lastErr)
}
