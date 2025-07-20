package storage

import (
	"os"
	"path/filepath"
	"testing"

	"smit/server/api/models"
)

func TestJSONStorage(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test_data.json")

	// Test creating new storage
	store, err := NewJSONStorage(testFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test GetAll on empty storage
	vlans, err := store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all VLANs: %v", err)
	}
	if len(vlans) != 0 {
		t.Errorf("Expected 0 VLANs, got %d", len(vlans))
	}

	// Test Create
	input := &models.VLANInput{
		Name:    "Test VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	}

	created, err := store.Create(input)
	if err != nil {
		t.Fatalf("Failed to create VLAN: %v", err)
	}

	if created.ID != 1 {
		t.Errorf("Expected ID 1, got %d", created.ID)
	}
	if created.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, created.Name)
	}

	// Test GetByID
	retrieved, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get VLAN by ID: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	// Test GetByID with non-existent ID
	_, err = store.GetByID(999)
	if err != ErrVLANNotFound {
		t.Errorf("Expected ErrVLANNotFound, got %v", err)
	}

	// Test duplicate VLAN ID
	_, err = store.Create(input)
	if err != ErrVLANExists {
		t.Errorf("Expected ErrVLANExists, got %v", err)
	}

	// Test Update
	updateInput := &models.VLANInput{
		Name:    "Updated VLAN",
		VlanID:  100,
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "maintenance",
	}

	updated, err := store.Update(created.ID, updateInput)
	if err != nil {
		t.Fatalf("Failed to update VLAN: %v", err)
	}
	if updated.Name != updateInput.Name {
		t.Errorf("Expected name %s, got %s", updateInput.Name, updated.Name)
	}
	if updated.Status != updateInput.Status {
		t.Errorf("Expected status %s, got %s", updateInput.Status, updated.Status)
	}

	// Test Update with non-existent ID
	_, err = store.Update(999, updateInput)
	if err != ErrVLANNotFound {
		t.Errorf("Expected ErrVLANNotFound, got %v", err)
	}

	// Test Create second VLAN
	input2 := &models.VLANInput{
		Name:    "Second VLAN",
		VlanID:  200,
		Subnet:  "192.168.200.0/24",
		Gateway: "192.168.200.1",
		Status:  "active",
	}

	created2, err := store.Create(input2)
	if err != nil {
		t.Fatalf("Failed to create second VLAN: %v", err)
	}

	if created2.VlanID != 200 {
		t.Errorf("Expected VLAN ID 200, got %d", created2.VlanID)
	}

	// Test GetAll with multiple VLANs
	vlans, err = store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all VLANs: %v", err)
	}
	if len(vlans) != 2 {
		t.Errorf("Expected 2 VLANs, got %d", len(vlans))
	}

	// Test Update with conflicting VLAN ID
	conflictInput := &models.VLANInput{
		Name:    "Conflict VLAN",
		VlanID:  200, // Same as created2
		Subnet:  "192.168.100.0/24",
		Gateway: "192.168.100.1",
		Status:  "active",
	}

	_, err = store.Update(created.ID, conflictInput)
	if err != ErrVLANExists {
		t.Errorf("Expected ErrVLANExists, got %v", err)
	}

	// Test Delete
	err = store.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete VLAN: %v", err)
	}

	// Verify deletion
	_, err = store.GetByID(created.ID)
	if err != ErrVLANNotFound {
		t.Errorf("Expected ErrVLANNotFound after deletion, got %v", err)
	}

	// Test Delete with non-existent ID
	err = store.Delete(999)
	if err != ErrVLANNotFound {
		t.Errorf("Expected ErrVLANNotFound, got %v", err)
	}

	// Verify remaining VLANs
	vlans, err = store.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all VLANs: %v", err)
	}
	if len(vlans) != 1 {
		t.Errorf("Expected 1 VLAN after deletion, got %d", len(vlans))
	}
}

func TestJSONStoragePersistence(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "storage_persist_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "persist_test.json")

	// Create storage and add data
	store1, err := NewJSONStorage(testFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	input := &models.VLANInput{
		Name:    "Persistent VLAN",
		VlanID:  300,
		Subnet:  "192.168.300.0/24",
		Gateway: "192.168.300.1",
		Status:  "active",
	}

	created, err := store1.Create(input)
	if err != nil {
		t.Fatalf("Failed to create VLAN: %v", err)
	}

	// Create new storage instance with same file
	store2, err := NewJSONStorage(testFile)
	if err != nil {
		t.Fatalf("Failed to create second storage instance: %v", err)
	}

	// Verify data persisted
	vlans, err := store2.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all VLANs from second instance: %v", err)
	}

	if len(vlans) != 1 {
		t.Errorf("Expected 1 VLAN, got %d", len(vlans))
	}

	if vlans[0].ID != created.ID || vlans[0].VlanID != created.VlanID {
		t.Errorf("Data not persisted correctly")
	}
}

func TestJSONStorageInvalidFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "storage_invalid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with invalid JSON file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	store, err := NewJSONStorage(invalidFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	_, err = store.GetAll()
	if err == nil {
		t.Error("Expected error when reading invalid JSON, got nil")
	}
}

func TestJSONStorageEdgeCases(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "storage_edge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "edge_test.json")
	store, err := NewJSONStorage(testFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test creating multiple VLANs and checking ID generation
	for i := 1; i <= 5; i++ {
		input := &models.VLANInput{
			Name:    "VLAN",
			VlanID:  i * 100,
			Subnet:  "192.168.1.0/24",
			Gateway: "192.168.1.1",
			Status:  "active",
		}
		created, err := store.Create(input)
		if err != nil {
			t.Fatalf("Failed to create VLAN %d: %v", i, err)
		}
		if created.ID != i {
			t.Errorf("Expected ID %d, got %d", i, created.ID)
		}
	}

	// Delete a VLAN in the middle
	err = store.Delete(3)
	if err != nil {
		t.Fatalf("Failed to delete VLAN: %v", err)
	}

	// Create new VLAN and check ID is properly incremented
	input := &models.VLANInput{
		Name:    "New VLAN",
		VlanID:  600,
		Subnet:  "192.168.6.0/24",
		Gateway: "192.168.6.1",
		Status:  "active",
	}
	created, err := store.Create(input)
	if err != nil {
		t.Fatalf("Failed to create new VLAN: %v", err)
	}
	if created.ID != 6 {
		t.Errorf("Expected ID 6, got %d", created.ID)
	}
}
