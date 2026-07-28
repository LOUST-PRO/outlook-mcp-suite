// list_rules.go — handler MCP para list_rules.
//
// Devuelve las reglas activas del inbox. Cada regla expone sus
// condiciones, acciones y excepciones como JSON crudo — los consumidores
// pueden decodificarlas contra los schemas de microsoft.graph si
// necesitan tipos finos.

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

// HandleListRules devuelve el handler MCP bound al Manager + Client.
func HandleListRules(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("list_rules: 'account' argument is required"), nil
		}

		opts := mail.ListRulesOptions{
			Top: int(req.GetInt("top", 100)),
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_rules: auth: %s", err)), nil
		}

		resp, err := graph.ListRules(ctx, bearer, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_rules: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_rules: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}