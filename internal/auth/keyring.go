package auth

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/dedene/strawpoll-cli/internal/config"
)

// Store defines the interface for API key storage.
type Store interface {
	SetAPIKey(key string) error
	GetAPIKey() (string, error)
	DeleteAPIKey() error
	HasAPIKey() (bool, error)
}

// KeyringStore implements Store using the system keyring.
type KeyringStore struct {
	ring keyring.Keyring
}

const (
	apiKeyKey          = "api_key"
	keyringPasswordEnv = "STRAWPOLL_KEYRING_PASSWORD" //nolint:gosec // env var name
	keyringBackendEnv  = "STRAWPOLL_KEYRING_BACKEND"  //nolint:gosec // env var name
	apiKeyEnv          = "STRAWPOLL_API_KEY"          //nolint:gosec // env var name
)

var (
	ErrNoAPIKey              = errors.New("no API key configured")
	errNoTTY                 = errors.New("no TTY available for keyring file backend password prompt")
	errInvalidKeyringBackend = errors.New("invalid keyring backend")
	errKeyringTimeout        = errors.New("keyring connection timed out")
	errEmptyAPIKey           = errors.New("API key cannot be empty")

	openKeyringFunc = openKeyring
	keyringOpenFunc = keyring.Open
)

// KeyringBackendInfo holds the resolved keyring backend value and its source.
type KeyringBackendInfo struct {
	Value  string
	Source string
}

const (
	keyringBackendSourceEnv     = "env"
	keyringBackendSourceConfig  = "config"
	keyringBackendSourceDefault = "default"
	keyringBackendAuto          = "auto"
	keyringOpenTimeout          = 5 * time.Second
)

// ResolveKeyringBackendInfo determines the keyring backend from env, config, or default.
func ResolveKeyringBackendInfo() (KeyringBackendInfo, error) {
	if v := normalizeKeyringBackend(os.Getenv(keyringBackendEnv)); v != "" {
		return KeyringBackendInfo{Value: v, Source: keyringBackendSourceEnv}, nil
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		return KeyringBackendInfo{}, fmt.Errorf("resolve keyring backend: %w", err)
	}

	if cfg.KeyringBackend != "" {
		if v := normalizeKeyringBackend(cfg.KeyringBackend); v != "" {
			return KeyringBackendInfo{Value: v, Source: keyringBackendSourceConfig}, nil
		}
	}

	return KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}, nil
}

func allowedBackends(info KeyringBackendInfo) ([]keyring.BackendType, error) {
	switch info.Value {
	case "", keyringBackendAuto:
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q (expected %s, keychain, or file)", errInvalidKeyringBackend, info.Value, keyringBackendAuto)
	}
}

func wrapKeychainError(err error) error {
	if err == nil {
		return nil
	}

	if isKeychainLockedError(err.Error()) {
		return fmt.Errorf("%w\n\nYour macOS keychain is locked. To unlock it, run:\n  security unlock-keychain ~/Library/Keychains/login.keychain-db", err)
	}

	return err
}

func isKeychainLockedError(msg string) bool {
	return strings.Contains(msg, "errSecInteractionNotAllowed") ||
		strings.Contains(msg, "The user name or passphrase you entered is not correct")
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	password := os.Getenv(keyringPasswordEnv)
	if password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

func normalizeKeyringBackend(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func shouldForceFileBackend(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr == ""
}

func shouldUseKeyringTimeout(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr != ""
}

func openKeyring() (keyring.Keyring, error) {
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	backendInfo, err := ResolveKeyringBackendInfo()
	if err != nil {
		return nil, err
	}

	backends, err := allowedBackends(backendInfo)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	if shouldForceFileBackend(runtime.GOOS, backendInfo, dbusAddr) {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:              config.AppName,
		KeychainTrustApplication: false,
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         fileKeyringPasswordFunc(),
	}

	if shouldUseKeyringTimeout(runtime.GOOS, backendInfo, dbusAddr) {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}

	ring, err := keyringOpenFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

type keyringResult struct {
	ring keyring.Keyring
	err  error
}

func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	ch := make(chan keyringResult, 1)

	go func() {
		ring, err := keyringOpenFunc(cfg)
		ch <- keyringResult{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("%w after %v (D-Bus SecretService may be unresponsive); "+
			"set STRAWPOLL_KEYRING_BACKEND=file and STRAWPOLL_KEYRING_PASSWORD=<password> to use encrypted file storage instead",
			errKeyringTimeout, timeout)
	}
}

// OpenDefault opens the default keyring store.
func OpenDefault() (Store, error) {
	ring, err := openKeyringFunc()
	if err != nil {
		return nil, err
	}

	return &KeyringStore{ring: ring}, nil
}

// SetAPIKey stores an API key in the keyring.
func (s *KeyringStore) SetAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errEmptyAPIKey
	}

	if err := s.ring.Set(keyring.Item{
		Key:  apiKeyKey,
		Data: []byte(key),
	}); err != nil {
		return wrapKeychainError(fmt.Errorf("store API key: %w", err))
	}

	return nil
}

// GetAPIKey retrieves the API key from the keyring.
func (s *KeyringStore) GetAPIKey() (string, error) {
	item, err := s.ring.Get(apiKeyKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", ErrNoAPIKey
		}

		return "", fmt.Errorf("read API key: %w", err)
	}

	return string(item.Data), nil
}

// DeleteAPIKey removes the API key from the keyring.
func (s *KeyringStore) DeleteAPIKey() error {
	if err := s.ring.Remove(apiKeyKey); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return ErrNoAPIKey
		}

		return fmt.Errorf("delete API key: %w", err)
	}

	return nil
}

// HasAPIKey checks whether an API key is stored in the keyring.
func (s *KeyringStore) HasAPIKey() (bool, error) {
	_, err := s.ring.Get(apiKeyKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("check API key: %w", err)
	}

	return true, nil
}

// GetAPIKey checks the env var first, then falls back to the keyring.
func GetAPIKey() (string, error) {
	if envKey := os.Getenv(apiKeyEnv); envKey != "" {
		return envKey, nil
	}

	store, err := OpenDefault()
	if err != nil {
		return "", err
	}

	key, err := store.GetAPIKey()
	if err != nil {
		return "", fmt.Errorf("get API key: %w", err)
	}

	return key, nil
}
