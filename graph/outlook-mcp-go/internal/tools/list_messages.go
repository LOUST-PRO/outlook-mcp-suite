// list_messages.go — MCP tool handler for list_messages.
//
// Reads `account` (required) + optional filters from the MCP request,
// fetches a bearer token from the auth.Manager, calls Graph, and
// returns a JSON-formatted text response.

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/auth"
	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/mail"
)

// HandleListMessages returns an mcp-go tool handler bound to the
// given auth.Manager and mail.Client. The handler is the canonical
// pattern for the other 10 Fase 1 tools (Stage 2+).
func HandleListMessages(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("list_messages: 'account' argument is required"), nil
		}

		// Graph access requires a valid bearer token. Manager enforces
		// default-deny: account must be in [accounts].allowed AND a
		// token must already exist on disk (login pre-required).
		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_messages: auth: %s", err)), nil
		}

		// Parse optional arguments. mcp-go exposes typed accessors
		// (GetString, GetInt, ...) that default to "" / 0 when unset.
		opts := mail.ListMessagesOptions{
			FolderID: req.GetString("folder", ""),
			Filter:   req.GetString("filter", ""),
			Search:   req.GetString("search", ""),
			OrderBy:  req.GetString("orderby", ""),
			Top:      int(req.GetInt("top", 25)),
		}
		// Default orderby for predictable UX.
		if opts.OrderBy == "" {
			opts.OrderBy = "receivedDateTime desc"
		}

		resp, err := graph.ListMessages(ctx, bearer, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_messages: Graph: %s", err)), nil
		}

		// Serialize the full OData envelope. The MCP client (Claude
		// Code, etc.) gets a stable JSON shape it can parse.
		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_messages: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}
