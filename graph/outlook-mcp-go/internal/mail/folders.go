// Package mail — folders.go: mailbox folder hierarchy via /me/mailFolders.
//
// Graph expone /me/mailFolders (well-known + custom), /me/mailFolders/{id}
// (single), y /me/mailFolders/{id}/childFolders (hierarchy descendente).
// Esta implementación soporta navegación top-level y single-folder; el
// subtree recursivo queda fuera de Fase 1 (lo cubre Fase 3 con un tool
// de traversal bajo demanda del LLM).
package mail

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// MailFolder es el subset del resource mailFolder de Graph. Incluye
// contadores para evitar un round-trip extra en el caso común.
type MailFolder struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName,omitempty"`
	ParentFolderID   string `json:"parentFolderId,omitempty"`
	ChildFolderCount int    `json:"childFolderCount,omitzero"`
	TotalItemCount   int    `json:"totalItemCount,omitzero"`
	UnreadItemCount  int    `json:"unreadItemCount,omitzero"`
	WellKnownName    string `json:"wellKnownName,omitempty"`
}

// MailFolderList es el OData envelope estándar de Graph.
type MailFolderList struct {
	Context  string       `json:"@odata.context,omitempty"`
	Count    int          `json:"@odata.count,omitempty"`
	NextLink string       `json:"@odata.nextLink,omitempty"`
	Value    []MailFolder `json:"value"`
}

// ListFoldersOptions controla el alcance de list_folders.
type ListFoldersOptions struct {
	// ParentID scopes la lista a /me/mailFolders/{id}/childFolders.
	// Empty = top-level /me/mailFolders.
	ParentID string

	// Top es el tamaño de página (1-1000).
	Top int
}

// ListFolders devuelve la primera página del folder list. Para una
// jerarquía completa, el caller itera con ParentID desde childFolderCount.
func (c *Client) ListFolders(ctx context.Context, bearer string, opts ListFoldersOptions) (*MailFolderList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if opts.Top <= 0 || opts.Top > 1000 {
		opts.Top = 100 // Graph default para mailFolders
	}

	endpoint := "/me/mailFolders"
	if opts.ParentID != "" {
		endpoint = fmt.Sprintf("/me/mailFolders/%s/childFolders", url.PathEscape(opts.ParentID))
	}

	q := url.Values{}
	q.Set("$top", fmt.Sprintf("%d", opts.Top))
	full := c.BaseURL + endpoint + "?" + q.Encode()

	resp := &MailFolderList{}
	if err := c.doGet(ctx, bearer, full, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetFolder devuelve un único folder por ID. Útil para inspeccionar
// contadores (TotalItemCount / UnreadItemCount) sin listar todo el árbol.
func (c *Client) GetFolder(ctx context.Context, bearer, folderID string) (*MailFolder, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if folderID == "" {
		return nil, errors.New("mail: empty folderID")
	}

	endpoint := fmt.Sprintf("/me/mailFolders/%s", url.PathEscape(folderID))
	full := c.BaseURL + endpoint

	var out MailFolder
	if err := c.doGet(ctx, bearer, full, &out); err != nil {
		return nil, err
	}
	return &out, nil
}