package client

import (
	"testing"
	"time"
)

func TestClientConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ClientConfig{
				LoginServerHost: "127.0.0.1",
				LoginServerPort: 2106,
				GameServerHost:  "127.0.0.1",
				GameServerPort:  7777,
				Username:        "testuser",
				Password:        "testpass",
				AutoCreate:      true,
				Timeout:         30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty login server host",
			config: ClientConfig{
				LoginServerHost: "",
				LoginServerPort: 2106,
				GameServerHost:  "127.0.0.1",
				GameServerPort:  7777,
				Username:        "testuser",
				Password:        "testpass",
			},
			wantErr: true,
		},
		{
			name: "invalid login server port",
			config: ClientConfig{
				LoginServerHost: "127.0.0.1",
				LoginServerPort: 0,
				GameServerHost:  "127.0.0.1",
				GameServerPort:  7777,
				Username:        "testuser",
				Password:        "testpass",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			config: ClientConfig{
				LoginServerHost: "127.0.0.1",
				LoginServerPort: 2106,
				GameServerHost:  "127.0.0.1",
				GameServerPort:  7777,
				Username:        "",
				Password:        "testpass",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultToolkitConfig(t *testing.T) {
	config := DefaultToolkitConfig()

	if err := config.Validate(); err != nil {
		t.Errorf("DefaultToolkitConfig() validation failed: %v", err)
	}

	// Check that default values are set correctly
	if config.Client.LoginServerHost == "" {
		t.Error("Default config should have login server host set")
	}

	if config.Manager.MaxClients <= 0 {
		t.Error("Default config should have positive max clients")
	}

	if config.Profiles.Active == "" {
		t.Error("Default config should have active profile set")
	}
}

func TestManagerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ManagerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ManagerConfig{
				MaxClients:      100,
				ConnectInterval: 100 * time.Millisecond,
				HealthCheck:     5 * time.Second,
				RetryAttempts:   3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "zero max clients",
			config: ManagerConfig{
				MaxClients:      0,
				ConnectInterval: 100 * time.Millisecond,
				HealthCheck:     5 * time.Second,
				RetryAttempts:   3,
				RetryDelay:      1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative retry attempts",
			config: ManagerConfig{
				MaxClients:      100,
				ConnectInterval: 100 * time.Millisecond,
				HealthCheck:     5 * time.Second,
				RetryAttempts:   -1,
				RetryDelay:      1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ManagerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
