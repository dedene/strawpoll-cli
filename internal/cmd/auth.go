package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/dedene/strawpoll-cli/internal/auth"
	"github.com/dedene/strawpoll-cli/internal/config"
)

// AuthCmd manages API key storage.
type AuthCmd struct {
	SetKey AuthSetKeyCmd `cmd:"" name:"set-key" help:"Store API key in keyring"`
	Status AuthStatusCmd `cmd:"" help:"Show API key status"`
	Remove AuthRemoveCmd `cmd:"" help:"Remove stored API key"`
}

// AuthSetKeyCmd stores an API key in the system keyring.
type AuthSetKeyCmd struct {
	Stdin bool `help:"Read API key from stdin (for scripts)"`
}

// Run prompts for an API key and stores it.
func (c *AuthSetKeyCmd) Run() error {
	var key string

	if c.Stdin {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			key = strings.TrimSpace(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read from stdin: %w", err)
		}
	} else {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("not a terminal; use --stdin flag to read from pipe")
		}

		fmt.Print("Enter your StrawPoll API key: ")

		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()

		if err != nil {
			return fmt.Errorf("read API key: %w", err)
		}

		key = strings.TrimSpace(string(bytes))
	}

	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := store.SetAPIKey(key); err != nil {
		return fmt.Errorf("store API key: %w", err)
	}

	fmt.Fprintln(os.Stdout, "API key stored successfully.")
	fmt.Fprintln(os.Stdout, "Get your API key from https://strawpoll.com/account/settings")

	return nil
}

// AuthStatusCmd shows the current API key status.
type AuthStatusCmd struct{}

// Run checks env var, keyring, and reports status.
func (c *AuthStatusCmd) Run() error {
	// Check env var first
	if os.Getenv("STRAWPOLL_API_KEY") != "" {
		fmt.Fprintln(os.Stdout, "API key: set via STRAWPOLL_API_KEY environment variable")

		return nil
	}

	backendInfo, err := auth.ResolveKeyringBackendInfo()
	if err != nil {
		return err
	}

	configPath, _ := config.ConfigPath()
	keyringDir, _ := config.KeyringDir()

	fmt.Fprintf(os.Stdout, "Config path:     %s\n", configPath)
	fmt.Fprintf(os.Stdout, "Keyring dir:     %s\n", keyringDir)
	fmt.Fprintf(os.Stdout, "Keyring backend: %s (source: %s)\n", backendInfo.Value, backendInfo.Source)

	store, err := auth.OpenDefault()
	if err != nil {
		fmt.Fprintf(os.Stdout, "API key:         error opening keyring: %v\n", err)

		return nil
	}

	hasKey, err := store.HasAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stdout, "API key:         error checking: %v\n", err)

		return nil
	}

	if hasKey {
		fmt.Fprintln(os.Stdout, "API key:         stored in system keyring")
	} else {
		fmt.Fprintln(os.Stdout, "API key:         not configured")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "Run 'strawpoll auth set-key' to configure your API key.")
		fmt.Fprintln(os.Stdout, "Get your API key from https://strawpoll.com/account/settings")
	}

	return nil
}

// AuthRemoveCmd removes the stored API key.
type AuthRemoveCmd struct{}

// Run deletes the API key from keyring.
func (c *AuthRemoveCmd) Run() error {
	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := store.DeleteAPIKey(); err != nil {
		if errors.Is(err, auth.ErrNoAPIKey) {
			fmt.Fprintln(os.Stdout, "No API key was stored.")

			return nil
		}

		return fmt.Errorf("remove API key: %w", err)
	}

	fmt.Fprintln(os.Stdout, "API key removed.")

	return nil
}
