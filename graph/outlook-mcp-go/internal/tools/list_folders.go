// list_folders.go — handler MCP para list_folders.
//
// Devuelve la jerarquía top-level (o los hijos de un folder específico).
// El caller itera con parent=<id> si necesita descender.

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

// HandleListFolders devuelve el handler MCP bound al Manager + Client.
func HandleListFolders(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("list_folders: 'account' argument is required"), nil
		}

		opts := mail.ListFoldersOptions{
			ParentID: req.GetString("parent", ""),
			Top:      int(req.GetInt("top", 100)),
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_folders: auth: %s", err)), nil
		}

		resp, err := graph.ListFolders(ctx, bearer, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_folders: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_folders: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}