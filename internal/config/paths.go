package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// AppName is the application name used for config directories.
const AppName = "ponto"

// Dir returns the config directory path.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, AppName), nil
}

// EnsureDir creates the config directory if it doesn't exist.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("ensure config dir: %w", err)
	}

	return dir, nil
}

// KeyringDir returns the keyring directory path.
func KeyringDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "keyring"), nil
}

// EnsureKeyringDir creates the keyring directory if it doesn't exist.
func EnsureKeyringDir() (string, error) {
	dir, err := KeyringDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("ensure keyring dir: %w", err)
	}

	return dir, nil
}

// ConfigPath returns the config file path.
func ConfigPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.yaml"), nil
}
