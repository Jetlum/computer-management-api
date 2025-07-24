package handlers

import (
	"encoding/json"
	"greenbone-case-study/pkg/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ComputerHandler handles HTTP requests for computers
type ComputerHandler struct {
	service models.ComputerService
}

// NewComputerHandler creates a new computer handler
func NewComputerHandler(service models.ComputerService) *ComputerHandler {
	return &ComputerHandler{
		service: service,
	}
}

// CreateComputer handles POST /computers
func (h *ComputerHandler) CreateComputer(w http.ResponseWriter, r *http.Request) {
	var computer models.Computer

	if err := json.NewDecoder(r.Body).Decode(&computer); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if err := h.service.CreateComputer(&computer); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, computer)
}

// GetAllComputers handles GET /computers
func (h *ComputerHandler) GetAllComputers(w http.ResponseWriter, r *http.Request) {
	computers, err := h.service.GetAllComputers()
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve computers")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, computers)
}

// GetComputerByID handles GET /computers/{id}
func (h *ComputerHandler) GetComputerByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid computer ID")
		return
	}

	computer, err := h.service.GetComputerByID(uint(id))
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Computer not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, computer)
}

// GetComputersByEmployee handles GET /employees/{abbr}/computers
func (h *ComputerHandler) GetComputersByEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	abbr := vars["abbr"]

	computers, err := h.service.GetComputersByEmployee(abbr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, computers)
}

// UpdateComputer handles PUT /computers/{id}
func (h *ComputerHandler) UpdateComputer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid computer ID")
		return
	}

	var computer models.Computer
	if err := json.NewDecoder(r.Body).Decode(&computer); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	computer.ID = uint(id)

	if err := h.service.UpdateComputer(&computer); err != nil {
		if err.Error() == "computer not found" {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, computer)
}

// DeleteComputer handles DELETE /computers/{id}
func (h *ComputerHandler) DeleteComputer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid computer ID")
		return
	}

	if err := h.service.DeleteComputer(uint(id)); err != nil {
		if err.Error() == "computer not found" {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete computer")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Computer deleted successfully",
	})
}

// writeJSONResponse writes a JSON response
func (h *ComputerHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *ComputerHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
