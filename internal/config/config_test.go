package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")

	cfg, err := readConfigFrom(path)
	if err != nil {
		t.Fatalf("readConfigFrom(missing) error: %v", err)
	}

	if cfg != (File{}) {
		t.Errorf("readConfigFrom(missing) = %+v, want zero File{}", cfg)
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	priv := true
	comments := false

	want := File{
		KeyringBackend:    "file",
		Dupcheck:          "ip",
		ResultsVisibility: "after_vote",
		IsPrivate:         &priv,
		AllowComments:     &comments,
		EditVotePerms:     "nobody",
	}

	if err := writeConfigTo(path, want); err != nil {
		t.Fatalf("writeConfigTo() error: %v", err)
	}

	got, err := readConfigFrom(path)
	if err != nil {
		t.Fatalf("readConfigFrom() error: %v", err)
	}

	if got.KeyringBackend != want.KeyringBackend {
		t.Errorf("KeyringBackend = %q, want %q", got.KeyringBackend, want.KeyringBackend)
	}

	if got.Dupcheck != want.Dupcheck {
		t.Errorf("Dupcheck = %q, want %q", got.Dupcheck, want.Dupcheck)
	}

	if got.ResultsVisibility != want.ResultsVisibility {
		t.Errorf("ResultsVisibility = %q, want %q", got.ResultsVisibility, want.ResultsVisibility)
	}

	if got.IsPrivate == nil || *got.IsPrivate != true {
		t.Errorf("IsPrivate = %v, want true", got.IsPrivate)
	}

	if got.AllowComments == nil || *got.AllowComments != false {
		t.Errorf("AllowComments = %v, want false", got.AllowComments)
	}

	if got.AllowVPN != nil {
		t.Errorf("AllowVPN = %v, want nil (unset)", got.AllowVPN)
	}

	if got.EditVotePerms != want.EditVotePerms {
		t.Errorf("EditVotePerms = %q, want %q", got.EditVotePerms, want.EditVotePerms)
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := File{Dupcheck: "cookie"}

	if err := writeConfigTo(path, cfg); err != nil {
		t.Fatalf("writeConfigTo() error: %v", err)
	}

	// .tmp file should not persist after successful write
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf(".tmp file still exists after atomic write: %v", err)
	}

	// Final file should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file does not exist after write: %v", err)
	}
}

func TestConfigExistsAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Should not exist initially
	exists, err := configExistsAt(path)
	if err != nil {
		t.Fatalf("configExistsAt() error: %v", err)
	}

	if exists {
		t.Error("configExistsAt() = true for missing file")
	}

	// Create file
	if err := writeConfigTo(path, File{Dupcheck: "ip"}); err != nil {
		t.Fatalf("writeConfigTo() error: %v", err)
	}

	exists, err = configExistsAt(path)
	if err != nil {
		t.Fatalf("configExistsAt() error: %v", err)
	}

	if !exists {
		t.Error("configExistsAt() = false for existing file")
	}
}
