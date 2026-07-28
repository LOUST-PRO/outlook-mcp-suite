// get_message.go — MCP tool handler for get_message.
//
// Fetches the full resource for a single message including its body
// (text or HTML). Use sparingly — large HTML bodies can be several
// hundred KiB. Pass select="id,subject,from,receivedDateTime" to
// fetch metadata only when the body isn't needed.

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

// DefaultMessageSelect is the Graph $select projection used when the
// caller doesn't specify one. Excludes body for first-page reads;
// callers can override to add body or strip further.
const DefaultMessageSelect = "id,conversationId,subject,sender,from,toRecipients,receivedDateTime,isRead,hasAttachments,importance,bodyPreview,categories"

// HandleGetMessage returns an mcp-go tool handler bound to the given
// auth.Manager and mail.Client.
func HandleGetMessage(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("get_message: 'account' argument is required"), nil
		}
		messageID := req.GetString("message_id", "")
		if messageID == "" {
			return mcp.NewToolResultError("get_message: 'message_id' argument is required (the Graph message ID, not a conversation ID)"), nil
		}
		selectFields := req.GetString("select", "")
		if selectFields == "" {
			selectFields = DefaultMessageSelect
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_message: auth: %s", err)), nil
		}

		resp, err := graph.GetMessage(ctx, bearer, messageID, selectFields)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_message: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_message: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}
