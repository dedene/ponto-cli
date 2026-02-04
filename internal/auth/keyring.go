package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/dedene/ponto-cli/internal/config"
)

const (
	keyringPasswordEnv = "PONTO_KEYRING_PASSWORD"
	keyringBackendEnv  = "PONTO_KEYRING_BACKEND"
	keyringBackendAuto = "auto"
)

var (
	errMissingProfile  = errors.New("missing profile")
	errNoTTY           = errors.New("no TTY available for keyring password prompt")
	errInvalidBackend  = errors.New("invalid keyring backend")
	errKeyringTimeout  = errors.New("keyring connection timed out")
	keyringOpenTimeout = 5 * time.Second
)

// Store provides credential storage.
type Store interface {
	GetCredentials(profile string) (clientID, clientSecret string, err error)
	SetCredentials(profile, clientID, clientSecret string) error
	DeleteCredentials(profile string) error
}

// KeyringStore stores credentials in the OS keyring.
type KeyringStore struct {
	ring keyring.Keyring
}

// OpenKeyring opens the default keyring store.
func OpenKeyring() (Store, error) {
	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}

	return &KeyringStore{ring: ring}, nil
}

func openKeyring() (keyring.Keyring, error) {
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	backendInfo := resolveKeyringBackend()

	backends, err := allowedBackends(backendInfo)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	// On Linux with "auto" backend and no D-Bus, force file backend
	if runtime.GOOS == "linux" && backendInfo == keyringBackendAuto && dbusAddr == "" {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:              config.AppName,
		KeychainTrustApplication: true,
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         fileKeyringPasswordFunc(),
	}

	// Use timeout on Linux with D-Bus
	if runtime.GOOS == "linux" && backendInfo == keyringBackendAuto && dbusAddr != "" {
		return openKeyringWithTimeout(cfg)
	}

	ring, err := keyring.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

func openKeyringWithTimeout(cfg keyring.Config) (keyring.Keyring, error) {
	type result struct {
		ring keyring.Keyring
		err  error
	}

	ch := make(chan result, 1)

	go func() {
		ring, err := keyring.Open(cfg)
		ch <- result{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(keyringOpenTimeout):
		return nil, fmt.Errorf("%w; set %s=file and %s=<password>",
			errKeyringTimeout, keyringBackendEnv, keyringPasswordEnv)
	}
}

func resolveKeyringBackend() string {
	if v := os.Getenv(keyringBackendEnv); v != "" {
		return strings.ToLower(strings.TrimSpace(v))
	}

	cfg, err := config.ReadConfig()
	if err == nil && cfg.KeyringBackend != "" {
		return strings.ToLower(strings.TrimSpace(cfg.KeyringBackend))
	}

	return keyringBackendAuto
}

func allowedBackends(backend string) ([]keyring.BackendType, error) {
	switch backend {
	case "", keyringBackendAuto:
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidBackend, backend)
	}
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	if password := os.Getenv(keyringPasswordEnv); password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

// credentials is stored as JSON in a single keyring item.
type credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func credentialsKey(profile string) string {
	return fmt.Sprintf("ponto:%s:credentials", profile)
}

// Legacy keys for migration.
func legacyClientIDKey(profile string) string {
	return fmt.Sprintf("ponto:%s:client_id", profile)
}

func legacyClientSecretKey(profile string) string {
	return fmt.Sprintf("ponto:%s:client_secret", profile)
}

// GetCredentials retrieves credentials from the keyring.
func (s *KeyringStore) GetCredentials(profile string) (string, string, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return "", "", errMissingProfile
	}

	// Try new unified key first
	item, err := s.ring.Get(credentialsKey(profile))
	if err == nil {
		var creds credentials
		if jsonErr := json.Unmarshal(item.Data, &creds); jsonErr == nil {
			return creds.ClientID, creds.ClientSecret, nil
		}
	}

	// Fall back to legacy separate keys and migrate
	idItem, err := s.ring.Get(legacyClientIDKey(profile))
	if err != nil {
		return "", "", fmt.Errorf("get credentials: %w", err)
	}

	secretItem, err := s.ring.Get(legacyClientSecretKey(profile))
	if err != nil {
		return "", "", fmt.Errorf("get credentials: %w", err)
	}

	clientID := string(idItem.Data)
	clientSecret := string(secretItem.Data)

	// Migrate to new format
	_ = s.SetCredentials(profile, clientID, clientSecret)

	// Clean up legacy keys
	_ = s.ring.Remove(legacyClientIDKey(profile))
	_ = s.ring.Remove(legacyClientSecretKey(profile))

	return clientID, clientSecret, nil
}

// SetCredentials stores credentials in the keyring.
func (s *KeyringStore) SetCredentials(profile, clientID, clientSecret string) error {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return errMissingProfile
	}

	creds := credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  credentialsKey(profile),
		Data: data,
	}); err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	return nil
}

// DeleteCredentials removes credentials from the keyring.
func (s *KeyringStore) DeleteCredentials(profile string) error {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return errMissingProfile
	}

	// Remove new key
	if err := s.ring.Remove(credentialsKey(profile)); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("delete credentials: %w", err)
	}

	// Also clean up legacy keys if they exist
	_ = s.ring.Remove(legacyClientIDKey(profile))
	_ = s.ring.Remove(legacyClientSecretKey(profile))

	return nil
}
