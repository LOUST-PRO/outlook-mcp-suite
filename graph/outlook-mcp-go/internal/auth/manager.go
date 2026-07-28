// manager.go — high-level "give me a bearer token for account X".
//
// Glues together: Config (which app to use) + TokenStore (encrypted
// persistence) + Authenticator (MSAL client). This is what the MCP
// tool handlers call into.

package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// PassphraseFunc returns the encryption passphrase. The default impl
// prompts on stderr and reads from stdin. Tests inject a constant.
// IMPORTANT: under stdio MCP transport, stdin is the JSON-RPC stream,
// so the MCP server MUST inject a func that prompts the user BEFORE
// calling server.Listen() (e.g. in main, before the server starts).
type PassphraseFunc func() (string, error)

// DefaultPassphraseFunc prompts on stderr, reads from stdin. Not safe
// under stdio MCP transport — see PassphraseFunc doc.
func DefaultPassphraseFunc() (string, error) {
	fmt.Fprintln(os.Stderr, "Passphrase for outlook-mcp-go token store:")
	fmt.Fprint(os.Stderr, "> ")
	var pw string
	if _, err := fmt.Scanln(&pw); err != nil {
		return "", fmt.Errorf("auth: read passphrase: %w", err)
	}
	if pw == "" {
		return "", errors.New("auth: empty passphrase")
	}
	return pw, nil
}

// StaticPassphraseFunc returns a constant passphrase. Useful for tests
// and for callers that captured the passphrase from an earlier prompt.
func StaticPassphraseFunc(pw string) PassphraseFunc {
	return func() (string, error) { return pw, nil }
}

// Manager coordinates config + token store + MSAL client.
type Manager struct {
	Config     *Config
	TokenStore *Store
	Passphrase PassphraseFunc
}

// NewManager wires the three pieces together. PassphraseFunc may be
// nil to fall back to DefaultPassphraseFunc.
func NewManager(cfg *Config, store *Store, passphrase PassphraseFunc) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("auth: nil config")
	}
	if store == nil {
		return nil, errors.New("auth: nil token store")
	}
	if passphrase == nil {
		passphrase = DefaultPassphraseFunc
	}
	return &Manager{Config: cfg, TokenStore: store, Passphrase: passphrase}, nil
}

// Login runs the device-code flow for account and persists the token
// cache encrypted on disk. Called from the `auth <account>` CLI sub.
func (m *Manager) Login(ctx context.Context, account string, ui *DeviceCodeUI) (AccountEntry, error) {
	entry, ok := m.Config.Find(account)
	if !ok {
		return AccountEntry{}, fmt.Errorf("auth: account %q not in config (allowed: %v)", account, m.Config.Allowed())
	}

	auth, err := NewAuthenticator(entry.ClientID, entry.TenantID)
	if err != nil {
		return AccountEntry{}, err
	}
	if _, err := auth.AcquireTokenByDeviceCode(ctx, []string{"User.Read", "Mail.Read"}, ui); err != nil {
		return AccountEntry{}, err
	}
	// Note: with the cache-less flow above, MSAL does NOT persist the
	// token. For Fase 1 we accept this — user re-auths each session.
	// Fase 2 will attach a FileCacheAccessor at Login time so the
	// resulting token survives process exit.
	return entry, nil
}

// BearerFor returns an OAuth access token for account. Loads the
// encrypted cache, hydrates MSAL, acquires silently with refresh.
// If no cached token exists, returns an error directing the caller
// to run Login first.
func (m *Manager) BearerFor(ctx context.Context, account string) (string, error) {
	entry, ok := m.Config.Find(account)
	if !ok {
		return "", fmt.Errorf("auth: account %q not in config", account)
	}
	if !m.Config.IsAllowed(account) {
		return "", fmt.Errorf("auth: account %q not in [accounts].allowed (default-deny)", account)
	}
	if !m.TokenStore.Exists(account) {
		return "", fmt.Errorf("auth: no token for account %q (run 'outlook-mcp-go auth %s')", account, account)
	}

	passphrase, err := m.Passphrase()
	if err != nil {
		return "", err
	}

	// Build a one-shot MSAL client with a FileCacheAccessor so that
	// the encrypted token persists transparently across this call
	// (and any subsequent refresh during it).
	accessor := &FileCacheAccessor{
		Account:    account,
		Passphrase: passphrase,
		Store:      m.TokenStore,
	}
	client, err := public.New(entry.ClientID,
		public.WithAuthority(AuthorityURL(entry.TenantID)),
		public.WithCache(accessor),
	)
	if err != nil {
		return "", fmt.Errorf("auth: MSAL client with cache: %w", err)
	}

	accounts, err := client.Accounts(ctx)
	if err != nil {
		return "", fmt.Errorf("auth: list MSAL accounts: %w", err)
	}
	if len(accounts) == 0 {
		return "", fmt.Errorf("auth: no MSAL accounts hydrated for %q (run 'outlook-mcp-go auth %s')", account, account)
	}

	// 30s ceiling covers transient MSAL network blips without blocking
	// the MCP request indefinitely.
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	result, err := client.AcquireTokenSilent(ctx2, []string{"User.Read", "Mail.Read"}, public.WithSilentAccount(accounts[0]))
	if err != nil {
		return "", fmt.Errorf("auth: silent acquire: %w", err)
	}
	if result.AccessToken == "" {
		return "", errors.New("auth: silent acquire returned empty access token")
	}
	return result.AccessToken, nil
}

// Ensure cache.ExportReplace is satisfied at compile time.
var _ cache.ExportReplace = (*FileCacheAccessor)(nil)
