// attachments_test.go — tests para el cap de tamaño de get_attachment.

package mail

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

// TestGetAttachment_RejectsOversizeContent verifica que si el contentBytes
// decodificado excede maxBytes, devuelve error sin transferirlo entero.
// (Aquí no transferimos nada porque el guard es post-fetch.)
func TestGetAttachment_RejectsOversizeContent(t *testing.T) {
	// Genera un payload base64 de ~5 MiB decodificados.
	big := make([]byte, 5*1024*1024)
	for i := range big {
		big[i] = byte(i % 256)
	}
	encoded := base64.StdEncoding.EncodeToString(big)

	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"@odata.type":"#microsoft.graph.fileAttachment","id":"att1","name":"big.bin","contentBytes":"` + encoded + `"}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL

	// Pedimos cap de 1 MiB — el adjunto real son ~5 MiB.
	_, err := client.GetAttachment(context.Background(), "bearer", "msg1", "att1", 1*1024*1024)
	if err == nil {
		t.Fatal("expected error for oversize attachment, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds cap") {
		t.Errorf("error should mention 'exceeds cap': got %q", err.Error())
	}
}

// TestGetAttachment_AllowsUnderCap verifica el happy path.
func TestGetAttachment_AllowsUnderCap(t *testing.T) {
	// 1 KiB decodificado.
	small := make([]byte, 1024)
	for i := range small {
		small[i] = byte(i % 256)
	}
	encoded := base64.StdEncoding.EncodeToString(small)

	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"@odata.type":"#microsoft.graph.fileAttachment","id":"att2","name":"small.bin","size":1024,"contentBytes":"` + encoded + `"}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL

	resp, err := client.GetAttachment(context.Background(), "bearer", "msg1", "att2", 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachment: %v", err)
	}
	if resp.Name != "small.bin" {
		t.Errorf("Name: %q", resp.Name)
	}
	if len(resp.ContentBytes) == 0 {
		t.Error("ContentBytes: empty")
	}
}

// TestGetAttachment_NegativeMaxBytes_Rejected verifica el guard de input.
func TestGetAttachment_NegativeMaxBytes_Rejected(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetAttachment(context.Background(), "bearer", "m", "a", -1)
	if err == nil {
		t.Fatal("expected error for negative maxBytes, got nil")
	}
}

// TestGetAttachment_HardCapRejected verifica que pedir cap > MaxAttachmentBytes
// falla sin hacer request.
func TestGetAttachment_HardCapRejected(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetAttachment(context.Background(), "bearer", "m", "a", MaxAttachmentBytes+1)
	if err == nil {
		t.Fatal("expected error for cap above MaxAttachmentBytes, got nil")
	}
	if !strings.Contains(err.Error(), "hard cap") {
		t.Errorf("error should mention 'hard cap': got %q", err.Error())
	}
}

// TestListAttachments_BuildsURL verifica que el path es /me/messages/{id}/attachments.
func TestListAttachments_BuildsURL(t *testing.T) {
	var capturedPath string
	srv, httpClient := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"value":[]}`))
	})

	client := NewClient(httpClient)
	client.BaseURL = srv.URL
	_, err := client.ListAttachments(context.Background(), "bearer", "msg-123", ListAttachmentsOptions{Top: 5})
	if err != nil {
		t.Fatalf("ListAttachments: %v", err)
	}
	if !strings.HasSuffix(capturedPath, "/me/messages/msg-123/attachments") {
		t.Errorf("path: got %q", capturedPath)
	}
}