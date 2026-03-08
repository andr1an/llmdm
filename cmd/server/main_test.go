package main

import "testing"

func TestValidateCampaignName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid simple", input: "Lost Mines", wantErr: false},
		{name: "valid max length", input: "A123456789012345678901234567890123456789012345678901234567890123", wantErr: false},
		{name: "starts with space", input: " Lost Mines", wantErr: true},
		{name: "starts with digit", input: "1st Campaign", wantErr: true},
		{name: "too long", input: "A1234567890123456789012345678901234567890123456789012345678901234", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCampaignName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("validateCampaignName(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateCampaignName(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}
