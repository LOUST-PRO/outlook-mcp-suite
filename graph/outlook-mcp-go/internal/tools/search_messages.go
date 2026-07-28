// search_messages.go — MCP tool handler for search_messages.
//
// Thin wrapper around list_messages with the $search parameter forced.
// Results are ranked by Graph's search relevance; sort order is
// receivedDateTime desc as a stable fallback. Supports folder scoping
// to limit search to a specific mailbox folder.

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

// HandleSearchMessages returns an mcp-go tool handler bound to the
// given auth.Manager and mail.Client.
func HandleSearchMessages(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("search_messages: 'account' argument is required"), nil
		}
		query := req.GetString("query", "")
		if query == "" {
			return mcp.NewToolResultError("search_messages: 'query' argument is required (KQL-style search string)"), nil
		}

		opts := mail.ListMessagesOptions{
			FolderID: req.GetString("folder", ""),
			Top:      int(req.GetInt("top", 25)),
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search_messages: auth: %s", err)), nil
		}

		resp, err := graph.SearchMessages(ctx, bearer, query, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search_messages: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search_messages: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}
