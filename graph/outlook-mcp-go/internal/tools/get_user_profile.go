// get_user_profile.go — handler MCP para get_user_profile.
//
// Devuelve la identidad del usuario autenticado: nombre visible,
// correo principal, UPN, puesto, teléfonos, ubicación, idioma.
// Aplica locale opcional vía Accept-Language.

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

// HandleGetUserProfile devuelve el handler MCP bound al Manager + Client.
func HandleGetUserProfile(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("get_user_profile: 'account' argument is required"), nil
		}
		locale := req.GetString("locale", "")

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_user_profile: auth: %s", err)), nil
		}

		resp, err := graph.GetUserProfile(ctx, bearer, locale)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_user_profile: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_user_profile: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}