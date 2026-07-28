// list_categories.go — handler MCP para list_categories.
//
// Devuelve las categorías de mensaje definidas por el usuario. Cada
// mensaje expone categories[] (lista de IDs); cruzar esos IDs contra
// este endpoint devuelve los nombres legibles.

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

// HandleListCategories devuelve el handler MCP bound al Manager + Client.
func HandleListCategories(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("list_categories: 'account' argument is required"), nil
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_categories: auth: %s", err)), nil
		}

		resp, err := graph.ListCategories(ctx, bearer)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_categories: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_categories: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}