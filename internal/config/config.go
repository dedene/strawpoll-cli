package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// File represents the strawpoll-cli YAML configuration.
type File struct {
	KeyringBackend    string `yaml:"keyring_backend,omitempty" json:"keyring_backend,omitempty"`
	Dupcheck          string `yaml:"dupcheck,omitempty" json:"dupcheck,omitempty"`
	ResultsVisibility string `yaml:"results_visibility,omitempty" json:"results_visibility,omitempty"`
	IsPrivate         *bool  `yaml:"is_private,omitempty" json:"is_private,omitempty"`
	AllowComments     *bool  `yaml:"allow_comments,omitempty" json:"allow_comments,omitempty"`
	AllowVPN          *bool  `yaml:"allow_vpn_users,omitempty" json:"allow_vpn_users,omitempty"`
	HideParticipants  *bool  `yaml:"hide_participants,omitempty" json:"hide_participants,omitempty"`
	EditVotePerms     string `yaml:"edit_vote_permissions,omitempty" json:"edit_vote_permissions,omitempty"`
}

// ConfigExists checks whether the config file exists on disk.
func ConfigExists() (bool, error) {
	path, err := ConfigPath()
	if err != nil {
		return false, err
	}

	return configExistsAt(path)
}

func configExistsAt(path string) (bool, error) {
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return false, nil
		}

		return false, fmt.Errorf("stat config: %w", statErr)
	}

	return true, nil
}

// ReadConfig reads the YAML config file. Returns zero File{} if file does not exist.
func ReadConfig() (File, error) {
	path, err := ConfigPath()
	if err != nil {
		return File{}, err
	}

	return readConfigFrom(path)
}

func readConfigFrom(path string) (File, error) {
	b, err := os.ReadFile(path) //nolint:gosec // config file path
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

// WriteConfig writes the YAML config file atomically using a .tmp + rename pattern.
func WriteConfig(cfg File) error {
	_, err := EnsureDir()
	if err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	return writeConfigTo(path, cfg)
}

func writeConfigTo(path string, cfg File) error {
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
