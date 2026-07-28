// Package mail — rules.go: reglas de bandeja de entrada vía /me/mailFolders/inbox/messageRules.
//
// Las messageRules son el equivalente Outlook de los "filtros" del
// cliente clásico: reglas del lado servidor que mueven, marcan,
// categorizan o reenvían mensajes según condiciones. Esta herramienta
// entrega la configuración en lectura; crear/actualizar/eliminar reglas
// queda fuera de esta tanda.
//
// CAVEAT DE LECTURA: las acciones de una regla (forwardTo, moveToFolder,
// delete) son DATOS que la regla TIENE configurada, no operaciones que
// esta herramienta ejecuta. La tool es read-only per ToolAnnotation.
package mail

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// MessageRule es el resource de regla. Conditions, Actions y Exceptions
// son sub-recursos complejos de Graph; en esta tanda los exponemos como
// JSON crudo (map[string]any) para mantener el handler simple. Un
// consumidor que quiera tipar fino puede decodificar contra los schemas
// de microsoft.graph.* directamente.
type MessageRule struct {
	ID                  string         `json:"id"`
	DisplayName         string         `json:"displayName,omitempty"`
	Sequence            int            `json:"sequence,omitzero"`
	IsEnabled           bool           `json:"isEnabled,omitzero"`
	IsReadOnly          bool           `json:"isReadOnly,omitzero"`
	Conditions          map[string]any `json:"conditions,omitempty"`
	Actions             map[string]any `json:"actions,omitempty"`
	Exceptions          map[string]any `json:"exceptions,omitempty"`
}

// RuleList es el OData envelope estándar.
type RuleList struct {
	Context  string        `json:"@odata.context,omitempty"`
	Count    int           `json:"@odata.count,omitempty"`
	NextLink string        `json:"@odata.nextLink,omitempty"`
	Value    []MessageRule `json:"value"`
}

// ListRulesOptions controla el listado.
type ListRulesOptions struct {
	// Top limita el tamaño de página (1-1000). Default: 100.
	Top int
}

// ListRules devuelve las reglas del inbox. Endpoint scoped a inbox
// (no mailbox-wide — el inbox es donde viven las reglas activas).
func (c *Client) ListRules(ctx context.Context, bearer string, opts ListRulesOptions) (*RuleList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if opts.Top <= 0 || opts.Top > 1000 {
		opts.Top = 100
	}

	endpoint := "/me/mailFolders/inbox/messageRules"
	q := url.Values{}
	q.Set("$top", fmt.Sprintf("%d", opts.Top))
	full := c.BaseURL + endpoint + "?" + q.Encode()

	resp := &RuleList{}
	if err := c.doGet(ctx, bearer, full, resp); err != nil {
		return nil, err
	}
	return resp, nil
}