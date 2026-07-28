// Package mail implements read-only access to Microsoft Graph mail
// endpoints (/me/messages, /me/mailFolders, /me/messages/{id}/attachments).
//
// Fase 1 Stage 1: list_messages (paginated OData list).
// Fase 1 Stage 2: get_message (single message + body), search_messages
//                  (KQL body+subject search with relevance ranking).
// Subsequent stages land in ../ARCHITECTURE.md §Fase 1 plan.
package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GraphBaseURL is the canonical Microsoft Graph v1.0 endpoint.
const GraphBaseURL = "https://graph.microsoft.com/v1.0"

// MaxBodyBytes caps HTTP response bodies to prevent OOM on runaway
// Graph responses (e.g. an attachment blob).
const MaxBodyBytes = 25 * 1024 * 1024 // 25 MiB

// Client is a thin HTTP wrapper for Microsoft Graph.
// Bearer is injected per-request by the tool handler.
type Client struct {
	HTTP    *http.Client
	BaseURL string
}

// NewClient returns a Client with sensible HTTP defaults.
// httpClient may be nil to use net/http default.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{HTTP: httpClient, BaseURL: GraphBaseURL}
}

// Message is the subset of Microsoft Graph message resource exposed
// by list_messages. Body/BodyPreview are omitted for Fase 1 to keep
// payloads small; tools can fetch full body via get_message in a
// later stage.
type Message struct {
	ID               string       `json:"id"`
	ConversationID   string       `json:"conversationId,omitempty"`
	Subject          string       `json:"subject,omitempty"`
	Sender           *Recipient   `json:"sender,omitzero"`
	From             *Recipient   `json:"from,omitzero"`
	ToRecipients     []*Recipient `json:"toRecipients,omitempty"`
	ReceivedDateTime time.Time    `json:"receivedDateTime,omitzero"`
	IsRead           bool         `json:"isRead,omitzero"`
	HasAttachments   bool         `json:"hasAttachments,omitzero"`
	Importance       string       `json:"importance,omitempty"`
	Preview          string       `json:"bodyPreview,omitempty"`
}

// Recipient is the minimal address shape Graph returns.
type Recipient struct {
	EmailAddress *EmailAddress `json:"emailAddress,omitzero"`
}

// EmailAddress is the leaf address structure.
type EmailAddress struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

// MessageList is the OData envelope returned by /me/messages.
// NextLink is the @odata.nextLink for pagination; empty if no more.
type MessageList struct {
	Context  string    `json:"@odata.context,omitempty"`
	Count    int       `json:"@odata.count,omitempty"`
	NextLink string    `json:"@odata.nextLink,omitempty"`
	Value    []Message `json:"value"`
}

// ListMessagesOptions controls list_messages behavior.
type ListMessagesOptions struct {
	// FolderID scopes the query to a specific mail folder
	// (e.g. "inbox", "drafts", "sentitems"). Empty = default folder.
	FolderID string

	// Filter applies an OData $filter expression (server-side).
	// Use sparingly — invalid syntax returns HTTP 400.
	Filter string

	// Search applies an OData $search expression (server-side,
	// KQL-style). Subject + body are searched.
	Search string

	// Top is the page size (max 1000 per Graph limit).
	Top int

	// Select limits the fields returned (e.g. "id,subject,from").
	// Empty returns the Graph default projection.
	Select string

	// OrderBy is an OData $orderby expression (e.g. "receivedDateTime desc").
	OrderBy string
}

// ListMessages fetches the first page of messages matching opts.
// bearer is the OAuth access token. Returns the envelope with
// NextLink for pagination (caller can re-call with NextLink as a
// custom $skip token, but Fase 1 keeps it single-page).
func (c *Client) ListMessages(ctx context.Context, bearer string, opts ListMessagesOptions) (*MessageList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if opts.Top <= 0 || opts.Top > 1000 {
		opts.Top = 25 // Graph default
	}

	endpoint := "/me/messages"
	if opts.FolderID != "" {
		// /me/mailFolders/{id}/messages — folder IDs are case-insensitive
		// GUIDs OR well-known names ("inbox", "drafts", "sentitems",
		// "deleteditems", "junkemail", "archive", ...).
		endpoint = fmt.Sprintf("/me/mailFolders/%s/messages", url.PathEscape(opts.FolderID))
	}

	q := url.Values{}
	if opts.Top > 0 {
		q.Set("$top", fmt.Sprintf("%d", opts.Top))
	}
	if opts.Filter != "" {
		q.Set("$filter", opts.Filter)
	}
	if opts.Search != "" {
		q.Set("$search", fmt.Sprintf(`"%s"`, opts.Search))
	}
	if opts.Select != "" {
		q.Set("$select", opts.Select)
	}
	if opts.OrderBy != "" {
		q.Set("$orderby", opts.OrderBy)
	}
	full := c.BaseURL + endpoint
	if len(q) > 0 {
		full += "?" + q.Encode()
	}

	resp := &MessageList{}
	if err := c.doGet(ctx, bearer, full, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// doGet is the shared HTTP plumbing: build request, set auth header,
// enforce MaxBodyBytes, decode JSON into out.
func (c *Client) doGet(ctx context.Context, bearer, fullURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("mail: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")
	// ConsistencyLevel: Required for $search across multiple folders.
	// Negligible cost when $search isn't used.
	req.Header.Set("ConsistencyLevel", "eventual")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("mail: HTTP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	if err != nil {
		return fmt.Errorf("mail: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mail: Graph %s: %s (status %d)",
			fullURL, truncate(string(body), 500), resp.StatusCode)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("mail: decode JSON: %w (body: %s)", err, truncate(string(body), 500))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…(truncated)"
}

// ItemBody represents the body of a message, event, or other item.
// ContentType is "text" or "html" per Graph spec.
type ItemBody struct {
	ContentType string `json:"contentType,omitempty"`
	Content     string `json:"content,omitempty"`
}

// MessageDetail is the extended message resource returned by
// get_message. Embeds Message (so common fields stay accessible) and
// adds the body + categories. Used by the get_message tool handler.
type MessageDetail struct {
	Message
	Body       ItemBody `json:"body,omitzero"`
	Categories []string `json:"categories,omitempty"`
}

// GetMessage fetches a single message by ID. If selectFields is non-empty,
// it limits the returned projection via OData $select. bearer is the
// OAuth access token injected by the caller.
func (c *Client) GetMessage(ctx context.Context, bearer, messageID, selectFields string) (*MessageDetail, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if messageID == "" {
		return nil, errors.New("mail: empty messageID")
	}
	endpoint := fmt.Sprintf("/me/messages/%s", url.PathEscape(messageID))
	q := url.Values{}
	if selectFields != "" {
		q.Set("$select", selectFields)
	}
	full := c.BaseURL + endpoint
	if len(q) > 0 {
		full += "?" + q.Encode()
	}

	var out MessageDetail
	if err := c.doGet(ctx, bearer, full, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchMessages is a typed convenience over ListMessages with the
// $search query parameter forced. It overrides OrderBy and Filter
// because Graph's $search implicitly ranks by relevance — caller-
// supplied $filter or $orderby would be ignored server-side.
//
// Graph's $search is KQL-style and requires ConsistencyLevel: eventual,
// which ListMessages sets unconditionally (no cost when search is empty).
func (c *Client) SearchMessages(ctx context.Context, bearer, query string, opts ListMessagesOptions) (*MessageList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if query == "" {
		return nil, errors.New("mail: empty search query")
	}
	opts.Search = query
	opts.Filter = ""
	opts.OrderBy = "receivedDateTime desc" // relevance implied by $search
	return c.ListMessages(ctx, bearer, opts)
}
