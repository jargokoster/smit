package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smit/server/api/models"
	"smit/server/api/storage"
)

const AppVersion = "1.0.0"

// Handler holds the storage dependency
type Handler struct {
	storage storage.Storage
}

// Create a new handler instance
func NewHandler(storage storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// Send error responses
func (h *Handler) sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.ErrorResponse{
		Error:     message,
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Send JSON responses
func (h *Handler) sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to encode response")
	}
}

// Extract ID from URL path
func extractIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 4 {
		return 0, errors.New("invalid path")
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, errors.New("invalid ID format")
	}

	if id < 1 || id > 4094 {
		return 0, errors.New("ID must be between 1 and 4094")
	}

	return id, nil
}

// Handles GET /api/v1/vlans
func (h *Handler) GetVLANs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	vlans, err := h.storage.GetAll()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve VLANs")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, vlans)
}

// Handles POST /api/v1/vlans
func (h *Handler) CreateVLAN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input models.VLANInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if err := input.Validate(); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create VLAN
	vlan, err := h.storage.Create(&input)
	if err != nil {
		if errors.Is(err, storage.ErrVLANExists) {
			h.sendErrorResponse(w, http.StatusConflict, "VLAN with this ID already exists")
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create VLAN")
		return
	}

	h.sendJSONResponse(w, http.StatusCreated, vlan)
}

// Handles GET /api/v1/vlans/{id}
func (h *Handler) GetVLAN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	vlan, err := h.storage.GetByID(id)
	if err != nil {
		if errors.Is(err, storage.ErrVLANNotFound) {
			h.sendErrorResponse(w, http.StatusNotFound, "VLAN not found")
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve VLAN")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, vlan)
}

// Handles PUT /api/v1/vlans/{id}
func (h *Handler) UpdateVLAN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var input models.VLANInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if err := input.Validate(); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update VLAN
	vlan, err := h.storage.Update(id, &input)
	if err != nil {
		if errors.Is(err, storage.ErrVLANNotFound) {
			h.sendErrorResponse(w, http.StatusNotFound, "VLAN not found")
			return
		}
		if errors.Is(err, storage.ErrVLANExists) {
			h.sendErrorResponse(w, http.StatusConflict, "VLAN with this ID already exists")
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to update VLAN")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, vlan)
}

// Handles DELETE /api/v1/vlans/{id}
func (h *Handler) DeleteVLAN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.storage.Delete(id)
	if err != nil {
		if errors.Is(err, storage.ErrVLANNotFound) {
			h.sendErrorResponse(w, http.StatusNotFound, "VLAN not found")
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete VLAN")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   AppVersion,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// Handler for VLAN endpoints
func (h *Handler) VLANHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle /api/v1/vlans
	if path == "/api/v1/vlans" {
		switch r.Method {
		case http.MethodGet:
			h.GetVLANs(w, r)
		case http.MethodPost:
			h.CreateVLAN(w, r)
		default:
			h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// Handle /api/v1/vlans/{id}
	if strings.HasPrefix(path, "/api/v1/vlans/") {
		switch r.Method {
		case http.MethodGet:
			h.GetVLAN(w, r)
		case http.MethodPut:
			h.UpdateVLAN(w, r)
		case http.MethodDelete:
			h.DeleteVLAN(w, r)
		default:
			h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	h.sendErrorResponse(w, http.StatusNotFound, "Endpoint not found")
}
