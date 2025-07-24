package services

import (
	"errors"
	"fmt"
	"greenbone-case-study/pkg/models"
	"greenbone-case-study/pkg/notifications"
	"strings"
)

type computerService struct {
	repo         models.ComputerRepository
	notifyClient notifications.NotificationClient
}

// NewComputerService creates a new computer service
func NewComputerService(repo models.ComputerRepository, notifyClient notifications.NotificationClient) models.ComputerService {
	return &computerService{
		repo:         repo,
		notifyClient: notifyClient,
	}
}

// CreateComputer creates a new computer with validation
func (s *computerService) CreateComputer(computer *models.Computer) error {
	// Validate input
	if err := s.validateComputer(computer); err != nil {
		return err
	}

	// Check if employee already has computers and count them
	var currentCount int64 = 0
	if computer.EmployeeAbbreviation != nil && *computer.EmployeeAbbreviation != "" {
		count, err := s.repo.CountByEmployee(*computer.EmployeeAbbreviation)
		if err != nil {
			return fmt.Errorf("failed to count employee computers: %w", err)
		}
		currentCount = count
	}

	// Create the computer
	if err := s.repo.Create(computer); err != nil {
		return fmt.Errorf("failed to create computer: %w", err)
	}

	// Check if employee now has 3 or more computers and send notification
	if computer.EmployeeAbbreviation != nil && *computer.EmployeeAbbreviation != "" {
		newCount := currentCount + 1
		if newCount >= 3 {
			go s.sendComputerLimitNotification(*computer.EmployeeAbbreviation, int(newCount))
		}
	}

	return nil
}

// GetAllComputers retrieves all computers
func (s *computerService) GetAllComputers() ([]models.Computer, error) {
	computers, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get computers: %w", err)
	}
	return computers, nil
}

// GetComputerByID retrieves a computer by ID
func (s *computerService) GetComputerByID(id uint) (*models.Computer, error) {
	if id == 0 {
		return nil, errors.New("invalid computer ID")
	}

	computer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get computer: %w", err)
	}
	return computer, nil
}

// GetComputersByEmployee retrieves computers by employee abbreviation
func (s *computerService) GetComputersByEmployee(abbr string) ([]models.Computer, error) {
	if err := s.validateEmployeeAbbreviation(abbr); err != nil {
		return nil, err
	}

	computers, err := s.repo.GetByEmployeeAbbreviation(abbr)
	if err != nil {
		return nil, fmt.Errorf("failed to get computers for employee %s: %w", abbr, err)
	}
	return computers, nil
}

// UpdateComputer updates a computer with validation
func (s *computerService) UpdateComputer(computer *models.Computer) error {
	if computer.ID == 0 {
		return errors.New("invalid computer ID")
	}

	// Get existing computer to check for employee changes
	existingComputer, err := s.repo.GetByID(computer.ID)
	if err != nil {
		return fmt.Errorf("computer not found: %w", err)
	}

	// Validate input
	if err := s.validateComputer(computer); err != nil {
		return err
	}

	// Check if employee assignment changed
	oldEmployee := ""
	newEmployee := ""

	if existingComputer.EmployeeAbbreviation != nil {
		oldEmployee = *existingComputer.EmployeeAbbreviation
	}
	if computer.EmployeeAbbreviation != nil {
		newEmployee = *computer.EmployeeAbbreviation
	}

	// Update the computer
	if err := s.repo.Update(computer); err != nil {
		return fmt.Errorf("failed to update computer: %w", err)
	}

	// If employee changed, check new employee's computer count
	if oldEmployee != newEmployee && newEmployee != "" {
		count, err := s.repo.CountByEmployee(newEmployee)
		if err != nil {
			// Log error but don't fail the update
			fmt.Printf("Warning: failed to count computers for employee %s: %v\n", newEmployee, err)
		} else if count >= 3 {
			go s.sendComputerLimitNotification(newEmployee, int(count))
		}
	}

	return nil
}

// DeleteComputer deletes a computer by ID
func (s *computerService) DeleteComputer(id uint) error {
	if id == 0 {
		return errors.New("invalid computer ID")
	}

	// Check if computer exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("computer not found: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete computer: %w", err)
	}

	return nil
}

// validateComputer validates computer input data
func (s *computerService) validateComputer(computer *models.Computer) error {
	if computer.MACAddress == "" {
		return errors.New("MAC address is required")
	}
	if computer.ComputerName == "" {
		return errors.New("computer name is required")
	}
	if computer.IPAddress == "" {
		return errors.New("IP address is required")
	}

	// Validate MAC address format (basic validation)
	if len(computer.MACAddress) != 17 {
		return errors.New("MAC address must be 17 characters long (XX:XX:XX:XX:XX:XX)")
	}

	// Validate employee abbreviation if provided
	if computer.EmployeeAbbreviation != nil && *computer.EmployeeAbbreviation != "" {
		if err := s.validateEmployeeAbbreviation(*computer.EmployeeAbbreviation); err != nil {
			return err
		}
	}

	return nil
}

// validateEmployeeAbbreviation validates employee abbreviation
func (s *computerService) validateEmployeeAbbreviation(abbr string) error {
	if len(abbr) != 3 {
		return errors.New("employee abbreviation must be exactly 3 characters")
	}
	if abbr != strings.ToLower(abbr) {
		return errors.New("employee abbreviation must be lowercase")
	}
	return nil
}

// sendComputerLimitNotification sends a notification when employee has 3+ computers
func (s *computerService) sendComputerLimitNotification(employeeAbbr string, count int) {
	notification := notifications.Notification{
		Level:                "warning",
		EmployeeAbbreviation: employeeAbbr,
		Message:              fmt.Sprintf("Employee %s has been assigned %d computers", employeeAbbr, count),
	}

	if err := s.notifyClient.SendNotification(notification); err != nil {
		fmt.Printf("Failed to send notification for employee %s: %v\n", employeeAbbr, err)
	}
}
