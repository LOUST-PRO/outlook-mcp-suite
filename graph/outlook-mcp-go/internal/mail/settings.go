// Package mail — settings.go: configuración de buzón vía /me/mailboxSettings.
//
// mailboxSettings cubre automaticReplies (fuera de oficina), timeZone,
// workingHours, formato de fecha/hora, idioma y delegado de reuniones.
// Solo lectura en esta tanda.
package mail

import (
	"context"
	"errors"
)

// AutomaticRepliesSetting describe el estado y mensaje de fuera de oficina.
// Graph lo devuelve anidado dentro de mailboxSettings.
type AutomaticRepliesSetting struct {
	Status           string `json:"status,omitempty"` // disabled | alwaysEnabled | scheduled
	ExternalAudience string `json:"externalAudience,omitempty"`
	ScheduledStart   string `json:"scheduledStartDateTime,omitempty,omitzero"`
	ScheduledEnd     string `json:"scheduledEndDateTime,omitempty,omitzero"`
	InternalMessage  string `json:"internalReplyMessage,omitempty"`
	ExternalMessage  string `json:"externalReplyMessage,omitempty"`
}

// WorkingHours describe la jornada laboral del usuario.
type WorkingHours struct {
	DaysOfWeek  []string `json:"daysOfWeek,omitempty"`
	StartTime   string   `json:"startTime,omitempty"` // "HH:MM:SS"
	EndTime     string   `json:"endTime,omitempty"`
	TimeZone    *TimeZone `json:"timeZone,omitempty,omitzero"`
}

// TimeZone es el subset relevante del resource timeZoneBase de Graph.
type TimeZone struct {
	Name string `json:"name,omitempty"`
}

// MailboxSettings es el subset expuesto por get_mailbox_settings.
type MailboxSettings struct {
	AutomaticReplies                            AutomaticRepliesSetting `json:"automaticReplies,omitzero"`
	TimeZone                                    *TimeZone               `json:"timeZone,omitzero"`
	WorkingHours                                *WorkingHours           `json:"workingHours,omitzero"`
	Language                                    *TimeZone               `json:"language,omitzero"`
	DateFormat                                  string                  `json:"dateFormat,omitempty"`
	TimeFormat                                  string                  `json:"timeFormat,omitempty"`
	DelegateMeetingMessageDeliveryOptions       string                  `json:"delegateMeetingMessageDeliveryOptions,omitempty"`
	AllowExternalSender                         string                  `json:"allowExternalSender,omitempty"`
}

// GetMailboxSettings devuelve la configuración del buzón. El locale
// opcional (Accept-Language) se aplica solo a esta request.
func (c *Client) GetMailboxSettings(ctx context.Context, bearer, locale string) (*MailboxSettings, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}

	full := c.BaseURL + "/me/mailboxSettings"
	var out MailboxSettings
	if err := c.doGetWithLocale(ctx, bearer, full, locale, &out); err != nil {
		return nil, err
	}
	return &out, nil
}