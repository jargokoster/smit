package models

import (
	"testing"
)

func TestVLANInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   VLANInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid input",
			input: VLANInput{
				Name:    "Production",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			input: VLANInput{
				Name:    "",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "name must be between 1 and 255 characters",
		},
		{
			name: "Name too long",
			input: VLANInput{
				Name:    string(make([]byte, 256)),
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "name must be between 1 and 255 characters",
		},
		{
			name: "VLAN ID too low",
			input: VLANInput{
				Name:    "Test",
				VlanID:  0,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "vlan_id must be between 1 and 4094",
		},
		{
			name: "VLAN ID too high",
			input: VLANInput{
				Name:    "Test",
				VlanID:  4095,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "vlan_id must be between 1 and 4094",
		},
		{
			name: "Invalid subnet - not CIDR",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "invalid subnet format, must be in CIDR notation (e.g., 192.168.1.0/24)",
		},
		{
			name: "Invalid subnet - wrong format",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "not-a-subnet",
				Gateway: "192.168.1.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "invalid subnet format, must be in CIDR notation (e.g., 192.168.1.0/24)",
		},
		{
			name: "Invalid gateway - wrong format",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "not-an-ip",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "invalid gateway IP address format",
		},
		{
			name: "Invalid gateway - incomplete IP",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "invalid gateway IP address format",
		},
		{
			name: "Invalid gateway - out of range",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.256",
				Status:  "active",
			},
			wantErr: true,
			errMsg:  "invalid gateway IP address format",
		},
		{
			name: "Invalid status",
			input: VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: "192.168.1.1",
				Status:  "unknown",
			},
			wantErr: true,
			errMsg:  "status must be one of: active, inactive, maintenance",
		},
		{
			name: "Valid with max VLAN ID",
			input: VLANInput{
				Name:    "MaxVLAN",
				VlanID:  4094,
				Subnet:  "10.0.0.0/8",
				Gateway: "10.0.0.1",
				Status:  "maintenance",
			},
			wantErr: false,
		},
		{
			name: "Valid with min VLAN ID",
			input: VLANInput{
				Name:    "MinVLAN",
				VlanID:  1,
				Subnet:  "172.16.0.0/16",
				Gateway: "172.16.0.1",
				Status:  "inactive",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCIDRValidation(t *testing.T) {
	tests := []struct {
		cidr  string
		valid bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/8", true},
		{"172.16.0.0/16", true},
		{"192.168.1.0/32", true},
		{"192.168.1.0/0", true},
		{"192.168.1.0", false},
		{"192.168.1.0/", false},
		{"192.168.1.0/33", false},
		{"not-a-cidr", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			input := VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  tt.cidr,
				Gateway: "192.168.1.1",
				Status:  "active",
			}
			err := input.Validate()
			hasError := err != nil && err.Error() == "invalid subnet format, must be in CIDR notation (e.g., 192.168.1.0/24)"
			if tt.valid && hasError {
				t.Errorf("Expected valid CIDR %s to pass validation", tt.cidr)
			}
			if !tt.valid && !hasError {
				t.Errorf("Expected invalid CIDR %s to fail validation", tt.cidr)
			}
		})
	}
}

func TestIPValidation(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"192.168.1", false},
		{"192.168.1.1.1", false},
		{"256.1.1.1", false},
		{"1.256.1.1", false},
		{"1.1.256.1", false},
		{"1.1.1.256", false},
		{"not-an-ip", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			input := VLANInput{
				Name:    "Test",
				VlanID:  100,
				Subnet:  "192.168.1.0/24",
				Gateway: tt.ip,
				Status:  "active",
			}
			err := input.Validate()
			hasError := err != nil && err.Error() == "invalid gateway IP address format"
			if tt.valid && hasError {
				t.Errorf("Expected valid IP %s to pass validation", tt.ip)
			}
			if !tt.valid && !hasError {
				t.Errorf("Expected invalid IP %s to fail validation", tt.ip)
			}
		})
	}
}
