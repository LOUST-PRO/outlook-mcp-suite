// main.go — outlook-mcp-go entry point.
//
// Subcommands:
//
//   auth <account>   Run OAuth device-code flow, save encrypted token.
//   serve            Start MCP stdio server (default if no subcommand).
//
// Flags (env-var equivalents in parens):
//
//   -config PATH     (OUTLOOK_CONFIG)      Config file path.
//   -token-dir DIR   (OUTLOOK_TOKEN_DIR)   Encrypted token directory.
//   -passphrase-file PATH                  Read passphrase from file instead of stdin.
//
// In MCP server mode (subcommand `serve` or default), stdout is the
// JSON-RPC stream; all operator-facing output goes to stderr.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/auth"
	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/server"
)

const usage = `outlook-mcp-go — MCP server for Outlook/Microsoft 365 (Path A: Graph API)

Usage:
  outlook-mcp-go auth <account>     Run OAuth device-code flow and save token.
  outlook-mcp-go serve              Start MCP stdio server (default).
  outlook-mcp-go -h | -help         Show this help.
  outlook-mcp-go -version           Show version.

Flags (env-var equivalents in parens):
  -config PATH         (OUTLOOK_CONFIG)      Config file path.
                                            Default: ~/.config/lzt-outlook/config.toml
  -token-dir DIR       (OUTLOOK_TOKEN_DIR)   Encrypted token directory.
                                            Default: ~/.local/share/lzt-outlook/tokens
  -passphrase-file PATH                      Read passphrase from file instead of stdin.
                                            Useful for systemd / launchd units.

Exit codes:
  0   Success
  1   Runtime error (auth failed, Graph returned 4xx/5xx, ...)
  2   Misconfiguration (config missing, account not in allowlist, ...)

Fase 1 (this build): read-only, 1 tool (list_messages) registered.
Fase 2 will add send/mutate tools + shadow-mode previews.
`

func main() {
	// Subcommand routing
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "auth":
			os.Exit(runAuth(os.Args[2:]))
		case "-h", "-help", "--help", "help":
			fmt.Fprint(os.Stderr, usage)
			os.Exit(0)
		case "-version", "--version", "version":
			fmt.Fprintln(os.Stderr, "outlook-mcp-go 0.1.0-stage4")
			os.Exit(0)
		case "serve":
			os.Exit(runServe(os.Args[2:]))
		}
	}
	// Default: serve (stdio MCP)
	os.Exit(runServe(os.Args[1:]))
}

func runAuth(args []string) int {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	configPath := fs.String("config", auth.DefaultConfigPath(), "Config file path (OUTLOOK_CONFIG)")
	tokenDir := fs.String("token-dir", auth.DefaultStorePath(), "Token directory (OUTLOOK_TOKEN_DIR)")
	passphraseFile := fs.String("passphrase-file", "", "Read passphrase from file (0600)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "auth: account name required (see `outlook-mcp-go auth <account>`)")
		return 2
	}
	account := fs.Arg(0)

	cfg, err := auth.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth: config: %s\n", err)
		return 2
	}
	if !cfg.IsAllowed(account) {
		fmt.Fprintf(os.Stderr, "auth: account %q is not in [accounts].allowed (default-deny). Add it to %s.\n", account, *configPath)
		return 2
	}

	passphraseFn, err := resolvePassphrase(*passphraseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth: passphrase: %s\n", err)
		return 2
	}

	store := auth.NewStore(*tokenDir)
	mgr, err := auth.NewManager(cfg, store, passphraseFn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth: manager: %s\n", err)
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	entry, err := mgr.Login(ctx, account, auth.NewDeviceCodeUI())
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth: login: %s\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "auth: %s authenticated (client_id=%s, tenant=%s). Token encrypted at %s/<account>.age.\n",
		account, entry.ClientID, entry.TenantID, *tokenDir)
	return 0
}

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	configPath := fs.String("config", auth.DefaultConfigPath(), "Config file path (OUTLOOK_CONFIG)")
	tokenDir := fs.String("token-dir", auth.DefaultStorePath(), "Token directory (OUTLOOK_TOKEN_DIR)")
	passphraseFile := fs.String("passphrase-file", "", "Read passphrase from file (0600)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := auth.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: config: %s\n", err)
		return 2
	}

	passphraseFn, err := resolvePassphrase(*passphraseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: passphrase: %s\n", err)
		return 2
	}

	store := auth.NewStore(*tokenDir)
	mgr, err := auth.NewManager(cfg, store, passphraseFn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: manager: %s\n", err)
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := server.ServeStdio(ctx, mgr); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "serve: %s\n", err)
		return 1
	}
	return 0
}

// resolvePassphrase picks a PassphraseFunc based on flags / env.
// -passphrase-file PATH → reads first line, trims newline.
// (no flag) → DefaultPassphraseFunc (prompts on stderr, reads stdin).
func resolvePassphrase(filePath string) (auth.PassphraseFunc, error) {
	if filePath == "" {
		if v := os.Getenv("OUTLOOK_PASSPHRASE_FILE"); v != "" {
			filePath = v
		}
	}
	if filePath == "" {
		return auth.DefaultPassphraseFunc, nil
	}
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", filePath, err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w (set mode 0600 to avoid accidental world-readable)", abs, err)
	}
	pw := strings.TrimRight(string(data), "\r\n")
	if pw == "" {
		return nil, errors.New("empty passphrase in file")
	}
	return auth.StaticPassphraseFunc(pw), nil
}
