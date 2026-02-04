package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// File represents the config file structure.
type File struct {
	DefaultProfile string             `yaml:"default_profile,omitempty"`
	Profiles       map[string]Profile `yaml:"profiles,omitempty"`
	KeyringBackend string             `yaml:"keyring_backend,omitempty"`
}

// Profile represents a named profile configuration.
type Profile struct {
	// Credentials are stored in keyring, not here
	AccountID string `yaml:"account_id,omitempty"`
}

// ReadConfig reads the config file.
func ReadConfig() (File, error) {
	path, err := ConfigPath()
	if err != nil {
		return File{}, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}

		return File{}, fmt.Errorf("read config: %w", err)
	}

	var cfg File
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	return cfg, nil
}

// WriteConfig writes the config file atomically.
func WriteConfig(cfg File) error {
	if _, err := EnsureDir(); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config yaml: %w", err)
	}

	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit config: %w", err)
	}

	return nil
}

// ConfigExists checks if the config file exists.
func ConfigExists() (bool, error) {
	path, err := ConfigPath()
	if err != nil {
		return false, err
	}

	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return false, nil
		}

		return false, fmt.Errorf("stat config: %w", statErr)
	}

	return true, nil
}
