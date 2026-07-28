// server.go — MCP server bootstrap.
//
// Stage 1: list_messages.
// Stage 2: get_message + search_messages.
// Stage 3: list_folders + get_folder + list_categories.
// Stage 4: list_attachments + get_attachment + list_rules.
// Stage 5: get_user_profile + get_mailbox_settings.
// Stage 6 (final): smoke test + onboarding docs.

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
		"0.1.0-stage5",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Shared HTTP client. 30s timeout covers slow Graph endpoints
	// without blocking the MCP request indefinitely.
	httpClient := defaultHTTPClient()
	graph := mail.NewClient(httpClient)

	// Tool: list_messages (Fase 1 Stage 1)
	listMessages := mcp.NewTool("list_messages",
		mcp.WithDescription("List messages in a Microsoft account folder via Microsoft Graph. "+
			"Returns the first page of an OData envelope; NextLink is included for pagination. "+
			"Fase 1: read-only, no body content (use get_message to fetch full body)."),
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

	// Tool: get_message (Fase 1 Stage 2)
	getMessage := mcp.NewTool("get_message",
		mcp.WithDescription("Fetch a single message by Graph message ID. "+
			"Returns the full resource including body (text or HTML). "+
			"Use select to project only specific fields and reduce payload size."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("message_id",
			mcp.Description("Graph message ID (returned by list_messages / search_messages)."),
			mcp.Required(),
		),
		mcp.WithString("select",
			mcp.Description("OData $select projection. Default includes bodyPreview but excludes body. "+
				"Pass 'body' explicitly to fetch full HTML/text body. "+
				"Example: 'id,subject,from,body'."),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Get Outlook message by ID",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(getMessage, tools.HandleGetMessage(manager, graph))

	// Tool: search_messages (Fase 1 Stage 2)
	searchMessages := mcp.NewTool("search_messages",
		mcp.WithDescription("KQL-style search across message subject + body. "+
			"Results ranked by Graph search relevance. Optionally scope to a folder."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("query",
			mcp.Description("KQL search string. Examples: '\"quarterly report\"', "+
				"'from:boss subject:urgent', 'hasAttachments:true'."),
			mcp.Required(),
		),
		mcp.WithString("folder",
			mcp.Description("Optional folder scope ('inbox', 'sentitems', etc.). Default: all folders."),
		),
		mcp.WithNumber("top",
			mcp.Description("Maximum messages to return (1-1000). Default: 25."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Search Outlook messages (KQL)",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(searchMessages, tools.HandleSearchMessages(manager, graph))

	// Tool: list_folders (Stage 3)
	listFolders := mcp.NewTool("list_folders",
		mcp.WithDescription("List mail folders in the mailbox. Top-level by default; pass parent=<id> to list children of a specific folder."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("parent",
			mcp.Description("Optional parent folder ID. Empty = top-level /me/mailFolders."),
		),
		mcp.WithNumber("top",
			mcp.Description("Maximum folders to return (1-1000). Default: 100."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "List Outlook mail folders",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(listFolders, tools.HandleListFolders(manager, graph))

	// Tool: get_folder (Stage 3)
	getFolder := mcp.NewTool("get_folder",
		mcp.WithDescription("Fetch a single mail folder by ID. Returns counters (totalItemCount, unreadItemCount, childFolderCount)."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("folder_id",
			mcp.Description("Graph folder ID (returned by list_folders)."),
			mcp.Required(),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Get Outlook mail folder",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(getFolder, tools.HandleGetFolder(manager, graph))

	// Tool: list_categories (Stage 3)
	listCategories := mcp.NewTool("list_categories",
		mcp.WithDescription("List all message categories defined by the user. Categories are referenced by ID from the categories[] field on each message."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "List Outlook message categories",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(listCategories, tools.HandleListCategories(manager, graph))

	// Tool: list_attachments (Stage 4)
	listAttachments := mcp.NewTool("list_attachments",
		mcp.WithDescription("List attachments on a message. Returns metadata only (name, contentType, size, isInline) — does NOT include the binary content. Use get_attachment with max_bytes to fetch the actual bytes."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("message_id",
			mcp.Description("Graph message ID."),
			mcp.Required(),
		),
		mcp.WithNumber("top",
			mcp.Description("Maximum attachments to return (1-1000). Default: 25."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "List Outlook message attachments",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(listAttachments, tools.HandleListAttachments(manager, graph))

	// Tool: get_attachment (Stage 4)
	getAttachment := mcp.NewTool("get_attachment",
		mcp.WithDescription("Fetch a single attachment by ID. The binary content is returned base64-encoded in the contentBytes field. Always pass max_bytes to cap the response size; recommended cap is 10 MiB."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("message_id",
			mcp.Description("Graph message ID containing the attachment."),
			mcp.Required(),
		),
		mcp.WithString("attachment_id",
			mcp.Description("Graph attachment ID (returned by list_attachments)."),
			mcp.Required(),
		),
		mcp.WithNumber("max_bytes",
			mcp.Description("Cap on attachment size in bytes (decoded). Recommended: 10485760 (10 MiB). Hard limit: 26214400 (25 MiB). Default: 10 MiB. Returns an error before downloading if the attachment exceeds this."),
			mcp.Min(0),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Get Outlook attachment content",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(getAttachment, tools.HandleGetAttachment(manager, graph))

	// Tool: list_rules (Stage 4)
	listRules := mcp.NewTool("list_rules",
		mcp.WithDescription("List inbox message rules. Each rule exposes its conditions, actions, and exceptions. The actions describe what the rule is configured to do (forward, move, mark) — this tool only reads the configuration; it does not execute any rule."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithNumber("top",
			mcp.Description("Maximum rules to return (1-1000). Default: 100."),
			mcp.Min(1), mcp.Max(1000),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "List Outlook inbox rules",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(listRules, tools.HandleListRules(manager, graph))

	// Tool: get_user_profile (Stage 5)
	getUserProfile := mcp.NewTool("get_user_profile",
		mcp.WithDescription("Fetch the authenticated user's profile from /me: display name, UPN, primary mail, job title, phones, office location, preferred language."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("locale",
			mcp.Description("Optional Accept-Language header (e.g. 'en-US', 'es-MX'). Empty = no header."),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Get Outlook user profile",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(getUserProfile, tools.HandleGetUserProfile(manager, graph))

	// Tool: get_mailbox_settings (Stage 5)
	getMailboxSettings := mcp.NewTool("get_mailbox_settings",
		mcp.WithDescription("Fetch mailbox settings: out-of-office (automaticReplies), time zone, working hours, date/time format, language. Read-only."),
		mcp.WithString("account",
			mcp.Description("Account name from [accounts].allowed."),
			mcp.Required(),
		),
		mcp.WithString("locale",
			mcp.Description("Optional Accept-Language header. Empty = no header."),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Get Outlook mailbox settings",
			ReadOnlyHint:    new(true),
			DestructiveHint: new(false),
			IdempotentHint:  new(true),
			OpenWorldHint:   new(false),
		}),
	)
	s.AddTool(getMailboxSettings, tools.HandleGetMailboxSettings(manager, graph))

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
