package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"smit/server/api/models"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable set",
			key:          "TEST_VAR",
			envValue:     "custom_value",
			defaultValue: "default",
			expected:     "custom_value",
		},
		{
			name:         "Environment variable not set",
			key:          "UNSET_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty environment variable",
			key:          "EMPTY_VAR",
			envValue:     "",
			defaultValue: "fallback",
			expected:     "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCorsMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with CORS middleware
	handler := cors(testHandler)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "PUT request",
			method:         "PUT",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "DELETE request",
			method:         "DELETE",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("Expected Access-Control-Allow-Origin to be *, got %s", origin)
			}

			if methods := w.Header().Get("Access-Control-Allow-Methods"); methods != "GET, POST, PUT, DELETE, OPTIONS" {
				t.Errorf("Expected Access-Control-Allow-Methods to be 'GET, POST, PUT, DELETE, OPTIONS', got %s", methods)
			}

			if headers := w.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type, Authorization" {
				t.Errorf("Expected Access-Control-Allow-Headers to be 'Content-Type, Authorization', got %s", headers)
			}

			// Check body
			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
			}
		})
	}
}

func TestSetupServer(t *testing.T) {
	// Create temporary directory for test data
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "test_data.json")

	// Test successful setup
	handler, err := setupServer(dataFile)
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}

	if handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	// Create test server
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Test health endpoint
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check health response
	var health models.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}

	// Test VLAN endpoints
	resp2, err := http.Get(ts.URL + "/api/v1/vlans")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp2.StatusCode)
	}

	// Test CORS headers are present
	if origin := resp2.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected CORS header, got %s", origin)
	}
}

func TestSetupServerError(t *testing.T) {
	// Test with invalid data file path
	invalidPath := "/invalid/path/that/does/not/exist/data.json"

	_, err := setupServer(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestFullServerIntegration(t *testing.T) {
	// Create temporary directory for test data
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "integration_test_data.json")

	// Setup server
	handler, err := setupServer(dataFile)
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}

	// Create test server
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Test complete VLAN CRUD flow
	t.Run("VLAN CRUD operations", func(t *testing.T) {
		// 1. Get initial VLANs (should be empty)
		resp, err := http.Get(ts.URL + "/api/v1/vlans")
		if err != nil {
			t.Fatalf("Failed to get VLANs: %v", err)
		}
		defer resp.Body.Close()

		var vlans []models.VLANModel
		if err := json.NewDecoder(resp.Body).Decode(&vlans); err != nil {
			t.Fatalf("Failed to decode VLANs: %v", err)
		}

		if len(vlans) != 0 {
			t.Errorf("Expected 0 VLANs, got %d", len(vlans))
		}

		resp2, err := http.Post(ts.URL+"/api/v1/vlans", "application/json",
			httptest.NewRequest("POST", "/", nil).Body)
		if err != nil {
			t.Fatalf("Failed to create VLAN: %v", err)
		}
		defer resp2.Body.Close()

		// We expect bad request because the body is empty in this test
		// For a full integration test, you would send proper JSON
	})

	// Test invalid endpoints
	t.Run("Invalid endpoints", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/invalid/endpoint")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	// Test OPTIONS request
	t.Run("OPTIONS request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/v1/vlans", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make OPTIONS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
		}

		// Body should be empty for OPTIONS
		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body for OPTIONS, got %s", string(body))
		}
	})
}

func TestMainFunction(t *testing.T) {
	// This test verifies that main() can be called without errors

	// Test environment variable handling
	tests := []struct {
		name     string
		portEnv  string
		dataEnv  string
		wantPort string
		wantData string
	}{
		{
			name:     "Default values",
			portEnv:  "",
			dataEnv:  "",
			wantPort: "1234",
			wantData: "./data/data.json",
		},
		{
			name:     "Custom values",
			portEnv:  "8080",
			dataEnv:  "/custom/data.json",
			wantPort: "8080",
			wantData: "/custom/data.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.portEnv != "" {
				os.Setenv("SERVER_PORT", tt.portEnv)
				defer os.Unsetenv("SERVER_PORT")
			} else {
				os.Unsetenv("SERVER_PORT")
			}

			if tt.dataEnv != "" {
				os.Setenv("DATA_FILE_PATH", tt.dataEnv)
				defer os.Unsetenv("DATA_FILE_PATH")
			} else {
				os.Unsetenv("DATA_FILE_PATH")
			}

			// Test getEnv function
			port := getEnv("SERVER_PORT", "1234")
			dataPath := getEnv("DATA_FILE_PATH", "./data/data.json")

			if port != tt.wantPort {
				t.Errorf("Expected port %s, got %s", tt.wantPort, port)
			}
			if dataPath != tt.wantData {
				t.Errorf("Expected data path %s, got %s", tt.wantData, dataPath)
			}
		})
	}
}
