package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds cc-queue configuration.
type Config struct {
	Debug bool `json:"debug"`
}

// ConfigDir returns the configuration directory for cc-queue.
// Uses $XDG_CONFIG_HOME/cc-queue or defaults to ~/.config/cc-queue.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cc-queue")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cc-queue")
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// ReadConfig reads the configuration from disk.
// Returns a zero Config if the file doesn't exist or can't be parsed.
func ReadConfig() Config {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return Config{}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}
	return cfg
}

// WriteConfig writes the configuration to disk, creating the directory if needed.
func WriteConfig(cfg Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(ConfigPath(), data, 0644)
}

// DefaultConfigJSON returns the pretty-printed JSON of a zero-value Config.
func DefaultConfigJSON() string {
	data, _ := json.MarshalIndent(Config{}, "", "  ")
	return string(data)
}
