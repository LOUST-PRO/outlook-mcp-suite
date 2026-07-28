# Onboarding with a real Outlook account

This guide walks through setting up `outlook-mcp-go` against a real
Microsoft 365 / Outlook.com account. It covers app registration in
Azure, config and token storage, and the OAuth device-code flow.

For automated smoke testing against a mock Graph server, see the
`*_test.go` files under `internal/mail/`.

## 1. Register an app in Microsoft Entra (Azure AD)

You need an Azure tenant to create an app registration. Personal
Microsoft accounts work too if you pick the "personal Microsoft
account only" option during registration.

1. Go to the Azure portal: https://portal.azure.com/
2. Navigate to **Microsoft Entra ID** → **App registrations** → **New
   registration**.
3. Fill in:
   - **Name**: `outlook-mcp-go` (or whatever you prefer)
   - **Supported account types**: choose one of
     - **Accounts in this organizational directory only** (single
       tenant)
     - **Accounts in any organizational directory** (multi-tenant)
     - **Personal Microsoft accounts only** (consumer)
   - **Redirect URI**: leave empty (device-code flow does not use a
     redirect URI).
4. Click **Register**. Note the **Application (client) ID** and
   **Directory (tenant) ID** from the Overview page — you'll need both.

## 2. Configure API permissions

The OAuth scopes currently exercised by the tools are:

- `User.Read` — read the authenticated user's profile (`/me`).
- `Mail.Read` — read mail in the user's mailbox (messages, folders,
  attachments, rules, categories).

In the Azure portal, under your app registration:

1. Go to **API permissions** → **Add a permission**.
2. Choose **Microsoft Graph** → **Delegated permissions**.
3. Search for `User.Read` and `Mail.Read`, check both, and click
   **Add permissions**.
4. (Optional but recommended for tenant admins) Click **Grant admin
   consent for `<tenant>`** so end users do not see a consent prompt.

## 3. Enable public client flows (device-code)

Device-code flow is a "public client" flow in Microsoft Entra terms.
Public client flows are off by default for new app registrations.

1. In your app registration, go to **Authentication**.
2. Under **Advanced settings** → **Allow public client flows**, set
   **Enable the following mobile and desktop flows** to **Yes**.
3. Click **Save**.

Without this, the device-code flow returns `AADSTS700025` or similar.

## 4. Install and configure `outlook-mcp-go`

```bash
# Clone and build.
git clone https://github.com/LOUST-PRO/outlook-mcp-suite.git
cd outlook-mcp-suite/graph/outlook-mcp-go
go build -o outlook-mcp-go ./cmd/outlook-mcp-go

# (Optional) install to PATH.
cp outlook-mcp-go ~/.local/bin/
```

Create the config file at `~/.config/lzt-outlook/config.toml`:

```toml
[server]
path = "~/.local/share/lzt-outlook/tokens"

[[accounts]]
name         = "work"      # any alias you choose
client_id    = "<Application (client) ID from step 1>"
tenant_id    = "<Directory (tenant) ID from step 1>"
# redirect_uri is unused for device-code flow; leave commented.
# redirect_uri = ""
```

The `name` field is what tool calls reference (`account="work"`). It
does NOT need to match the email address.

## 5. Run the device-code flow

```bash
./outlook-mcp-go auth work
```

The binary will print something like:

```
To sign in, use a web browser to open the page https://microsoft.com/devicelogin
and enter the code ABCDEFGH to authenticate.
```

Open that URL, enter the code, sign in with the account you want the
MCP server to access, and approve the permissions. The token is
encrypted with `age` (scrypt-derived passphrase) and saved under the
`path` you configured.

## 6. Set up a passphrase for the token store

If you ran `auth` interactively, the binary prompts for a passphrase on
stderr and reads it from stdin. Under the MCP stdio transport, stdin
is the JSON-RPC stream — so for `serve` mode, **always pass a
passphrase file**:

```bash
# Create a 0600-mode file with the passphrase.
umask 077
echo -n 'your-passphrase-here' > ~/.config/lzt-outlook/passphrase
chmod 600 ~/.config/lzt-outlook/passphrase

# Pass it explicitly.
./outlook-mcp-go serve -passphrase-file ~/.config/lzt-outlook/passphrase
```

Or via env var (for systemd / launchd):

```ini
# ~/.config/systemd/user/outlook-mcp-go.service
[Service]
ExecStart=/home/you/.local/bin/outlook-mcp-go serve
Environment=OUTLOOK_PASSPHRASE_FILE=/home/you/.config/lzt-outlook/passphrase
```

## 7. Test the server

In another shell, with the binary running on stdio:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./outlook-mcp-go serve -passphrase-file ~/.config/lzt-outlook/passphrase
```

You should see all 11 tools registered:

- `list_messages`, `get_message`, `search_messages`
- `list_folders`, `get_folder`, `list_categories`
- `list_attachments`, `get_attachment`, `list_rules`
- `get_user_profile`, `get_mailbox_settings`

A first real call:

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_messages","arguments":{"account":"work","top":5}}}' | \
  ./outlook-mcp-go serve -passphrase-file ~/.config/lzt-outlook/passphrase
```

If everything is wired correctly, you get back a JSON envelope with
the first 5 messages from your inbox.

## 8. Wire it to an MCP host

Point your MCP client (Claude Code, Claude Desktop, etc.) at the
binary:

```json
{
  "mcpServers": {
    "outlook": {
      "command": "/home/you/.local/bin/outlook-mcp-go",
      "args": ["serve"],
      "env": {
        "OUTLOOK_PASSPHRASE_FILE": "/home/you/.config/lzt-outlook/passphrase"
      }
    }
  }
}
```

The host will spawn the binary per session; the encrypted token is
decrypted on demand using the passphrase.

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| `AADSTS700025: ... public client ...` | Public client flows not enabled (step 3). |
| `AADSTS65001: user or admin has not consented` | Tenant requires admin consent for `Mail.Read`. Either grant admin consent in step 2 or have the user consent interactively (the device-code flow does this). |
| `account "work" not in config` | `name` in `[accounts]` block doesn't match the `account` argument. |
| `no token for account "work"` | You skipped step 5 (device-code login). |
| HTTP 401 from Graph | The encrypted token could not be decrypted with the passphrase (wrong passphrase, or passphrase file changed). |
| HTTP 403 from Graph | Token is valid but the account lacks the requested permission (e.g. `Mail.Read` was not granted). |
| Empty `value: []` from `list_messages` | Inbox is genuinely empty, OR the folder ID is wrong (e.g. `Inbox` vs `inbox`). Folder names are case-insensitive per Graph spec, but be careful. |

## Security notes

- The encrypted token file uses `filippo.io/age` with scrypt-derived
  keys. Treat the passphrase file (mode 0600) with the same care as a
  private SSH key.
- The token cache is per-account on disk; do NOT symlink it across
  machines.
- If you rotate the tenant secret or app registration, you must
  re-run `auth <account>` — old tokens become invalid.
- `get_attachment` defaults to a 10 MiB cap on attachment size. Adjust
  with `max_bytes` if you need a different limit; the absolute hard
  cap is 25 MiB.

## What this guide does NOT cover

- Setting up multi-tenant app registrations (requires admin
  verification in the Azure partner center).
- Using managed-identity / service-principal flows (the server is
  designed for interactive user auth).
- Migration from a previous token store format — there is no
  migration yet because this is the first release.