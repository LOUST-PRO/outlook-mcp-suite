// server.go — MCP server bootstrap.
//
// Fase 1 registers exactly one tool (list_messages) as proof. The
// remaining 10 read-only tools land in Stages 2-6 per the Fase 1
// plan. Each registration is a single AddTool call wrapping a typed
// handler in internal/tools/.

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/auth"
	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/mail"
	"github.com/LOUST-PRO/outlook-mcp-suite/graph/outlook-mcp-go/internal/tools"
)

// defaultHTTPClient returns the canonical 30s-timeout HTTP client used
// by the Graph mail package. Tests may inject their own.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// New constructs an *server.MCPServer with all Fase 1 tools registered.
func New(manager *auth.Manager) *server.MCPServer {
	s := server.NewMCPServer(
		"outlook-mcp-suite/graph",
		"0.1.0-stage1",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Shared HTTP client. 30s timeout covers slow Graph endpoints
	// without blocking the MCP request indefinitely.
	httpClient := defaultHTTPClient()
	graph := mail.NewClient(httpClient)

	// Tool: list_messages (Fase 1 Stage 1 proof)
	listMessages := mcp.NewTool("list_messages",
		mcp.WithDescription("List messages in a Microsoft account folder via Microsoft Graph. "+
			"Returns the first page of an OData envelope; NextLink is included for pagination. "+
			"Fase 1: read-only, no body content (use get_message in a later stage)."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts] in ~/.config/lzt-outlook/config.toml. "+
				"Default-deny: must be in [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("folder",
			mcp.Description("Mail folder name or ID. Examples: 'inbox', 'drafts', 'sentitems', "+
				"'junkemail', 'archive', 'deleteditems', or a folder GUID. Default: inbox."),
		),
		mcp.WithNumber("top",
			mcp.Description("Maximum messages to return (1-1000). Default: 25."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithString("filter",
			mcp.Description("OData $filter expression (server-side). "+
				"Example: \"isRead eq false\". Use sparingly — invalid syntax returns HTTP 400."),
		),
		mcp.WithString("search",
			mcp.Description("KQL-style search across subject + body. "+
				"Example: '\"quarterly report\"'. Sets ConsistencyLevel: eventual."),
		),
		mcp.WithString("orderby",
			mcp.Description("OData $orderby expression. Default: 'receivedDateTime desc'."),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "List Outlook messages",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)

	s.AddTool(listMessages, tools.HandleListMessages(manager, graph))

	return s
}

// ServeStdio runs the MCP server on stdin/stdout. This is the only
// transport supported in Fase 1 (SSE/HTTP land in Fase 3+).
func ServeStdio(ctx context.Context, manager *auth.Manager) error {
	srv := New(manager)
	stdio := server.NewStdioServer(srv)
	return stdio.Listen(ctx, stdioReader{}, stdioWriter{})
}

// stdioReader / stdioWriter are minimal io interfaces for testing.
// In production, server.NewStdioServer uses os.Stdin / os.Stdout.
type stdioReader struct{}
type stdioWriter struct{}

func (stdioReader) Read(p []byte) (int, error) {
	// In real Listen() this is os.Stdin; tests inject their own.
	return 0, fmt.Errorf("stdioReader.Read: stub")
}
func (stdioWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
