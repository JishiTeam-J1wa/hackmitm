package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig represents the application configuration
type AppConfig struct {
	// Connection settings
	LastConnectionMode string `json:"lastConnectionMode"`

	// Local mode settings
	LocalDataDir string `json:"localDataDir,omitempty"`
	LocalAPIPort int    `json:"localApiPort,omitempty"`
	LocalProxyPort int  `json:"localProxyPort,omitempty"`

	// Remote mode settings
	RemoteHost   string `json:"remoteHost,omitempty"`
	RemotePort   int    `json:"remotePort,omitempty"`
	RemoteAPIKey string `json:"remoteApiKey,omitempty"`

	// UI settings
	Theme          string `json:"theme,omitempty"`
	WindowWidth    int    `json:"windowWidth,omitempty"`
	WindowHeight   int    `json:"windowHeight,omitempty"`
	WindowX        int    `json:"windowX,omitempty"`
	WindowY        int    `json:"windowY,omitempty"`

	// Proxy settings
	InterceptEnabled bool `json:"interceptEnabled"`

	// Filter settings
	DefaultMethodFilter string `json:"defaultMethodFilter,omitempty"`
	DefaultStatusFilter string `json:"defaultStatusFilter,omitempty"`
}

// ConfigManager handles application configuration persistence
type ConfigManager struct {
	config     *AppConfig
	configPath string
	mu         sync.RWMutex
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		configDir = "."
	}

	configPath := filepath.Join(configDir, "config.json")

	cm := &ConfigManager{
		configPath: configPath,
		config:     &AppConfig{},
	}

	// Load existing config
	cm.Load()

	return cm
}

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	// Try user config directory first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".hackmitm")
	} else {
		configDir = filepath.Join(configDir, "hackmitm")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// Load loads the configuration from disk
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			cm.config = &AppConfig{
				LastConnectionMode: "",
				LocalAPIPort:       9090,
				LocalProxyPort:     4443,
				RemotePort:         9090,
				Theme:              "light",
				WindowWidth:        1280,
				WindowHeight:       800,
			}
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save saves the configuration to disk
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Set updates the configuration and saves to disk
func (cm *ConfigManager) Set(config *AppConfig) error {
	cm.mu.Lock()
	cm.config = config
	cm.mu.Unlock()
	return cm.Save()
}

// Update updates specific fields of the configuration
func (cm *ConfigManager) Update(updates map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Apply updates
	if v, ok := updates["lastConnectionMode"].(string); ok {
		cm.config.LastConnectionMode = v
	}
	if v, ok := updates["localDataDir"].(string); ok {
		cm.config.LocalDataDir = v
	}
	if v, ok := updates["localApiPort"].(float64); ok {
		cm.config.LocalAPIPort = int(v)
	}
	if v, ok := updates["localProxyPort"].(float64); ok {
		cm.config.LocalProxyPort = int(v)
	}
	if v, ok := updates["remoteHost"].(string); ok {
		cm.config.RemoteHost = v
	}
	if v, ok := updates["remotePort"].(float64); ok {
		cm.config.RemotePort = int(v)
	}
	if v, ok := updates["remoteApiKey"].(string); ok {
		cm.config.RemoteAPIKey = v
	}
	if v, ok := updates["theme"].(string); ok {
		cm.config.Theme = v
	}
	if v, ok := updates["windowWidth"].(float64); ok {
		cm.config.WindowWidth = int(v)
	}
	if v, ok := updates["windowHeight"].(float64); ok {
		cm.config.WindowHeight = int(v)
	}
	if v, ok := updates["windowX"].(float64); ok {
		cm.config.WindowX = int(v)
	}
	if v, ok := updates["windowY"].(float64); ok {
		cm.config.WindowY = int(v)
	}
	if v, ok := updates["interceptEnabled"].(bool); ok {
		cm.config.InterceptEnabled = v
	}
	if v, ok := updates["defaultMethodFilter"].(string); ok {
		cm.config.DefaultMethodFilter = v
	}
	if v, ok := updates["defaultStatusFilter"].(string); ok {
		cm.config.DefaultStatusFilter = v
	}

	return cm.Save()
}

// Reset resets the configuration to defaults
func (cm *ConfigManager) Reset() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = &AppConfig{
		LastConnectionMode: "",
		LocalAPIPort:       9090,
		LocalProxyPort:     4443,
		RemotePort:         9090,
		Theme:              "light",
		WindowWidth:        1280,
		WindowHeight:       800,
	}

	return cm.Save()
}

// GetConfigPath returns the path to the config file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
