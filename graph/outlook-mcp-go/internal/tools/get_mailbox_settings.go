// get_mailbox_settings.go — handler MCP para get_mailbox_settings.
//
// Devuelve la configuración del buzón: estado de fuera de oficina,
// zona horaria, horario laboral, formato de fecha/hora, idioma.
// Solo lectura — modificar la configuración queda fuera de esta tanda.

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

// HandleGetMailboxSettings devuelve el handler MCP bound al Manager + Client.
func HandleGetMailboxSettings(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("get_mailbox_settings: 'account' argument is required"), nil
		}
		locale := req.GetString("locale", "")

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_mailbox_settings: auth: %s", err)), nil
		}

		resp, err := graph.GetMailboxSettings(ctx, bearer, locale)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_mailbox_settings: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_mailbox_settings: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}