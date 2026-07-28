// folders_test.go + categories_test.go + rules_test.go + profile_test.go +
// settings_test.go — coverage de los paquetes restantes.
//
// Cada test es minimo pero verifica: (a) path construido correctamente,
// (b) Authorization header presente, (c) shape JSON básico.

package mail

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestListFolders_TopLevelPath verifica /me/mailFolders.
func TestListFolders_TopLevelPath(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[]}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.ListFolders(context.Background(), "bearer", ListFoldersOptions{Top: 10})
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if capturedPath != "/me/mailFolders" {
		t.Errorf("path: got %q want /me/mailFolders", capturedPath)
	}
}

// TestListFolders_WithParent_ChildPath verifica /me/mailFolders/{id}/childFolders.
func TestListFolders_WithParent_ChildPath(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[]}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.ListFolders(context.Background(), "bearer", ListFoldersOptions{
		ParentID: "folder-xyz",
		Top:      50,
	})
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if !strings.HasSuffix(capturedPath, "/me/mailFolders/folder-xyz/childFolders") {
		t.Errorf("path: got %q", capturedPath)
	}
}

// TestGetFolder_BuildsPath verifica /me/mailFolders/{id}.
func TestGetFolder_BuildsPath(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = io.WriteString(w, `{"id":"folder-1","displayName":"Inbox","totalItemCount":42}`)
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.GetFolder(context.Background(), "bearer", "folder-1")
	if err != nil {
		t.Fatalf("GetFolder: %v", err)
	}
	if capturedPath != "/me/mailFolders/folder-1" {
		t.Errorf("path: got %q", capturedPath)
	}
	if resp.TotalItemCount != 42 {
		t.Errorf("TotalItemCount: got %d want 42", resp.TotalItemCount)
	}
}

// TestListCategories_Path verifica /me/outlook/masterCategories.
func TestListCategories_Path(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[{"id":"cat1","displayName":"Red category","color":"preset1"}]}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.ListCategories(context.Background(), "bearer")
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if capturedPath != "/me/outlook/masterCategories" {
		t.Errorf("path: got %q", capturedPath)
	}
	if len(resp.Value) != 1 || resp.Value[0].DisplayName != "Red category" {
		t.Errorf("response shape mismatch")
	}
}

// TestListRules_Path verifica /me/mailFolders/inbox/messageRules.
func TestListRules_Path(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[{"id":"rule1","displayName":"Move newsletters","isEnabled":true,"sequence":1}]}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.ListRules(context.Background(), "bearer", ListRulesOptions{Top: 10})
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	if capturedPath != "/me/mailFolders/inbox/messageRules" {
		t.Errorf("path: got %q", capturedPath)
	}
	if len(resp.Value) != 1 || !resp.Value[0].IsEnabled {
		t.Errorf("response shape mismatch")
	}
}

// TestGetUserProfile_Path verifica /me.
func TestGetUserProfile_Path(t *testing.T) {
	var capturedPath, capturedLocale string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedLocale = r.Header.Get("Accept-Language")
		_, _ = w.Write([]byte(`{"id":"u1","displayName":"Test User","mail":"test@example.com"}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.GetUserProfile(context.Background(), "bearer", "es-MX")
	if err != nil {
		t.Fatalf("GetUserProfile: %v", err)
	}
	if capturedPath != "/me" {
		t.Errorf("path: got %q want /me", capturedPath)
	}
	if capturedLocale != "es-MX" {
		t.Errorf("Accept-Language: got %q want es-MX", capturedLocale)
	}
	if resp.Mail != "test@example.com" {
		t.Errorf("mail: got %q", resp.Mail)
	}
}

// TestGetUserProfile_EmptyLocale_NoHeader verifica que locale vacío NO
// setea Accept-Language.
func TestGetUserProfile_EmptyLocale_NoHeader(t *testing.T) {
	var capturedLocale string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedLocale = r.Header.Get("Accept-Language")
		_, _ = w.Write([]byte(`{"id":"u1"}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, _ = client.GetUserProfile(context.Background(), "bearer", "")
	if capturedLocale != "" {
		t.Errorf("Accept-Language should be empty: got %q", capturedLocale)
	}
}

// TestGetMailboxSettings_Path verifica /me/mailboxSettings.
func TestGetMailboxSettings_Path(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"automaticReplies":{"status":"disabled"},"timeZone":{"name":"UTC"}}`))
	})
	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	resp, err := client.GetMailboxSettings(context.Background(), "bearer", "")
	if err != nil {
		t.Fatalf("GetMailboxSettings: %v", err)
	}
	if capturedPath != "/me/mailboxSettings" {
		t.Errorf("path: got %q want /me/mailboxSettings", capturedPath)
	}
	if resp.AutomaticReplies.Status != "disabled" {
		t.Errorf("automaticReplies.status: got %q", resp.AutomaticReplies.Status)
	}
	if resp.TimeZone == nil || resp.TimeZone.Name != "UTC" {
		t.Errorf("timeZone.name mismatch")
	}
}

// TestBearerRequired_AllMethods verifica el guard de bearer vacío en
// TODOS los métodos públicos. Si alguien agrega un método nuevo y olvida
// el guard, este test lo cacha.
func TestBearerRequired_AllMethods(t *testing.T) {
	client := NewClient(nil)
	ctx := context.Background()
	type call struct {
		name string
		fn   func() error
	}
	calls := []call{
		{"ListMessages_emptyBearer", func() error { _, e := client.ListMessages(ctx, "", ListMessagesOptions{}); return e }},
		{"GetMessage_emptyBearer", func() error { _, e := client.GetMessage(ctx, "", "m", ""); return e }},
		{"SearchMessages_emptyBearer", func() error { _, e := client.SearchMessages(ctx, "", "q", ListMessagesOptions{}); return e }},
		{"ListFolders_emptyBearer", func() error { _, e := client.ListFolders(ctx, "", ListFoldersOptions{}); return e }},
		{"GetFolder_emptyBearer", func() error { _, e := client.GetFolder(ctx, "", "f"); return e }},
		{"ListCategories_emptyBearer", func() error { _, e := client.ListCategories(ctx, ""); return e }},
		{"ListAttachments_emptyBearer", func() error { _, e := client.ListAttachments(ctx, "", "m", ListAttachmentsOptions{}); return e }},
		{"GetAttachment_emptyBearer", func() error { _, e := client.GetAttachment(ctx, "", "m", "a", 0); return e }},
		{"ListRules_emptyBearer", func() error { _, e := client.ListRules(ctx, "", ListRulesOptions{}); return e }},
		{"GetUserProfile_emptyBearer", func() error { _, e := client.GetUserProfile(ctx, "", ""); return e }},
		{"GetMailboxSettings_emptyBearer", func() error { _, e := client.GetMailboxSettings(ctx, "", ""); return e }},
	}
	for _, c := range calls {
		err := c.fn()
		if err == nil {
			t.Errorf("%s: expected error for empty bearer, got nil", c.name)
		}
	}
}