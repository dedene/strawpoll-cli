package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/dedene/strawpoll-cli/internal/config"
)

// ConfigCmd manages CLI configuration.
type ConfigCmd struct {
	Show ConfigShowCmd `cmd:"" help:"Display current configuration"`
	Set  ConfigSetCmd  `cmd:"" help:"Set a configuration value"`
	Path ConfigPathCmd `cmd:"" help:"Show configuration file path"`
}

// ConfigShowCmd displays the current configuration.
type ConfigShowCmd struct{}

// Run reads and prints the config file.
func (c *ConfigShowCmd) Run(flags *RootFlags) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if flags.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)

		return enc.Encode(cfg)
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	out := strings.TrimSpace(string(b))
	if out == "{}" || out == "" {
		fmt.Fprintln(os.Stdout, "# No configuration set (using defaults)")

		return nil
	}

	fmt.Fprintln(os.Stdout, out)

	return nil
}

// ConfigSetCmd sets a configuration value.
type ConfigSetCmd struct {
	Key   string `arg:"" required:"" help:"Configuration key"`
	Value string `arg:"" required:"" help:"Configuration value"`
}

// Run sets a config key to the given value.
func (c *ConfigSetCmd) Run() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	switch c.Key {
	case "keyring_backend":
		cfg.KeyringBackend = c.Value
	case "dupcheck":
		cfg.Dupcheck = c.Value
	case "results_visibility":
		cfg.ResultsVisibility = c.Value
	case "is_private":
		b, err := parseBool(c.Value)
		if err != nil {
			return fmt.Errorf("invalid boolean for is_private: %w", err)
		}

		cfg.IsPrivate = &b
	case "allow_comments":
		b, err := parseBool(c.Value)
		if err != nil {
			return fmt.Errorf("invalid boolean for allow_comments: %w", err)
		}

		cfg.AllowComments = &b
	case "allow_vpn_users":
		b, err := parseBool(c.Value)
		if err != nil {
			return fmt.Errorf("invalid boolean for allow_vpn_users: %w", err)
		}

		cfg.AllowVPN = &b
	case "hide_participants":
		b, err := parseBool(c.Value)
		if err != nil {
			return fmt.Errorf("invalid boolean for hide_participants: %w", err)
		}

		cfg.HideParticipants = &b
	case "edit_vote_permissions":
		cfg.EditVotePerms = c.Value
	default:
		return fmt.Errorf("unknown config key: %s\n\nValid keys: keyring_backend, dupcheck, results_visibility, is_private, allow_comments, allow_vpn_users, hide_participants, edit_vote_permissions", c.Key)
	}

	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Fprintf(os.Stdout, "%s = %s\n", c.Key, c.Value)

	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("expected true/false, got %q", s)
	}
}

// ConfigPathCmd shows the config file path.
type ConfigPathCmd struct{}

// Run prints the config file path.
func (c *ConfigPathCmd) Run() error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, path)

	return nil
}
