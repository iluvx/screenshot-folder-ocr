package sharex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type applicationConfig struct {
	UseCustomScreenshotsPath bool   `json:"UseCustomScreenshotsPath"`
	CustomScreenshotsPath    string `json:"CustomScreenshotsPath"`
}

// DefaultDirectory returns ShareX's default user data directory.
func DefaultDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("Documents", "ShareX")
	}
	return filepath.Join(home, "Documents", "ShareX")
}

// DefaultScreenshotsPath returns ShareX's configured screenshots directory.
func DefaultScreenshotsPath() (string, error) {
	configPath := filepath.Join(DefaultDirectory(), "ApplicationConfig.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Join(DefaultDirectory(), "Screenshots"), nil
		}
		return "", fmt.Errorf("read ShareX config: %w", err)
	}

	var config applicationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse ShareX config: %w", err)
	}
	if config.UseCustomScreenshotsPath && config.CustomScreenshotsPath != "" {
		return filepath.Clean(os.ExpandEnv(config.CustomScreenshotsPath)), nil
	}
	return filepath.Join(DefaultDirectory(), "Screenshots"), nil
}
