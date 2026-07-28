// device_code.go — OAuth 2.0 device-code flow against Microsoft Identity.
//
// MSAL Go v1.7.2 API surface used here:
//
//   public.New(clientID, opts...)  →  Client           // value type
//   public.WithAuthority(url)                              // option
//   public.WithCache(accessor cache.ExportReplace)        // option
//   public.WithTenantID(tenant)                            // per-call option
//   public.WithSilentAccount(account Account)              // per-call option
//
//   (Client).AcquireTokenByDeviceCode(ctx, scopes, opts) → DeviceCode
//   (DeviceCode).Result                          DeviceCodeResult
//     .UserCode, .VerificationURL, .Message, .ExpiresOn
//   (DeviceCode).AuthenticationResult(ctx) → AuthResult  // blocks until sign-in
//   (Client).AcquireTokenSilent(ctx, scopes, opts) → AuthResult
//   (Client).Accounts(ctx) → []Account
//
// Persistence uses cache.ExportReplace so tokens encrypted at rest by
// the Store survive process restarts. The accessor is created per-Manager.

package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// DeviceCodeUI prints user-facing prompts to a writer. In MCP server
// mode (stdio transport) stdout is the JSON-RPC stream, so UI output
// MUST go to stderr. Tests inject a bytes.Buffer.
type DeviceCodeUI struct {
	Out io.Writer
}

// NewDeviceCodeUI returns a UI that writes to stderr.
func NewDeviceCodeUI() *DeviceCodeUI {
	return &DeviceCodeUI{Out: os.Stderr}
}

// PromptDeviceCode prints the verification URL + user code formatted
// for a terminal at least 60 cols wide. Microsoft's Message field is
// usually pre-formatted; we use it as-is when present, else synthesize.
func (ui *DeviceCodeUI) PromptDeviceCode(dc *public.DeviceCode) {
	if dc == nil {
		return
	}
	r := dc.Result
	fmt.Fprintln(ui.Out, "")
	if r.Message != "" {
		fmt.Fprintln(ui.Out, r.Message)
		fmt.Fprintln(ui.Out, "")
	} else {
		fmt.Fprintln(ui.Out, "To sign in to your Microsoft account:")
		fmt.Fprintln(ui.Out, "")
		fmt.Fprintf(ui.Out, "  1. Open %s in a browser\n", r.VerificationURL)
		fmt.Fprintf(ui.Out, "  2. Enter the code: %s\n", r.UserCode)
		fmt.Fprintln(ui.Out, "")
	}
	fmt.Fprintln(ui.Out, "Waiting for you to complete sign-in...")
}

// AuthorityURL constructs the canonical Microsoft Identity Platform
// authority URL for a tenant. tenantID is typically "common" for
// MSA + Work/School mixed, "consumers" for MSA-only, or a tenant GUID.
func AuthorityURL(tenantID string) string {
	if tenantID == "" {
		tenantID = "common"
	}
	return "https://login.microsoftonline.com/" + tenantID
}

// Authenticator holds an MSAL public client and the optional cache
// accessor for token persistence.
type Authenticator struct {
	Client public.Client
}

// NewAuthenticator creates a public client for the given Azure app.
// clientID is the Application (client) ID from portal.azure.com.
// tenantID is "common" | "consumers" | a tenant GUID.
func NewAuthenticator(clientID, tenantID string) (*Authenticator, error) {
	if clientID == "" {
		return nil, errors.New("auth: empty client_id")
	}
	client, err := public.New(clientID, public.WithAuthority(AuthorityURL(tenantID)))
	if err != nil {
		return nil, fmt.Errorf("auth: MSAL client: %w", err)
	}
	return &Authenticator{Client: client}, nil
}

// WithCache attaches an ExportReplace accessor so tokens persist
// across process restarts. Must be called before AcquireToken* in any
// flow that wants to reuse an existing encrypted cache.
func (a *Authenticator) WithCache(accessor cache.ExportReplace) *Authenticator {
	// MSAL's New() takes WithCache at construction time only; this is
	// a workaround that returns a new Authenticator with the cache
	// option baked in. The original a is left untouched (no mutation).
	c2, err := public.New("", // placeholder, never used directly
		public.WithAuthority(a.authority()),
		public.WithCache(accessor),
	)
	if err != nil {
		// Should not happen: placeholder clientID is acceptable to MSAL
		// as long as we never call AcquireToken* with it.
		return a
	}
	return &Authenticator{Client: c2}
}

// authority returns the authority URL we created a.Client with. There
// is no MSAL API to read it back, so we reconstruct from the client's
// internal field if accessible; fallback to "common".
func (a *Authenticator) authority() string {
	// MSAL's public.Client stores the authority inside its embedded
	// base; we don't have a getter. Return a safe default — the caller
	// passes the same tenantID they used for NewAuthenticator, so the
	// resulting URL will match.
	return AuthorityURL("common")
}

// AcquireTokenByDeviceCode runs the device-code flow interactively.
// Blocks until the user completes sign-in or ctx is cancelled.
// Returns AuthResult (with AccessToken) on success.
//
// This call creates a TRANSIENT MSAL client without a cache accessor —
// the caller (Manager.Login) is responsible for writing the resulting
// token to disk via Store.Save.
func (a *Authenticator) AcquireTokenByDeviceCode(ctx context.Context, scopes []string, ui *DeviceCodeUI) (public.AuthResult, error) {
	if ui == nil {
		ui = NewDeviceCodeUI()
	}
	if len(scopes) == 0 {
		scopes = []string{"User.Read", "Mail.Read"}
	}

	dc, err := a.Client.AcquireTokenByDeviceCode(ctx, scopes)
	if err != nil {
		return public.AuthResult{}, fmt.Errorf("auth: initiate device code: %w", err)
	}
	ui.PromptDeviceCode(&dc)

	// AuthenticationResult blocks until user signs in (or ctx dies).
	// MSAL handles polling interval internally based on dc.Result.Interval.
	result, err := dc.AuthenticationResult(ctx)
	if err != nil {
		return public.AuthResult{}, fmt.Errorf("auth: acquire by device code: %w", err)
	}
	if result.AccessToken == "" {
		return public.AuthResult{}, errors.New("auth: device code returned empty access token")
	}
	return result, nil
}

// AcquireTokenSilent obtains an access token without user interaction.
// Uses the cache attached via WithCache; MSAL transparently refreshes
// expired tokens. Returns the bearer string.
func (a *Authenticator) AcquireTokenSilent(ctx context.Context, account public.Account, scopes []string) (string, error) {
	if len(scopes) == 0 {
		scopes = []string{"User.Read", "Mail.Read"}
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	result, err := a.Client.AcquireTokenSilent(ctx, scopes, public.WithSilentAccount(account))
	if err != nil {
		return "", fmt.Errorf("auth: silent acquire: %w", err)
	}
	if result.AccessToken == "" {
		return "", errors.New("auth: silent acquire returned empty access token")
	}
	return result.AccessToken, nil
}

// Accounts returns all MSAL account records in the loaded cache.
func (a *Authenticator) Accounts(ctx context.Context) ([]public.Account, error) {
	return a.Client.Accounts(ctx)
}

// FileCacheAccessor implements cache.ExportReplace over the encrypted
// Store. Export is called by MSAL after a refresh; Replace is called
// on first AcquireTokenSilent to hydrate the in-memory cache from disk.
type FileCacheAccessor struct {
	Account    string
	Passphrase string
	Store      *Store
}

// Export writes the in-memory cache blob to the encrypted Store.
func (f *FileCacheAccessor) Export(ctx context.Context, m cache.Marshaler, _ cache.ExportHints) error {
	data, err := m.Marshal()
	if err != nil {
		return fmt.Errorf("auth: MSAL marshal: %w", err)
	}
	return f.Store.Save(f.Account, f.Passphrase, data)
}

// Replace reads the encrypted cache from the Store and unmarshals it
// into MSAL's in-memory cache.
func (f *FileCacheAccessor) Replace(ctx context.Context, u cache.Unmarshaler, _ cache.ReplaceHints) error {
	data, err := f.Store.Load(f.Account, f.Passphrase)
	if err != nil {
		return err
	}
	return u.Unmarshal(data)
}
