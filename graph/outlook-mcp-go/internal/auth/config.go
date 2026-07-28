// config.go — load + parse ~/.config/lzt-outlook/config.toml.
//
// Default-deny semantics match lib-outlook-allowlist.sh: an empty or
// missing [accounts].allowed list blocks ALL account usage.

package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

// Config is the top-level structure of ~/.config/lzt-outlook/config.toml.
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Accounts []AccountEntry `toml:"accounts"`
}

// ServerConfig identifies which path the MCP binary implements.
// "graph" → Microsoft Graph API (this binary). "playwright" → web.outlook.com
// automation (separate binary). See ../ARCHITECTURE.md.
type ServerConfig struct {
	Path string `toml:"path"`
}

// AccountEntry describes one Microsoft account the user has registered
// an Azure app for. client_id is the Application (client) ID from
// portal.azure.com > App registrations > [app] > Overview.
// tenant_id is typically "common" for MSA + Work/School mixed flows.
type AccountEntry struct {
	Name        string `toml:"name"`
	ClientID    string `toml:"client_id"`
	TenantID    string `toml:"tenant_id"`
	RedirectURI string `toml:"redirect_uri"`
}

// Allowed returns the list of account names that the user has explicitly
// opted in. Empty slice means default-deny (no accounts allowed).
func (c *Config) Allowed() []string {
	out := make([]string, 0, len(c.Accounts))
	for _, a := range c.Accounts {
		if a.Name != "" {
			out = append(out, a.Name)
		}
	}
	return out
}

// IsAllowed reports whether account is in the explicit allowlist.
// Empty config or empty allowlist returns false (default-deny).
func (c *Config) IsAllowed(account string) bool {
	if account == "" {
		return false
	}
	return slices.Contains(c.Allowed(), account)
}

// Find returns the AccountEntry for a given account name, or false.
func (c *Config) Find(account string) (AccountEntry, bool) {
	for _, a := range c.Accounts {
		if a.Name == account {
			return a, true
		}
	}
	return AccountEntry{}, false
}

// DefaultConfigPath returns the canonical location of the config file,
// honoring $OUTLOOK_CONFIG for testing.
func DefaultConfigPath() string {
	if p := os.Getenv("OUTLOOK_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lzt-outlook", "config.toml")
}

// Load reads + parses a TOML config from path. Returns an error if the
// file is missing, unreadable, or malformed. An empty file is valid
// (yields a Config with no accounts → default-deny).
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("auth: empty config path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("auth: config not found at %s (run 'outlook-mcp-go init' or create manually)", path)
		}
		return nil, fmt.Errorf("auth: read config: %w", err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("auth: parse config: %w", err)
	}
	if c.Server.Path == "" {
		c.Server.Path = "graph"
	}
	return &c, nil
}
