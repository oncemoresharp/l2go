package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ToolkitConfig represents the complete configuration for the client toolkit
type ToolkitConfig struct {
	Client   ClientConfig   `json:"client"`
	Manager  ManagerConfig  `json:"manager"`
	LoadTest LoadTestConfig `json:"loadTest"`
	Logging  LoggingConfig  `json:"logging"`
	Profiles ProfilesConfig `json:"profiles"`
}

// ManagerConfig holds configuration for the client manager
type ManagerConfig struct {
	MaxClients      int           `json:"maxClients"`
	ConnectInterval time.Duration `json:"connectInterval"`
	HealthCheck     time.Duration `json:"healthCheck"`
	RetryAttempts   int           `json:"retryAttempts"`
	RetryDelay      time.Duration `json:"retryDelay"`
}

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	DefaultClientCount int           `json:"defaultClientCount"`
	DefaultDuration    time.Duration `json:"defaultDuration"`
	DefaultRampUpTime  time.Duration `json:"defaultRampUpTime"`
	MaxConcurrentTests int           `json:"maxConcurrentTests"`
	ReportFormat       string        `json:"reportFormat"`
}

// LoggingConfig holds configuration for logging
type LoggingConfig struct {
	Level         string `json:"level"`
	Format        string `json:"format"`
	Output        string `json:"output"`
	PacketLogging bool   `json:"packetLogging"`
	RotateSize    int64  `json:"rotateSize"`
	RotateCount   int    `json:"rotateCount"`
}

// ProfilesConfig holds different environment profiles
type ProfilesConfig struct {
	Development *EnvironmentProfile `json:"development"`
	Testing     *EnvironmentProfile `json:"testing"`
	Production  *EnvironmentProfile `json:"production"`
	Active      string              `json:"active"`
}

// EnvironmentProfile represents configuration for a specific environment
type EnvironmentProfile struct {
	LoginServer ServerProfile      `json:"loginServer"`
	GameServer  ServerProfile      `json:"gameServer"`
	Credentials CredentialsProfile `json:"credentials"`
}

// ServerProfile holds server connection configuration
type ServerProfile struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// CredentialsProfile holds authentication configuration
type CredentialsProfile struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	AutoCreate bool   `json:"autoCreate"`
}

// DefaultToolkitConfig returns a default configuration
func DefaultToolkitConfig() *ToolkitConfig {
	return &ToolkitConfig{
		Client: ClientConfig{
			LoginServerHost: "127.0.0.1",
			LoginServerPort: 2106,
			GameServerHost:  "127.0.0.1",
			GameServerPort:  7777,
			Username:        "testuser",
			Password:        "testpass",
			AutoCreate:      true,
			Timeout:         30 * time.Second,
		},
		Manager: ManagerConfig{
			MaxClients:      1000,
			ConnectInterval: 100 * time.Millisecond,
			HealthCheck:     5 * time.Second,
			RetryAttempts:   3,
			RetryDelay:      1 * time.Second,
		},
		LoadTest: LoadTestConfig{
			DefaultClientCount: 10,
			DefaultDuration:    60 * time.Second,
			DefaultRampUpTime:  10 * time.Second,
			MaxConcurrentTests: 5,
			ReportFormat:       "json",
		},
		Logging: LoggingConfig{
			Level:         "info",
			Format:        "json",
			Output:        "stdout",
			PacketLogging: false,
			RotateSize:    100 * 1024 * 1024, // 100MB
			RotateCount:   5,
		},
		Profiles: ProfilesConfig{
			Active: "development",
			Development: &EnvironmentProfile{
				LoginServer: ServerProfile{
					Host:    "127.0.0.1",
					Port:    2106,
					Timeout: 30 * time.Second,
				},
				GameServer: ServerProfile{
					Host:    "127.0.0.1",
					Port:    7777,
					Timeout: 30 * time.Second,
				},
				Credentials: CredentialsProfile{
					Username:   "devuser",
					Password:   "devpass",
					AutoCreate: true,
				},
			},
			Testing: &EnvironmentProfile{
				LoginServer: ServerProfile{
					Host:    "test.l2go.local",
					Port:    2106,
					Timeout: 15 * time.Second,
				},
				GameServer: ServerProfile{
					Host:    "test.l2go.local",
					Port:    7777,
					Timeout: 15 * time.Second,
				},
				Credentials: CredentialsProfile{
					Username:   "testuser",
					Password:   "testpass",
					AutoCreate: false,
				},
			},
			Production: &EnvironmentProfile{
				LoginServer: ServerProfile{
					Host:    "login.l2go.com",
					Port:    2106,
					Timeout: 10 * time.Second,
				},
				GameServer: ServerProfile{
					Host:    "game.l2go.com",
					Port:    7777,
					Timeout: 10 * time.Second,
				},
				Credentials: CredentialsProfile{
					Username:   "",
					Password:   "",
					AutoCreate: false,
				},
			},
		},
	}
}

// Validate validates the toolkit configuration
func (tc *ToolkitConfig) Validate() error {
	// Validate client configuration
	if err := tc.Client.Validate(); err != nil {
		return fmt.Errorf("client config validation failed: %w", err)
	}

	// Validate manager configuration
	if err := tc.Manager.Validate(); err != nil {
		return fmt.Errorf("manager config validation failed: %w", err)
	}

	// Validate load test configuration
	if err := tc.LoadTest.Validate(); err != nil {
		return fmt.Errorf("load test config validation failed: %w", err)
	}

	// Validate logging configuration
	if err := tc.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}

	// Validate profiles configuration
	if err := tc.Profiles.Validate(); err != nil {
		return fmt.Errorf("profiles config validation failed: %w", err)
	}

	return nil
}

// Validate validates the manager configuration
func (mc *ManagerConfig) Validate() error {
	if mc.MaxClients <= 0 {
		return fmt.Errorf("maxClients must be greater than 0, got %d", mc.MaxClients)
	}
	if mc.ConnectInterval < 0 {
		return fmt.Errorf("connectInterval must be non-negative, got %v", mc.ConnectInterval)
	}
	if mc.HealthCheck <= 0 {
		return fmt.Errorf("healthCheck must be greater than 0, got %v", mc.HealthCheck)
	}
	if mc.RetryAttempts < 0 {
		return fmt.Errorf("retryAttempts must be non-negative, got %d", mc.RetryAttempts)
	}
	if mc.RetryDelay < 0 {
		return fmt.Errorf("retryDelay must be non-negative, got %v", mc.RetryDelay)
	}
	return nil
}

// Validate validates the load test configuration
func (ltc *LoadTestConfig) Validate() error {
	if ltc.DefaultClientCount <= 0 {
		return fmt.Errorf("defaultClientCount must be greater than 0, got %d", ltc.DefaultClientCount)
	}
	if ltc.DefaultDuration <= 0 {
		return fmt.Errorf("defaultDuration must be greater than 0, got %v", ltc.DefaultDuration)
	}
	if ltc.DefaultRampUpTime < 0 {
		return fmt.Errorf("defaultRampUpTime must be non-negative, got %v", ltc.DefaultRampUpTime)
	}
	if ltc.MaxConcurrentTests <= 0 {
		return fmt.Errorf("maxConcurrentTests must be greater than 0, got %d", ltc.MaxConcurrentTests)
	}
	validFormats := map[string]bool{"json": true, "xml": true, "csv": true, "text": true}
	if !validFormats[ltc.ReportFormat] {
		return fmt.Errorf("invalid reportFormat: %s, must be one of: json, xml, csv, text", ltc.ReportFormat)
	}
	return nil
}

// Validate validates the logging configuration
func (lc *LoggingConfig) Validate() error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[lc.Level] {
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error", lc.Level)
	}
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[lc.Format] {
		return fmt.Errorf("invalid log format: %s, must be one of: json, text", lc.Format)
	}
	if lc.RotateSize <= 0 {
		return fmt.Errorf("rotateSize must be greater than 0, got %d", lc.RotateSize)
	}
	if lc.RotateCount <= 0 {
		return fmt.Errorf("rotateCount must be greater than 0, got %d", lc.RotateCount)
	}
	return nil
}

// Validate validates the profiles configuration
func (pc *ProfilesConfig) Validate() error {
	if pc.Active == "" {
		return fmt.Errorf("active profile must be specified")
	}

	profiles := map[string]*EnvironmentProfile{
		"development": pc.Development,
		"testing":     pc.Testing,
		"production":  pc.Production,
	}

	activeProfile := profiles[pc.Active]
	if activeProfile == nil {
		return fmt.Errorf("active profile '%s' not found", pc.Active)
	}

	return activeProfile.Validate()
}

// Validate validates an environment profile
func (ep *EnvironmentProfile) Validate() error {
	if err := ep.LoginServer.Validate(); err != nil {
		return fmt.Errorf("login server validation failed: %w", err)
	}
	if err := ep.GameServer.Validate(); err != nil {
		return fmt.Errorf("game server validation failed: %w", err)
	}
	if err := ep.Credentials.Validate(); err != nil {
		return fmt.Errorf("credentials validation failed: %w", err)
	}
	return nil
}

// Validate validates a server profile
func (sp *ServerProfile) Validate() error {
	if sp.Host == "" {
		return fmt.Errorf("host must not be empty")
	}
	if sp.Port <= 0 || sp.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", sp.Port)
	}
	if sp.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0, got %v", sp.Timeout)
	}
	return nil
}

// Validate validates credentials profile
func (cp *CredentialsProfile) Validate() error {
	if cp.Username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if cp.Password == "" {
		return fmt.Errorf("password must not be empty")
	}
	return nil
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*ToolkitConfig, error) {
	// If filename is empty, try default locations
	if filename == "" {
		filename = findConfigFile()
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config ToolkitConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *ToolkitConfig, filename string) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}

// findConfigFile searches for configuration files in standard locations
func findConfigFile() string {
	locations := []string{
		"./client-toolkit.json",
		"./config/client-toolkit.json",
		"~/.l2go/client-toolkit.json",
		"/etc/l2go/client-toolkit.json",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	// Return default location if none found
	return "./client-toolkit.json"
}

// GetActiveProfile returns the active environment profile
func (tc *ToolkitConfig) GetActiveProfile() (*EnvironmentProfile, error) {
	switch tc.Profiles.Active {
	case "development":
		if tc.Profiles.Development == nil {
			return nil, fmt.Errorf("development profile not configured")
		}
		return tc.Profiles.Development, nil
	case "testing":
		if tc.Profiles.Testing == nil {
			return nil, fmt.Errorf("testing profile not configured")
		}
		return tc.Profiles.Testing, nil
	case "production":
		if tc.Profiles.Production == nil {
			return nil, fmt.Errorf("production profile not configured")
		}
		return tc.Profiles.Production, nil
	default:
		return nil, fmt.Errorf("unknown active profile: %s", tc.Profiles.Active)
	}
}

// ApplyProfile applies the active profile settings to the client configuration
func (tc *ToolkitConfig) ApplyProfile() error {
	profile, err := tc.GetActiveProfile()
	if err != nil {
		return err
	}

	// Apply server settings
	tc.Client.LoginServerHost = profile.LoginServer.Host
	tc.Client.LoginServerPort = profile.LoginServer.Port
	tc.Client.GameServerHost = profile.GameServer.Host
	tc.Client.GameServerPort = profile.GameServer.Port
	tc.Client.Timeout = profile.LoginServer.Timeout

	// Apply credentials if not already set
	if tc.Client.Username == "" {
		tc.Client.Username = profile.Credentials.Username
	}
	if tc.Client.Password == "" {
		tc.Client.Password = profile.Credentials.Password
	}
	tc.Client.AutoCreate = profile.Credentials.AutoCreate

	return nil
}
