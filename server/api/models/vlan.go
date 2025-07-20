package models

import (
	"fmt"
	"net"
	"regexp"
	"time"
)

// VLAN configuration model
type VLANModel struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	VlanID    int       `json:"vlan_id"`
	Subnet    string    `json:"subnet"`
	Gateway   string    `json:"gateway"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Structure for creating/updating a VLAN
type VLANInput struct {
	Name    string `json:"name"`
	VlanID  int    `json:"vlan_id"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	Status  string `json:"status"`
}

// Structure for an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// Structure for health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Structure for JSON data structure
type VLANData struct {
	VLANs []VLANModel `json:"vlans"`
}

// Validate VLAN input
func (v *VLANInput) Validate() error {
	// Validate name
	if v.Name == "" || len(v.Name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}

	// Validate VLAN ID
	if v.VlanID < 1 || v.VlanID > 4094 {
		return fmt.Errorf("vlan_id must be between 1 and 4094")
	}

	// Validate subnet CIDR
	if !isValidCIDR(v.Subnet) {
		return fmt.Errorf("invalid subnet format, must be in CIDR notation (e.g., 192.168.1.0/24)")
	}

	// Validate gateway IP
	if !isValidIP(v.Gateway) {
		return fmt.Errorf("invalid gateway IP address format")
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":      true,
		"inactive":    true,
		"maintenance": true,
	}
	if !validStatuses[v.Status] {
		return fmt.Errorf("status must be one of: active, inactive, maintenance")
	}

	return nil
}

// Validate CIDR notation
func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// Validate IP address
func isValidIP(ip string) bool {
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		return false
	}

	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}
