// Package mail — messages_test.go: smoke tests con httptest.
//
// Estos tests NO tocan Microsoft Graph real. Usan httptest.NewServer
// para servir respuestas JSON controladas y verificar que el cliente
// construye URLs correctas, setea headers correctos, y deserializa
// responses en los structs esperados.
//
// Para tests contra una cuenta real (smoke E2E), ver
// docs/onboarding-real-account.md.

package mail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestServer wraps httptest.NewServer with a mux and a recording
// function so tests can inspect the last request URL + headers.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, srv.Client()
}

// TestListMessages_BuildsCorrectRequest verifica que ListMessages
// setea los query params correctos y el header ConsistencyLevel.
func TestListMessages_BuildsCorrectRequest(t *testing.T) {
	var capturedURL *url.URL
	var capturedHeaders http.Header
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"msg1","subject":"hi"}]}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL

	opts := ListMessagesOptions{
		FolderID: "inbox",
		Top:      10,
		Filter:   "isRead eq false",
		Search:   "quarterly report",
		OrderBy:  "receivedDateTime desc",
	}

	resp, err := client.ListMessages(context.Background(), "fake-bearer", opts)
	if err != nil {
		t.Fatalf("ListMessages: unexpected error: %v", err)
	}
	if len(resp.Value) != 1 || resp.Value[0].ID != "msg1" {
		t.Errorf("response shape mismatch: got %+v", resp)
	}

	// Path: /me/mailFolders/inbox/messages
	if !strings.HasSuffix(capturedURL.Path, "/me/mailFolders/inbox/messages") {
		t.Errorf("path mismatch: got %q", capturedURL.Path)
	}

	// Query params.
	q := capturedURL.Query()
	if q.Get("$top") != "10" {
		t.Errorf("$top: got %q want 10", q.Get("$top"))
	}
	if q.Get("$filter") != "isRead eq false" {
		t.Errorf("$filter: got %q", q.Get("$filter"))
	}
	if q.Get("$search") != `"quarterly report"` {
		t.Errorf("$search: got %q want %q (quoted)", q.Get("$search"), `"quarterly report"`)
	}
	if q.Get("$orderby") != "receivedDateTime desc" {
		t.Errorf("$orderby: got %q", q.Get("$orderby"))
	}

	// Headers.
	if capturedHeaders.Get("Authorization") != "Bearer fake-bearer" {
		t.Errorf("Authorization: got %q", capturedHeaders.Get("Authorization"))
	}
	if capturedHeaders.Get("ConsistencyLevel") != "eventual" {
		t.Errorf("ConsistencyLevel: got %q", capturedHeaders.Get("ConsistencyLevel"))
	}
	if capturedHeaders.Get("Accept") != "application/json" {
		t.Errorf("Accept: got %q", capturedHeaders.Get("Accept"))
	}
}

// TestListMessages_NoFolder_TopLevel verifica que sin FolderID se usa
// /me/messages (no /me/mailFolders/.../messages).
func TestListMessages_NoFolder_TopLevel(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[]}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.ListMessages(context.Background(), "bearer", ListMessagesOptions{})
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if capturedPath != "/me/messages" {
		t.Errorf("path: got %q want /me/messages", capturedPath)
	}
}

// TestListMessages_HTTPError_Propagates verifica que errores HTTP no-2xx
// se devuelven como Go errors con contexto útil.
func TestListMessages_HTTPError_Propagates(t *testing.T) {
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"ErrorItemNotFound","message":"Insufficient privileges"}}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.ListMessages(context.Background(), "bearer", ListMessagesOptions{Top: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Errorf("error should mention status 403: got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Insufficient privileges") {
		t.Errorf("error should include Graph message: got %q", err.Error())
	}
}

// TestGetMessage_WithSelect verifica que select=... se envía como $select.
func TestGetMessage_WithSelect(t *testing.T) {
	var capturedSelect string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedSelect = r.URL.Query().Get("$select")
		_, _ = io.WriteString(w, `{"id":"m1","body":{"contentType":"text","content":"hello"}}`)
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.GetMessage(context.Background(), "bearer", "msg-xyz", "id,subject,body")
	if err != nil {
		t.Fatalf("GetMessage: %v", err)
	}
	if capturedSelect != "id,subject,body" {
		t.Errorf("$select: got %q want %q", capturedSelect, "id,subject,body")
	}
	if resp.Body.Content != "hello" {
		t.Errorf("body.content: got %q", resp.Body.Content)
	}
	if resp.Body.ContentType != "text" {
		t.Errorf("body.contentType: got %q", resp.Body.ContentType)
	}
}

// TestGetMessage_EmptyMessageID_Rejected verifica el guard de input.
func TestGetMessage_EmptyMessageID_Rejected(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMessage(context.Background(), "bearer", "", "")
	if err == nil {
		t.Fatal("expected error for empty messageID, got nil")
	}
	if !strings.Contains(err.Error(), "empty messageID") {
		t.Errorf("error: got %q", err.Error())
	}
}

// TestSearchMessages_OverridesFilterAndOrderBy verifica que SearchMessages
// limpia $filter y fuerza $orderby (porque Graph los ignora cuando hay $search).
func TestSearchMessages_OverridesFilterAndOrderBy(t *testing.T) {
	var capturedURL *url.URL
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		_, _ = w.Write([]byte(`{"value":[]}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.SearchMessages(context.Background(), "bearer", "from:boss", ListMessagesOptions{
		Filter:  "isRead eq false", // debería ser sobreescrito
		OrderBy: "subject asc",     // debería ser sobreescrito
		Top:     5,
	})
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	q := capturedURL.Query()
	if q.Get("$filter") != "" {
		t.Errorf("$filter: got %q want empty", q.Get("$filter"))
	}
	if q.Get("$orderby") != "receivedDateTime desc" {
		t.Errorf("$orderby: got %q want receivedDateTime desc", q.Get("$orderby"))
	}
	if q.Get("$search") != `"from:boss"` {
		t.Errorf("$search: got %q want %q", q.Get("$search"), `"from:boss"`)
	}
}

// TestSearchMessages_EmptyQuery_Rejected verifica el guard.
func TestSearchMessages_EmptyQuery_Rejected(t *testing.T) {
	client := NewClient(nil)
	_, err := client.SearchMessages(context.Background(), "bearer", "", ListMessagesOptions{})
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

// TestBearerAuth_HeaderAlwaysSet verifica que CADA request lleva el
// header Authorization (defensa contra bugs donde se olvida en alguna ruta).
func TestBearerAuth_HeaderAlwaysSet(t *testing.T) {
	requests := 0
	var lastAuth string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		lastAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"value":[]}`)
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	ctx := context.Background()

	// Cada llamada debe llevar Authorization.
	_, _ = client.ListMessages(ctx, "tok-A", ListMessagesOptions{Top: 1})
	if lastAuth != "Bearer tok-A" {
		t.Errorf("ListMessages: Authorization=%q", lastAuth)
	}
	_, _ = client.ListFolders(ctx, "tok-B", ListFoldersOptions{Top: 1})
	if lastAuth != "Bearer tok-B" {
		t.Errorf("ListFolders: Authorization=%q", lastAuth)
	}
	_, _ = client.ListCategories(ctx, "tok-C")
	if lastAuth != "Bearer tok-C" {
		t.Errorf("ListCategories: Authorization=%q", lastAuth)
	}
	_, _ = client.GetUserProfile(ctx, "tok-D", "")
	if lastAuth != "Bearer tok-D" {
		t.Errorf("GetUserProfile: Authorization=%q", lastAuth)
	}
	if requests != 4 {
		t.Errorf("expected 4 requests, got %d", requests)
	}
}

// TestListMessages_EmptyBearer_Rejected verifica el guard.
func TestListMessages_EmptyBearer_Rejected(t *testing.T) {
	client := NewClient(nil)
	_, err := client.ListMessages(context.Background(), "", ListMessagesOptions{})
	if err == nil {
		t.Fatal("expected error for empty bearer, got nil")
	}
}

// TestListMessages_JSONShape decodifica un envelope real y verifica
// que los campos anidados se preservan.
func TestListMessages_JSONShape(t *testing.T) {
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('me')/messages",
			"@odata.nextLink": "https://graph.microsoft.com/v1.0/me/messages?$skiptoken=abc",
			"value": []map[string]any{
				{
					"id":               "msg1",
					"conversationId":   "conv1",
					"subject":          "Q3 report",
					"receivedDateTime": "2026-07-15T10:00:00Z",
					"isRead":           false,
					"hasAttachments":   true,
					"importance":       "high",
					"bodyPreview":      "Quarterly numbers...",
					"from": map[string]any{
						"emailAddress": map[string]any{
							"name":    "Boss Person",
							"address": "boss@example.com",
						},
					},
				},
			},
		})
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.ListMessages(context.Background(), "bearer", ListMessagesOptions{Top: 1})
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if resp.NextLink == "" {
		t.Error("NextLink not preserved")
	}
	if len(resp.Value) != 1 {
		t.Fatalf("Value: got %d entries", len(resp.Value))
	}
	m := resp.Value[0]
	if m.Subject != "Q3 report" {
		t.Errorf("Subject: %q", m.Subject)
	}
	if !m.HasAttachments {
		t.Error("HasAttachments: got false want true")
	}
	if m.From == nil || m.From.EmailAddress == nil {
		t.Fatal("From.EmailAddress: nil")
	}
	if m.From.EmailAddress.Address != "boss@example.com" {
		t.Errorf("From.EmailAddress.Address: %q", m.From.EmailAddress.Address)
	}
}