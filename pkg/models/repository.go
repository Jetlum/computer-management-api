package models

import (
	"gorm.io/gorm"
)

type computerRepository struct {
	db *gorm.DB
}

// NewComputerRepository creates a new computer repository
func NewComputerRepository(db *gorm.DB) ComputerRepository {
	return &computerRepository{db: db}
}

// Create adds a new computer to the database
func (r *computerRepository) Create(computer *Computer) error {
	return r.db.Create(computer).Error
}

// GetAll retrieves all computers
func (r *computerRepository) GetAll() ([]Computer, error) {
	var computers []Computer
	err := r.db.Find(&computers).Error
	return computers, err
}

// GetByID retrieves a computer by ID
func (r *computerRepository) GetByID(id uint) (*Computer, error) {
	var computer Computer
	err := r.db.First(&computer, id).Error
	if err != nil {
		return nil, err
	}
	return &computer, nil
}

// GetByEmployeeAbbreviation retrieves computers by employee abbreviation
func (r *computerRepository) GetByEmployeeAbbreviation(abbr string) ([]Computer, error) {
	var computers []Computer
	err := r.db.Where("employee_abbreviation = ?", abbr).Find(&computers).Error
	return computers, err
}

// Update updates a computer
func (r *computerRepository) Update(computer *Computer) error {
	return r.db.Save(computer).Error
}

// Delete removes a computer by ID
func (r *computerRepository) Delete(id uint) error {
	return r.db.Delete(&Computer{}, id).Error
}

// CountByEmployee counts computers assigned to an employee
func (r *computerRepository) CountByEmployee(abbr string) (int64, error) {
	var count int64
	err := r.db.Model(&Computer{}).Where("employee_abbreviation = ?", abbr).Count(&count).Error
	return count, err
}
