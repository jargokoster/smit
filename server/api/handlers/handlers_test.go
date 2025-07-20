package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"smit/server/api/models"
	"smit/server/api/storage"
)

// MockStorage implements the storage.Storage interface for testing
type MockStorage struct {
	vlans  []models.VLANModel
	nextID int
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		vlans:  []models.VLANModel{},
		nextID: 1,
	}
}

func (m *MockStorage) GetAll() ([]models.VLANModel, error) {
	return m.vlans, nil
}

func (m *MockStorage) GetByID(id int) (*models.VLANModel, error) {
	for _, vlan := range m.vlans {
		if vlan.ID == id {
			return &vlan, nil
		}
	}
	return nil, storage.ErrVLANNotFound
}

func (m *MockStorage) Create(input *models.VLANInput) (*models.VLANModel, error) {
	// Check if VLAN ID already exists
	for _, vlan := range m.vlans {
		if vlan.VlanID == input.VlanID {
			return nil, storage.ErrVLANExists
		}
	}

	vlan := models.VLANModel{
		ID:        m.nextID,
		Name:      input.Name,
		VlanID:    input.VlanID,
		Subnet:    input.Subnet,
		Gateway:   input.Gateway,
		Status:    input.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.vlans = append(m.vlans, vlan)
	m.nextID++

	return &vlan, nil
}

func (m *MockStorage) Update(id int, input *models.VLANInput) (*models.VLANModel, error) {
	for i, vlan := range m.vlans {
		if vlan.ID == id {
			// Check if new VLAN ID conflicts
			if vlan.VlanID != input.VlanID {
				for _, v := range m.vlans {
					if v.ID != id && v.VlanID == input.VlanID {
						return nil, storage.ErrVLANExists
					}
				}
			}

			m.vlans[i].Name = input.Name
			m.vlans[i].VlanID = input.VlanID
			m.vlans[i].Subnet = input.Subnet
			m.vlans[i].Gateway = input.Gateway
			m.vlans[i].Status = input.Status
			m.vlans[i].UpdatedAt = time.Now()

			return &m.vlans[i], nil
		}
	}
	return nil, storage.ErrVLANNotFound
}

func (m *MockStorage) Delete(id int) error {
	for i, vlan := range m.vlans {
		if vlan.ID == id {
			m.vlans = append(m.vlans[:i], m.vlans[i+1:]...)
			return nil
		}
	}
	return storage.ErrVLANNotFound
}

func TestHealthCheck(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var health models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}
}

func TestGetVLANs(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Add test data
	storage.Create(&models.VLANInput{
		Name:    "Test VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	})

	req := httptest.NewRequest("GET", "/api/v1/vlans", nil)
	w := httptest.NewRecorder()

	handler.GetVLANs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var vlans []models.VLANModel
	if err := json.NewDecoder(w.Body).Decode(&vlans); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(vlans) != 1 {
		t.Errorf("Expected 1 VLAN, got %d", len(vlans))
	}
}

func TestCreateVLAN(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	input := models.VLANInput{
		Name:    "New VLAN",
		VlanID:  200,
		Subnet:  "192.168.200.0/24",
		Gateway: "192.168.200.1",
		Status:  "active",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/v1/vlans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateVLAN(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var vlan models.VLANModel
	if err := json.NewDecoder(w.Body).Decode(&vlan); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if vlan.VlanID != 200 {
		t.Errorf("Expected VLAN ID 200, got %d", vlan.VlanID)
	}
}

func TestCreateVLANValidation(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	tests := []struct {
		name   string
		input  models.VLANInput
		status int
	}{
		{
			name: "Invalid VLAN ID",
			input: models.VLANInput{
				Name:    "Test",
				VlanID:  5000,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Invalid subnet",
			input: models.VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "invalid-subnet",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Invalid gateway",
			input: models.VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "invalid-ip",
				Status:  "active",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "Invalid status",
			input: models.VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "invalid",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/v1/vlans", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateVLAN(w, req)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestGetVLAN(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Add test data
	vlan, _ := storage.Create(&models.VLANInput{
		Name:    "Test VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	})

	req := httptest.NewRequest("GET", "/api/v1/vlans/1", nil)
	w := httptest.NewRecorder()

	handler.GetVLAN(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result models.VLANModel
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.ID != vlan.ID {
		t.Errorf("Expected VLAN ID %d, got %d", vlan.ID, result.ID)
	}
}

func TestGetVLANNotFound(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	req := httptest.NewRequest("GET", "/api/v1/vlans/999", nil)
	w := httptest.NewRecorder()

	handler.GetVLAN(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateVLAN(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Add test data
	storage.Create(&models.VLANInput{
		Name:    "Test VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	})

	update := models.VLANInput{
		Name:    "Updated VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "maintenance",
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/api/v1/vlans/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateVLAN(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var vlan models.VLANModel
	if err := json.NewDecoder(w.Body).Decode(&vlan); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if vlan.Name != "Updated VLAN" || vlan.Status != "maintenance" {
		t.Errorf("VLAN not updated correctly")
	}
}

func TestDeleteVLAN(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Add test data
	storage.Create(&models.VLANInput{
		Name:    "Test VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	})

	req := httptest.NewRequest("DELETE", "/api/v1/vlans/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteVLAN(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deletion
	_, err := storage.GetByID(1)
	if err == nil {
		t.Errorf("VLAN was not deleted")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	tests := []struct {
		method   string
		endpoint string
		handler  http.HandlerFunc
	}{
		{"POST", "/health", handler.HealthCheck},
		{"DELETE", "/api/v1/vlans", handler.GetVLANs},
		{"POST", "/api/v1/vlans/1", handler.GetVLAN},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.endpoint, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}

func TestVLANHandler(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"GET vlans", "GET", "/api/v1/vlans", http.StatusOK},
		{"POST vlans invalid body", "POST", "/api/v1/vlans", http.StatusBadRequest},
		{"GET vlan by id not found", "GET", "/api/v1/vlans/999", http.StatusNotFound},
		{"PUT vlan not found", "PUT", "/api/v1/vlans/999", http.StatusBadRequest},
		{"DELETE vlan not found", "DELETE", "/api/v1/vlans/999", http.StatusNotFound},
		{"Invalid path", "GET", "/api/v1/invalid", http.StatusNotFound},
		{"Options preflight", "OPTIONS", "/api/v1/vlans", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == "POST" || tt.method == "PUT" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte("{}")))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()

			handler.VLANHandler(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

func TestExtractIDFromPath(t *testing.T) {
	tests := []struct {
		path    string
		wantID  int
		wantErr bool
	}{
		{"/api/v1/vlans/1", 1, false},
		{"/api/v1/vlans/100", 100, false},
		{"/api/v1/vlans/4094", 4094, false},
		{"/api/v1/vlans/0", 0, true},
		{"/api/v1/vlans/4095", 0, true},
		{"/api/v1/vlans/abc", 0, true},
		{"/api/v1/", 0, true},
		{"/api/v1/vlans/", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			id, err := extractIDFromPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractIDFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if id != tt.wantID {
				t.Errorf("extractIDFromPath() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestCreateVLANConflict(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Create first VLAN
	input := models.VLANInput{
		Name:    "First VLAN",
		VlanID:  300,
		Subnet:  "192.168.30.0/24",
		Gateway: "192.168.30.1",
		Status:  "active",
	}
	storage.Create(&input)

	// Try to create VLAN with same ID
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/v1/vlans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateVLAN(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestUpdateVLANConflict(t *testing.T) {
	storage := NewMockStorage()
	handler := NewHandler(storage)

	// Create two VLANs
	storage.Create(&models.VLANInput{
		Name:    "VLAN 1",
		VlanID:  400,
		Subnet:  "192.168.40.0/24",
		Gateway: "192.168.40.1",
		Status:  "active",
	})
	storage.Create(&models.VLANInput{
		Name:    "VLAN 2",
		VlanID:  500,
		Subnet:  "192.168.50.0/24",
		Gateway: "192.168.50.1",
		Status:  "active",
	})

	// Try to update VLAN 1 with VLAN 2's ID
	update := models.VLANInput{
		Name:    "Updated VLAN",
		VlanID:  500, // Conflict with VLAN 2
		Subnet:  "192.168.40.0/24",
		Gateway: "192.168.40.1",
		Status:  "active",
	}

	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/api/v1/vlans/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateVLAN(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestInvalidJSONBody(t *testing.T) {
	handler := NewHandler(NewMockStorage())

	// Test invalid JSON for POST
	req := httptest.NewRequest("POST", "/api/v1/vlans", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateVLAN(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}

	// Test invalid JSON for PUT
	req = httptest.NewRequest("PUT", "/api/v1/vlans/1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.UpdateVLAN(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}
