package meos

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Hostname     string
	Port         int
	PortStr      string // Original port string (can be "none")
	PollInterval time.Duration
	HTTPS        bool
}

func NewConfig() *Config {
	return &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
		HTTPS:        false,
	}
}

func (c *Config) Validate() error {
	// Validate hostname
	if c.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// Check if hostname is valid (either hostname or IP address)
	if !isValidHostname(c.Hostname) && net.ParseIP(c.Hostname) == nil {
		return fmt.Errorf("invalid hostname or IP address: %s", c.Hostname)
	}

	// Validate port
	if c.PortStr != "none" {
		port, err := strconv.Atoi(c.PortStr)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", c.PortStr)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port number out of range (1-65535): %d", port)
		}
		c.Port = port
	}

	// Validate poll interval
	if c.PollInterval < 100*time.Millisecond {
		return fmt.Errorf("poll interval too small (minimum 100ms): %s", c.PollInterval)
	}
	if c.PollInterval > 1*time.Hour {
		return fmt.Errorf("poll interval too large (maximum 1 hour): %s", c.PollInterval)
	}

	return nil
}

// isValidHostname checks if the string is a valid hostname
func isValidHostname(hostname string) bool {
	if len(hostname) > 253 {
		return false
	}

	// Check for valid characters and format
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}

		// Check if label starts or ends with hyphen
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		// Check if all characters are valid (alphanumeric or hyphen)
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-') {
				return false
			}
		}
	}

	return true
}
