// get_attachment.go — handler MCP para get_attachment.
//
// Devuelve un adjunto individual con su contenido en base64 dentro del
// campo contentBytes. Por seguridad, el caller DEBE especificar un cap
// de tamaño (max_bytes) para evitar traer adjuntos enormes al contexto
// del LLM. El cap recomendado es 10 MiB; el cap absoluto del sistema
// es 25 MiB.

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

// HandleGetAttachment devuelve el handler MCP bound al Manager + Client.
// max_bytes se aplica post-fetch contra el tamaño aproximado (base64 * 3/4).
func HandleGetAttachment(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("get_attachment: 'account' argument is required"), nil
		}
		messageID := req.GetString("message_id", "")
		if messageID == "" {
			return mcp.NewToolResultError("get_attachment: 'message_id' argument is required"), nil
		}
		attachmentID := req.GetString("attachment_id", "")
		if attachmentID == "" {
			return mcp.NewToolResultError("get_attachment: 'attachment_id' argument is required"), nil
		}
		// max_bytes=0 = sin cap. El handler igualmente aplica el cap
		// duro de 25 MiB del cliente HTTP. Si querés seguro, pasá un
		// valor explícito (recomendado: 10 MiB).
		maxBytes := int(req.GetInt("max_bytes", mail.RecommendedAttachmentCap))

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_attachment: auth: %s", err)), nil
		}

		resp, err := graph.GetAttachment(ctx, bearer, messageID, attachmentID, maxBytes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_attachment: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_attachment: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}