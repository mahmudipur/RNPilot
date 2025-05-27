package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectConfig represents the structure of the project's configuration
type ProjectConfig struct {
	ProjectType    string `json:"projectType"`
	CreatedAt      string `json:"createdAt"`
	EnvInitialized bool   `json:"envInitialized"`
}

// ConfigFileName is the name of the config file
const ConfigFileName = ".RNPilot-config.json"

// LoadConfig loads the project configuration from the .RNPilot-config.json file
func LoadConfig(projectPath string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectPath, ConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %v", err)
	}

	var config ProjectConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project config: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the given ProjectConfig to the .RNPilot-config.json file
func SaveConfig(projectPath string, config *ProjectConfig) error {
	configPath := filepath.Join(projectPath, ConfigFileName)

	// Serialize the config struct to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize project config: %v", err)
	}

	// Write the config file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write project config: %v", err)
	}

	fmt.Printf("Project configuration saved to %s\n", configPath)
	return nil
}

// InitConfig creates and saves the initial configuration (after project init)
func InitConfig(projectType, projectPath string) error {
	config := &ProjectConfig{
		ProjectType:    projectType,
		CreatedAt:      time.Now().Format(time.RFC3339),
		EnvInitialized: false,
	}
	return SaveConfig(projectPath, config)
}

// UpdateConfig updates specific fields in the existing configuration
func UpdateConfig(projectPath string, updateFn func(config *ProjectConfig)) error {
	// Load the existing configuration
	config, err := LoadConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config for update: %v", err)
	}

	// Apply the update function
	updateFn(config)

	// Save the updated configuration
	return SaveConfig(projectPath, config)
}
