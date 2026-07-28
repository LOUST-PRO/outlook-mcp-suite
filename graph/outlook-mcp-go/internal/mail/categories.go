// Package mail — categories.go: categorías de mensaje via /me/outlook/masterCategories.
//
// Graph mantiene categorías por usuario (no por mensaje). El resource
// expone id, displayName y color (preset enum). Las categorías se aplican
// luego a mensajes vía el campo categories[] del message resource.
//
// Fase 1 entrega list_categories read-only. Crear / renombrar / borrar
// categorías queda fuera de scope (Fase 2+).
package mail

import (
	"context"
	"errors"
)

// OutlookCategory es el resource de categoría expuesta por Graph.
type OutlookCategory struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
	Color       string `json:"color,omitempty"`
}

// CategoryList es el OData envelope estándar.
type CategoryList struct {
	Context  string             `json:"@odata.context,omitempty"`
	Count    int                `json:"@odata.count,omitempty"`
	NextLink string             `json:"@odata.nextLink,omitempty"`
	Value    []OutlookCategory `json:"value"`
}

// ListCategories devuelve todas las categorías definidas por el usuario.
// El endpoint no acepta $top ni $skip en la mayoría de los tenants (límite
// ~100 categorías por usuario), así que se devuelve la lista completa.
func (c *Client) ListCategories(ctx context.Context, bearer string) (*CategoryList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	full := c.BaseURL + "/me/outlook/masterCategories"

	resp := &CategoryList{}
	if err := c.doGet(ctx, bearer, full, resp); err != nil {
		return nil, err
	}
	return resp, nil
}