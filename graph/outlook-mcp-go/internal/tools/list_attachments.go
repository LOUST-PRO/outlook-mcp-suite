// list_attachments.go — handler MCP para list_attachments.
//
// Devuelve la metadata de los adjuntos de un mensaje: nombre, tipo MIME,
// tamaño, si es inline. NO incluye el contenido binario — eso requiere
// get_attachment explícito con un cap de tamaño.

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

// HandleListAttachments devuelve el handler MCP bound al Manager + Client.
func HandleListAttachments(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("list_attachments: 'account' argument is required"), nil
		}
		messageID := req.GetString("message_id", "")
		if messageID == "" {
			return mcp.NewToolResultError("list_attachments: 'message_id' argument is required"), nil
		}

		opts := mail.ListAttachmentsOptions{
			Top: int(req.GetInt("top", 25)),
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_attachments: auth: %s", err)), nil
		}

		resp, err := graph.ListAttachments(ctx, bearer, messageID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_attachments: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_attachments: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}