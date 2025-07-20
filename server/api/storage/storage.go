package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"smit/server/api/models"
	"sync"
	"time"
)

var (
	ErrVLANNotFound = errors.New("VLAN not found")
	ErrVLANExists   = errors.New("VLAN already exists")
)

type Storage interface {
	GetAll() ([]models.VLANModel, error)
	GetByID(id int) (*models.VLANModel, error)
	Create(vlan *models.VLANInput) (*models.VLANModel, error)
	Update(id int, vlan *models.VLANInput) (*models.VLANModel, error)
	Delete(id int) error
}

type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
}

// New JSON storage instance
func NewJSONStorage(filePath string) (*JSONStorage, error) {
	storage := &JSONStorage{
		filePath: filePath,
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		initialData := models.VLANData{VLANs: []models.VLANModel{}}
		if err := storage.saveData(&initialData); err != nil {
			return nil, fmt.Errorf("failed to create initial data file: %w", err)
		}
	}

	return storage, nil
}

// Load data from JSON file
func (s *JSONStorage) loadData() (*models.VLANData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var vlanData models.VLANData
	if err := json.Unmarshal(data, &vlanData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &vlanData, nil
}

// Save data to JSON file
func (s *JSONStorage) saveData(data *models.VLANData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(s.filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Get all VLANs
func (s *JSONStorage) GetAll() ([]models.VLANModel, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	return data.VLANs, nil
}

// Get VLAN by ID
func (s *JSONStorage) GetByID(id int) (*models.VLANModel, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, vlan := range data.VLANs {
		if vlan.ID == id {
			return &vlan, nil
		}
	}

	return nil, ErrVLANNotFound
}

// Create new VLAN
func (s *JSONStorage) Create(input *models.VLANInput) (*models.VLANModel, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	// Check if VLAN ID already exists
	for _, vlan := range data.VLANs {
		if vlan.VlanID == input.VlanID {
			return nil, ErrVLANExists
		}
	}

	// Generate new ID
	newID := 1
	if len(data.VLANs) > 0 {
		maxID := 0
		for _, vlan := range data.VLANs {
			if vlan.ID > maxID {
				maxID = vlan.ID
			}
		}
		newID = maxID + 1
	}

	// Create new VLAN
	now := time.Now()
	newVLAN := models.VLANModel{
		ID:        newID,
		Name:      input.Name,
		VlanID:    input.VlanID,
		Subnet:    input.Subnet,
		Gateway:   input.Gateway,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data.VLANs = append(data.VLANs, newVLAN)

	if err := s.saveData(data); err != nil {
		return nil, err
	}

	return &newVLAN, nil
}

// Update existing VLAN
func (s *JSONStorage) Update(id int, input *models.VLANInput) (*models.VLANModel, error) {
	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	// Find VLAN to update
	for i, vlan := range data.VLANs {
		if vlan.ID == id {
			// Check if new VLAN ID conflicts with another VLAN
			if vlan.VlanID != input.VlanID {
				for _, v := range data.VLANs {
					if v.ID != id && v.VlanID == input.VlanID {
						return nil, ErrVLANExists
					}
				}
			}

			// Update VLAN
			data.VLANs[i].Name = input.Name
			data.VLANs[i].VlanID = input.VlanID
			data.VLANs[i].Subnet = input.Subnet
			data.VLANs[i].Gateway = input.Gateway
			data.VLANs[i].Status = input.Status
			data.VLANs[i].UpdatedAt = time.Now()

			if err := s.saveData(data); err != nil {
				return nil, err
			}

			return &data.VLANs[i], nil
		}
	}

	return nil, ErrVLANNotFound
}

// Delete VLAN
func (s *JSONStorage) Delete(id int) error {
	data, err := s.loadData()
	if err != nil {
		return err
	}

	// Find and remove VLAN
	found := false
	newVLANs := make([]models.VLANModel, 0, len(data.VLANs)-1)
	for _, vlan := range data.VLANs {
		if vlan.ID == id {
			found = true
			continue
		}
		newVLANs = append(newVLANs, vlan)
	}

	if !found {
		return ErrVLANNotFound
	}

	data.VLANs = newVLANs
	return s.saveData(data)
}
