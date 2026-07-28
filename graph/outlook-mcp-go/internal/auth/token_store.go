// token_store.go — age-encrypted MSAL token cache on disk.
//
// File layout:
//   <dir>/<account>.age   ← age-encrypted JSON blob (mode 0600)
//   <account> is sanitized to [_A-Za-z0-9-] for filesystem safety.
//
// Encryption is passphrase-based (X25519 + Scrypt). The passphrase is
// prompted at startup, never persisted. There is no keyring integration
// in Phase 1 — Fase 2 may add libsecret via zalando/go-keyring.
//
// Why age: simple API, no libsecret/gnome-keyring/CGo dependency,
// passphrase ergonomics match user expectations for personal tooling.

package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// Store persists per-account MSAL token caches, encrypted with age.
type Store struct {
	Dir string
}

// NewStore returns a Store rooted at dir. Dir is created (mode 0700) on
// first Save if it does not exist.
func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

// DefaultStorePath returns the canonical directory for token storage,
// honoring $OUTLOOK_TOKEN_DIR for testing.
func DefaultStorePath() string {
	if d := os.Getenv("OUTLOOK_TOKEN_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "lzt-outlook", "tokens")
}

// pathFor returns the on-disk path for a given account, after sanitizing.
func (s *Store) pathFor(account string) (string, error) {
	if account == "" {
		return "", errors.New("auth: empty account name")
	}
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-' || r == '_':
			return r
		default:
			return '_'
		}
	}, account)
	return filepath.Join(s.Dir, cleaned+".age"), nil
}

// Exists reports whether an encrypted token blob exists for account.
func (s *Store) Exists(account string) bool {
	p, err := s.pathFor(account)
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

// Save encrypts plaintext with passphrase (using age's scrypt recipient)
// and writes it atomically to <dir>/<account>.age (mode 0600).
func (s *Store) Save(account, passphrase string, plaintext []byte) error {
	if passphrase == "" {
		return errors.New("auth: empty passphrase")
	}
	p, err := s.pathFor(account)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return fmt.Errorf("auth: mkdir token dir: %w", err)
	}

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return fmt.Errorf("auth: new scrypt recipient: %w", err)
	}

	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)
	encWriter, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		return fmt.Errorf("auth: age encrypt: %w", err)
	}
	if _, err := encWriter.Write(plaintext); err != nil {
		return fmt.Errorf("auth: age write: %w", err)
	}
	if err := encWriter.Close(); err != nil {
		return fmt.Errorf("auth: age close: %w", err)
	}
	if err := armorWriter.Close(); err != nil {
		return fmt.Errorf("auth: armor close: %w", err)
	}

	// Atomic write: tmp + rename. Avoids partial writes if interrupted.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("auth: write tmp: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("auth: rename: %w", err)
	}
	return nil
}

// Load reads + decrypts the token blob for account. Returns an error if
// the file is missing, the passphrase is wrong, or the file is corrupt.
func (s *Store) Load(account, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("auth: empty passphrase")
	}
	p, err := s.pathFor(account)
	if err != nil {
		return nil, err
	}
	ciphertext, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("auth: no token for account %q (run 'outlook-mcp-go auth %s' to authenticate)", account, account)
		}
		return nil, fmt.Errorf("auth: read token: %w", err)
	}

	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("auth: new scrypt identity: %w", err)
	}

	armorReader := armor.NewReader(bytes.NewReader(ciphertext))
	decReader, err := age.Decrypt(armorReader, identity)
	if err != nil {
		return nil, fmt.Errorf("auth: decrypt (wrong passphrase?): %w", err)
	}
	var out bytes.Buffer
	if _, err := out.ReadFrom(decReader); err != nil {
		return nil, fmt.Errorf("auth: read decrypted: %w", err)
	}
	return out.Bytes(), nil
}

// Delete removes the token blob for account. No-op if it does not exist.
func (s *Store) Delete(account string) error {
	p, err := s.pathFor(account)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
