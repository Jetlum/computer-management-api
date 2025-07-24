package models

import (
	"time"
)

// Computer represents a company-issued computer
type Computer struct {
	ID                   uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MACAddress           string    `json:"mac_address" gorm:"not null;unique;size:17" validate:"required"`
	ComputerName         string    `json:"computer_name" gorm:"not null;size:100" validate:"required"`
	IPAddress            string    `json:"ip_address" gorm:"not null;size:15" validate:"required"`
	EmployeeAbbreviation *string   `json:"employee_abbreviation,omitempty" gorm:"size:3"`
	Description          string    `json:"description" gorm:"size:500"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ComputerRepository interface for database operations
type ComputerRepository interface {
	Create(computer *Computer) error
	GetAll() ([]Computer, error)
	GetByID(id uint) (*Computer, error)
	GetByEmployeeAbbreviation(abbr string) ([]Computer, error)
	Update(computer *Computer) error
	Delete(id uint) error
	CountByEmployee(abbr string) (int64, error)
}

// ComputerService interface for business logic
type ComputerService interface {
	CreateComputer(computer *Computer) error
	GetAllComputers() ([]Computer, error)
	GetComputerByID(id uint) (*Computer, error)
	GetComputersByEmployee(abbr string) ([]Computer, error)
	UpdateComputer(computer *Computer) error
	DeleteComputer(id uint) error
}
