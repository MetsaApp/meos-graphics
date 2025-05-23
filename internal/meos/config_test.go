package meos

import (
	"strings"
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
			name: "valid config with hostname",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with IP address",
			config: Config{
				Hostname:     "192.168.1.1",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with domain name",
			config: Config{
				Hostname:     "meos.example.com",
				PortStr:      "8080",
				PollInterval: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with 'none' port",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "none",
				PollInterval: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty hostname",
			config: Config{
				Hostname:     "",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "hostname cannot be empty",
		},
		{
			name: "invalid hostname - starts with hyphen",
			config: Config{
				Hostname:     "-invalid.com",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid hostname or IP address",
		},
		{
			name: "invalid hostname - ends with hyphen",
			config: Config{
				Hostname:     "invalid-.com",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid hostname or IP address",
		},
		{
			name: "invalid hostname - special characters",
			config: Config{
				Hostname:     "invalid@host.com",
				PortStr:      "2009",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid hostname or IP address",
		},
		{
			name: "invalid port - not a number",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "abc",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "invalid port number",
		},
		{
			name: "invalid port - zero",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "0",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "port number out of range",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "70000",
				PollInterval: 1 * time.Second,
			},
			wantErr:     true,
			errContains: "port number out of range",
		},
		{
			name: "poll interval too small",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "2009",
				PollInterval: 50 * time.Millisecond,
			},
			wantErr:     true,
			errContains: "poll interval too small",
		},
		{
			name: "poll interval too large",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "2009",
				PollInterval: 2 * time.Hour,
			},
			wantErr:     true,
			errContains: "poll interval too large",
		},
		{
			name: "valid config with minimum poll interval",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "2009",
				PollInterval: 100 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "valid config with maximum poll interval",
			config: Config{
				Hostname:     "localhost",
				PortStr:      "2009",
				PollInterval: 1 * time.Hour,
			},
			wantErr: false,
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

	if config.Hostname != "localhost" {
		t.Errorf("NewConfig() Hostname = %v, want %v", config.Hostname, "localhost")
	}
	if config.Port != 2009 {
		t.Errorf("NewConfig() Port = %v, want %v", config.Port, 2009)
	}
	if config.PortStr != "2009" {
		t.Errorf("NewConfig() PortStr = %v, want %v", config.PortStr, "2009")
	}
	if config.PollInterval != 1*time.Second {
		t.Errorf("NewConfig() PollInterval = %v, want %v", config.PollInterval, 1*time.Second)
	}
	if config.HTTPS != false {
		t.Errorf("NewConfig() HTTPS = %v, want %v", config.HTTPS, false)
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		hostname string
		want     bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"sub.example.com", true},
		{"example-site.com", true},
		{"123.example.com", true},
		{"example123.com", true},
		{"a.b.c.d.e.f", true},
		{"", false},
		{"-example.com", false},
		{"example-.com", false},
		{"example..com", false},
		{"example@.com", false},
		{"example .com", false},
		{"example.com-", false},
		{"-example", false},
		{"example-", false},
		{"exa mple.com", false},
		{"example!.com", false},
		{strings.Repeat("a", 254), false},         // Too long
		{strings.Repeat("a", 64) + ".com", false}, // Label too long
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			if got := isValidHostname(tt.hostname); got != tt.want {
				t.Errorf("isValidHostname(%q) = %v, want %v", tt.hostname, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) >= len(substr) && contains(s[1:], substr)
}
