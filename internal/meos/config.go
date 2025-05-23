package meos

import (
	"fmt"
	"time"
)

type Config struct {
	Hostname     string
	Port         int
	PollInterval time.Duration
	HTTPS        bool
}

func NewConfig() *Config {
	// In WSL, use the Windows host IP (default gateway)
	return &Config{
		Hostname:     "192.168.112.1",
		Port:         2009,
		PollInterval: 1 * time.Second,
		HTTPS:        false,
	}
}

func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}
	if c.PollInterval < 100*time.Millisecond {
		return fmt.Errorf("poll interval too small (minimum 100ms): %s", c.PollInterval)
	}
	if c.PollInterval > 1*time.Hour {
		return fmt.Errorf("poll interval too large (maximum 1 hour): %s", c.PollInterval)
	}
	return nil
}
