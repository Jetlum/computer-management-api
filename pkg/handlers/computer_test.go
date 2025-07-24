package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"greenbone-case-study/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// Mock service for testing
type mockComputerService struct {
	computers map[uint]*models.Computer
	nextID    uint
}

func newMockService() *mockComputerService {
	return &mockComputerService{
		computers: make(map[uint]*models.Computer),
		nextID:    1,
	}
}

func (m *mockComputerService) CreateComputer(computer *models.Computer) error {
	computer.ID = m.nextID
	m.nextID++
	m.computers[computer.ID] = computer
	return nil
}

func (m *mockComputerService) GetAllComputers() ([]models.Computer, error) {
	var result []models.Computer
	for _, computer := range m.computers {
		result = append(result, *computer)
	}
	return result, nil
}

func (m *mockComputerService) GetComputerByID(id uint) (*models.Computer, error) {
	computer, exists := m.computers[id]
	if !exists {
		return nil, errors.New("computer not found")
	}
	return computer, nil
}

func (m *mockComputerService) GetComputersByEmployee(abbr string) ([]models.Computer, error) {
	var result []models.Computer
	for _, computer := range m.computers {
		if computer.EmployeeAbbreviation != nil && *computer.EmployeeAbbreviation == abbr {
			result = append(result, *computer)
		}
	}
	return result, nil
}

func (m *mockComputerService) UpdateComputer(computer *models.Computer) error {
	if _, exists := m.computers[computer.ID]; !exists {
		return errors.New("computer not found")
	}
	m.computers[computer.ID] = computer
	return nil
}

func (m *mockComputerService) DeleteComputer(id uint) error {
	if _, exists := m.computers[id]; !exists {
		return errors.New("computer not found")
	}
	delete(m.computers, id)
	return nil
}

func TestCreateComputer(t *testing.T) {
	service := newMockService()
	handler := NewComputerHandler(service)

	computer := models.Computer{
		MACAddress:   "00:11:22:33:44:55",
		ComputerName: "Test Computer",
		IPAddress:    "192.168.1.100",
	}

	body, _ := json.Marshal(computer)
	req := httptest.NewRequest("POST", "/api/computers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateComputer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.Computer
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ID == 0 {
		t.Error("Expected computer ID to be set")
	}
}

func TestGetAllComputers(t *testing.T) {
	service := newMockService()
	handler := NewComputerHandler(service)

	// Add a test computer
	computer := &models.Computer{
		MACAddress:   "00:11:22:33:44:55",
		ComputerName: "Test Computer",
		IPAddress:    "192.168.1.100",
	}
	service.CreateComputer(computer)

	req := httptest.NewRequest("GET", "/api/computers", nil)
	w := httptest.NewRecorder()

	handler.GetAllComputers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var computers []models.Computer
	json.Unmarshal(w.Body.Bytes(), &computers)

	if len(computers) != 1 {
		t.Errorf("Expected 1 computer, got %d", len(computers))
	}
}

func TestGetComputerByID(t *testing.T) {
	service := newMockService()
	handler := NewComputerHandler(service)

	// Add a test computer
	computer := &models.Computer{
		MACAddress:   "00:11:22:33:44:55",
		ComputerName: "Test Computer",
		IPAddress:    "192.168.1.100",
	}
	service.CreateComputer(computer)

	req := httptest.NewRequest("GET", "/api/computers/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	handler.GetComputerByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetComputerByIDNotFound(t *testing.T) {
	service := newMockService()
	handler := NewComputerHandler(service)

	req := httptest.NewRequest("GET", "/api/computers/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	handler.GetComputerByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
