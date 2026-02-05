package auth

import (
	"errors"
	"testing"

	"github.com/99designs/keyring"
)

// mockKeyring implements keyring.Keyring for testing.
type mockKeyring struct {
	items map[string]keyring.Item
}

func newMockKeyring() *mockKeyring {
	return &mockKeyring{items: make(map[string]keyring.Item)}
}

func (m *mockKeyring) Get(key string) (keyring.Item, error) {
	item, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}

	return item, nil
}

func (m *mockKeyring) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func (m *mockKeyring) Set(item keyring.Item) error {
	m.items[item.Key] = item

	return nil
}

func (m *mockKeyring) Remove(key string) error {
	if _, ok := m.items[key]; !ok {
		return keyring.ErrKeyNotFound
	}

	delete(m.items, key)

	return nil
}

func (m *mockKeyring) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}

	return keys, nil
}

func TestGetAPIKeyEnvVar(t *testing.T) {
	t.Setenv("STRAWPOLL_API_KEY", "test-key-from-env")

	// Override openKeyringFunc to fail if called -- env should short-circuit
	origOpen := openKeyringFunc
	openKeyringFunc = func() (keyring.Keyring, error) {
		t.Fatal("keyring should not be opened when env var is set")

		return nil, nil
	}

	t.Cleanup(func() { openKeyringFunc = origOpen })

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey() error: %v", err)
	}

	if key != "test-key-from-env" {
		t.Errorf("GetAPIKey() = %q, want %q", key, "test-key-from-env")
	}
}

func TestShouldForceFileBackend(t *testing.T) {
	info := KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}

	if !shouldForceFileBackend("linux", info, "") {
		t.Error("shouldForceFileBackend(linux, auto, no-dbus) = false, want true")
	}

	if shouldForceFileBackend("darwin", info, "") {
		t.Error("shouldForceFileBackend(darwin, auto, no-dbus) = true, want false")
	}

	if shouldForceFileBackend("linux", info, "/run/user/1000/bus") {
		t.Error("shouldForceFileBackend(linux, auto, dbus-set) = true, want false")
	}
}

func TestShouldUseKeyringTimeout(t *testing.T) {
	info := KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}

	if !shouldUseKeyringTimeout("linux", info, "/run/user/1000/bus") {
		t.Error("shouldUseKeyringTimeout(linux, auto, dbus-set) = false, want true")
	}

	if shouldUseKeyringTimeout("darwin", info, "/run/user/1000/bus") {
		t.Error("shouldUseKeyringTimeout(darwin, auto, dbus-set) = true, want false")
	}

	if shouldUseKeyringTimeout("linux", info, "") {
		t.Error("shouldUseKeyringTimeout(linux, auto, no-dbus) = true, want false")
	}
}

func TestKeyringStoreSetGetDelete(t *testing.T) {
	mock := newMockKeyring()
	store := &KeyringStore{ring: mock}

	// Set
	if err := store.SetAPIKey("my-secret-key"); err != nil {
		t.Fatalf("SetAPIKey() error: %v", err)
	}

	// Get
	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey() error: %v", err)
	}

	if key != "my-secret-key" {
		t.Errorf("GetAPIKey() = %q, want %q", key, "my-secret-key")
	}

	// Has
	has, err := store.HasAPIKey()
	if err != nil {
		t.Fatalf("HasAPIKey() error: %v", err)
	}

	if !has {
		t.Error("HasAPIKey() = false, want true")
	}

	// Delete
	if err := store.DeleteAPIKey(); err != nil {
		t.Fatalf("DeleteAPIKey() error: %v", err)
	}

	// Verify deleted
	has, err = store.HasAPIKey()
	if err != nil {
		t.Fatalf("HasAPIKey() after delete error: %v", err)
	}

	if has {
		t.Error("HasAPIKey() after delete = true, want false")
	}

	// Get after delete returns ErrNoAPIKey
	_, err = store.GetAPIKey()
	if !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("GetAPIKey() after delete error = %v, want ErrNoAPIKey", err)
	}
}

func TestEmptyAPIKeyRejected(t *testing.T) {
	mock := newMockKeyring()
	store := &KeyringStore{ring: mock}

	if err := store.SetAPIKey(""); !errors.Is(err, errEmptyAPIKey) {
		t.Errorf("SetAPIKey(\"\") error = %v, want errEmptyAPIKey", err)
	}

	if err := store.SetAPIKey("   "); !errors.Is(err, errEmptyAPIKey) {
		t.Errorf("SetAPIKey(\"   \") error = %v, want errEmptyAPIKey", err)
	}
}
