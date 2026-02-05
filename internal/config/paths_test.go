package config

import (
	"strings"
	"testing"
)

func TestDir(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}

	if !strings.HasSuffix(dir, "strawpoll-cli") {
		t.Errorf("Dir() = %q, want suffix %q", dir, "strawpoll-cli")
	}
}

func TestConfigPath(t *testing.T) {
	p, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}

	if !strings.HasSuffix(p, "config.yaml") {
		t.Errorf("ConfigPath() = %q, want suffix %q", p, "config.yaml")
	}
}

func TestKeyringDir(t *testing.T) {
	p, err := KeyringDir()
	if err != nil {
		t.Fatalf("KeyringDir() error: %v", err)
	}

	if !strings.HasSuffix(p, "keyring") {
		t.Errorf("KeyringDir() = %q, want suffix %q", p, "keyring")
	}
}

func TestEnsureDir(t *testing.T) {
	dir, err := EnsureDir()
	if err != nil {
		t.Fatalf("EnsureDir() error: %v", err)
	}

	if !strings.HasSuffix(dir, "strawpoll-cli") {
		t.Errorf("EnsureDir() = %q, want suffix %q", dir, "strawpoll-cli")
	}
}

func TestEnsureKeyringDir(t *testing.T) {
	dir, err := EnsureKeyringDir()
	if err != nil {
		t.Fatalf("EnsureKeyringDir() error: %v", err)
	}

	if !strings.HasSuffix(dir, "keyring") {
		t.Errorf("EnsureKeyringDir() = %q, want suffix %q", dir, "keyring")
	}
}
