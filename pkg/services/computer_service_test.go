package services

import (
	"context"
	"errors"
	"fmt"
	"greenbone-case-study/pkg/models"
	"greenbone-case-study/pkg/notifications"
	"testing"
	"time"
)

// Mock repository for testing
type mockComputerRepository struct {
	computers map[uint]*models.Computer
	nextID    uint
	countMap  map[string]int64
}

func newMockRepository() *mockComputerRepository {
	return &mockComputerRepository{
		computers: make(map[uint]*models.Computer),
		nextID:    1,
		countMap:  make(map[string]int64),
	}
}

func (m *mockComputerRepository) Create(computer *models.Computer) error {
	computer.ID = m.nextID
	m.nextID++
	m.computers[computer.ID] = computer

	if computer.EmployeeAbbreviation != nil {
		m.countMap[*computer.EmployeeAbbreviation]++
	}

	return nil
}

func (m *mockComputerRepository) GetAll() ([]models.Computer, error) {
	var result []models.Computer
	for _, computer := range m.computers {
		result = append(result, *computer)
	}
	return result, nil
}

func (m *mockComputerRepository) GetByID(id uint) (*models.Computer, error) {
	computer, exists := m.computers[id]
	if !exists {
		return nil, errors.New("computer not found")
	}
	return computer, nil
}

func (m *mockComputerRepository) GetByEmployeeAbbreviation(abbr string) ([]models.Computer, error) {
	var result []models.Computer
	for _, computer := range m.computers {
		if computer.EmployeeAbbreviation != nil && *computer.EmployeeAbbreviation == abbr {
			result = append(result, *computer)
		}
	}
	return result, nil
}

func (m *mockComputerRepository) Update(computer *models.Computer) error {
	if _, exists := m.computers[computer.ID]; !exists {
		return errors.New("computer not found")
	}
	m.computers[computer.ID] = computer
	return nil
}

func (m *mockComputerRepository) Delete(id uint) error {
	if _, exists := m.computers[id]; !exists {
		return errors.New("computer not found")
	}
	delete(m.computers, id)
	return nil
}

func (m *mockComputerRepository) CountByEmployee(abbr string) (int64, error) {
	return m.countMap[abbr], nil
}

// Mock notification client for testing - FIXED
type mockNotificationClient struct {
	notifications []notifications.Notification
	shouldFail    bool
}

func (m *mockNotificationClient) SendNotification(notification notifications.Notification) error {
	return m.SendNotificationWithContext(context.Background(), notification)
}

func (m *mockNotificationClient) SendNotificationWithContext(ctx context.Context, notification notifications.Notification) error {
	if m.shouldFail {
		return errors.New("mock notification failed")
	}
	m.notifications = append(m.notifications, notification)
	return nil
}

func TestCreateComputer(t *testing.T) {
	repo := newMockRepository()
	notifyClient := &mockNotificationClient{}
	service := NewComputerService(repo, notifyClient)

	abbr := "abc"
	computer := &models.Computer{
		MACAddress:           "00:11:22:33:44:55",
		ComputerName:         "Test Computer",
		IPAddress:            "192.168.1.100",
		EmployeeAbbreviation: &abbr,
		Description:          "Test description",
	}

	err := service.CreateComputer(computer)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if computer.ID == 0 {
		t.Error("Expected computer ID to be set")
	}
}

func TestCreateComputerValidation(t *testing.T) {
	repo := newMockRepository()
	notifyClient := &mockNotificationClient{}
	service := NewComputerService(repo, notifyClient)

	tests := []struct {
		name     string
		computer *models.Computer
		wantErr  bool
	}{
		{
			name: "missing MAC address",
			computer: &models.Computer{
				ComputerName: "Test",
				IPAddress:    "192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "missing computer name",
			computer: &models.Computer{
				MACAddress: "00:11:22:33:44:55",
				IPAddress:  "192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "invalid employee abbreviation length",
			computer: &models.Computer{
				MACAddress:           "00:11:22:33:44:55",
				ComputerName:         "Test",
				IPAddress:            "192.168.1.1",
				EmployeeAbbreviation: func() *string { s := "ab"; return &s }(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateComputer(tt.computer)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComputer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateComputerNotificationTrigger(t *testing.T) {
	repo := newMockRepository()
	notifyClient := &mockNotificationClient{}
	service := NewComputerService(repo, notifyClient)

	abbr := "abc"

	// Create 3 computers for the same employee
	for i := 1; i <= 3; i++ {
		computer := &models.Computer{
			MACAddress:           fmt.Sprintf("00:11:22:33:44:%02d", i),
			ComputerName:         fmt.Sprintf("Test Computer %d", i),
			IPAddress:            fmt.Sprintf("192.168.1.%d", i),
			EmployeeAbbreviation: &abbr,
			Description:          "Test description",
		}

		err := service.CreateComputer(computer)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	}

	// Wait a bit for the goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Check that notification was sent when the 3rd computer was added
	if len(notifyClient.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifyClient.notifications))
	}

	if len(notifyClient.notifications) > 0 {
		notification := notifyClient.notifications[0]
		if notification.Level != "warning" {
			t.Errorf("Expected warning level, got %s", notification.Level)
		}
		if notification.EmployeeAbbreviation != abbr {
			t.Errorf("Expected employee %s, got %s", abbr, notification.EmployeeAbbreviation)
		}
	}
}
