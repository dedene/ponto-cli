package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	t.Parallel()

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}

	if dir == "" {
		t.Error("Dir() returned empty string")
	}

	// Should end with "ponto"
	if filepath.Base(dir) != AppName {
		t.Errorf("Dir() = %q, want to end with %q", dir, AppName)
	}
}

func TestConfigPath(t *testing.T) {
	t.Parallel()

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}

	// Should end with "config.yaml"
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("ConfigPath() = %q, want to end with config.yaml", path)
	}
}

func TestEnsureDir(t *testing.T) {
	// Use a temp dir for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)

	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	dir, err := EnsureDir()
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	// Directory should exist
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", dir, err)
	}

	if !info.IsDir() {
		t.Errorf("%q is not a directory", dir)
	}
}
