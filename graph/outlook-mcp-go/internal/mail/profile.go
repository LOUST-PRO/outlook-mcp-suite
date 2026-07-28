// Package mail — profile.go: identidad del usuario vía /me.
//
// /me en Graph devuelve el resource del usuario autenticado. Esto cubre
// lo que normalmente se ve en "Cuenta" del cliente Outlook: nombre
// visible, UPN, correo principal, puesto, teléfonos, ubicación.
package mail

import (
	"context"
	"errors"
)

// UserProfile es el subset expuesto por get_user_profile. Solo campos
// seguros de incluir en una respuesta MCP (sin números de contacto
// sensibles ni metadata de tenant).
type UserProfile struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"displayName,omitempty"`
	GivenName         string   `json:"givenName,omitempty"`
	Surname           string   `json:"surname,omitempty"`
	UserPrincipalName string   `json:"userPrincipalName,omitempty"`
	Mail              string   `json:"mail,omitempty"`
	JobTitle          string   `json:"jobTitle,omitempty"`
	MobilePhone       string   `json:"mobilePhone,omitempty"`
	BusinessPhones    []string `json:"businessPhones,omitempty"`
	OfficeLocation    string   `json:"officeLocation,omitempty"`
	PreferredLanguage string   `json:"preferredLanguage,omitempty"`
}

// GetUserProfile devuelve la identidad del usuario autenticado. El
// locale opcional (Accept-Language) se aplica solo a esta request.
func (c *Client) GetUserProfile(ctx context.Context, bearer, locale string) (*UserProfile, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}

	full := c.BaseURL + "/me"
	var out UserProfile
	if err := c.doGetWithLocale(ctx, bearer, full, locale, &out); err != nil {
		return nil, err
	}
	return &out, nil
}