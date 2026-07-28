// get_folder.go — handler MCP para get_folder.
//
// Devuelve un único folder con sus contadores (TotalItemCount,
// UnreadItemCount, ChildFolderCount). Útil para mostrar el estado
// del inbox antes de listar mensajes.

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

// HandleGetFolder devuelve el handler MCP bound al Manager + Client.
func HandleGetFolder(manager *auth.Manager, graph *mail.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		account := req.GetString("account", "")
		if account == "" {
			return mcp.NewToolResultError("get_folder: 'account' argument is required"), nil
		}
		folderID := req.GetString("folder_id", "")
		if folderID == "" {
			return mcp.NewToolResultError("get_folder: 'folder_id' argument is required (Graph folder ID, not display name)"), nil
		}

		bearer, err := manager.BearerFor(ctx, account)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_folder: auth: %s", err)), nil
		}

		resp, err := graph.GetFolder(ctx, bearer, folderID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_folder: Graph: %s", err)), nil
		}

		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_folder: marshal: %s", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}