package meos

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config with default values",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         2009,
				PollInterval: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with minimum poll interval",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         2009,
				PollInterval: 100 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "valid config with maximum poll interval",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         2009,
				PollInterval: 1 * time.Hour,
			},
			wantErr: false,
		},
		{
			name: "invalid port - negative",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         -1,
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid port number",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         70000,
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid port number",
		},
		{
			name: "poll interval too small",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         2009,
				PollInterval: 50 * time.Millisecond,
			},
			wantErr:     true,
			errContains: "poll interval too small",
		},
		{
			name: "poll interval too large",
			config: Config{
				Hostname:     "192.168.112.1",
				Port:         2009,
				PollInterval: 2 * time.Hour,
			},
			wantErr:     true,
			errContains: "poll interval too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	
	if config.Hostname != "192.168.112.1" {
		t.Errorf("NewConfig() Hostname = %v, want %v", config.Hostname, "192.168.112.1")
	}
	if config.Port != 2009 {
		t.Errorf("NewConfig() Port = %v, want %v", config.Port, 2009)
	}
	if config.PollInterval != 1*time.Second {
		t.Errorf("NewConfig() PollInterval = %v, want %v", config.PollInterval, 1*time.Second)
	}
	if config.HTTPS != false {
		t.Errorf("NewConfig() HTTPS = %v, want %v", config.HTTPS, false)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) >= len(substr) && contains(s[1:], substr)
}