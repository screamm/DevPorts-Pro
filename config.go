package main

import (
	"fmt"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Scanning configuration
	PortRangeStart int
	PortRangeEnd   int
	NumWorkers     int
	PortTimeout    time.Duration
	CommandTimeout time.Duration

	// Auto-refresh configuration
	AutoRefreshInterval time.Duration

	// Kill verification configuration
	KillVerifyAttempts   int
	KillVerifyBaseDelay  time.Duration
	PostKillRefreshDelay time.Duration

	// UI configuration
	WindowWidth  float32
	WindowHeight float32
}

// Validate checks that all configuration values are within acceptable ranges
func (c *Config) Validate() error {
	// Port range validation
	if c.PortRangeStart < 1 || c.PortRangeStart > 65535 {
		return fmt.Errorf("invalid PortRangeStart: %d (must be 1-65535)", c.PortRangeStart)
	}
	if c.PortRangeEnd < 1 || c.PortRangeEnd > 65535 {
		return fmt.Errorf("invalid PortRangeEnd: %d (must be 1-65535)", c.PortRangeEnd)
	}
	if c.PortRangeEnd < c.PortRangeStart {
		return fmt.Errorf("PortRangeEnd (%d) must be >= PortRangeStart (%d)", c.PortRangeEnd, c.PortRangeStart)
	}

	// Worker count validation
	if c.NumWorkers < 1 {
		return fmt.Errorf("invalid NumWorkers: %d (must be >= 1)", c.NumWorkers)
	}
	if c.NumWorkers > 10000 {
		return fmt.Errorf("invalid NumWorkers: %d (must be <= 10000)", c.NumWorkers)
	}

	// Timeout validation
	if c.PortTimeout <= 0 {
		return fmt.Errorf("invalid PortTimeout: %v (must be > 0)", c.PortTimeout)
	}
	if c.CommandTimeout <= 0 {
		return fmt.Errorf("invalid CommandTimeout: %v (must be > 0)", c.CommandTimeout)
	}

	// Auto-refresh validation
	if c.AutoRefreshInterval < 10*time.Second {
		return fmt.Errorf("invalid AutoRefreshInterval: %v (must be >= 10s)", c.AutoRefreshInterval)
	}

	// Kill verification validation
	if c.KillVerifyAttempts < 1 {
		return fmt.Errorf("invalid KillVerifyAttempts: %d (must be >= 1)", c.KillVerifyAttempts)
	}
	if c.KillVerifyBaseDelay <= 0 {
		return fmt.Errorf("invalid KillVerifyBaseDelay: %v (must be > 0)", c.KillVerifyBaseDelay)
	}
	if c.PostKillRefreshDelay < 0 {
		return fmt.Errorf("invalid PostKillRefreshDelay: %v (must be >= 0)", c.PostKillRefreshDelay)
	}

	// UI validation
	if c.WindowWidth < 400 {
		return fmt.Errorf("invalid WindowWidth: %v (must be >= 400)", c.WindowWidth)
	}
	if c.WindowHeight < 300 {
		return fmt.Errorf("invalid WindowHeight: %v (must be >= 300)", c.WindowHeight)
	}

	return nil
}

// DefaultConfig returns the default application configuration
func DefaultConfig() *Config {
	return &Config{
		// Scanning
		PortRangeStart: 1,
		PortRangeEnd:   9999,
		NumWorkers:     500,
		PortTimeout:    100 * time.Millisecond,
		CommandTimeout: 5 * time.Second,

		// Auto-refresh
		AutoRefreshInterval: 5 * time.Minute,

		// Kill verification
		KillVerifyAttempts:   5,
		KillVerifyBaseDelay:  200 * time.Millisecond,
		PostKillRefreshDelay: 1500 * time.Millisecond,

		// UI
		WindowWidth:  1100,
		WindowHeight: 800,
	}
}

// AppConfig is the global configuration instance
var AppConfig = DefaultConfig()
