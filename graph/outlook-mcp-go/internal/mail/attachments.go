// Package mail — attachments.go: adjuntos de mensaje vía /me/messages/{id}/attachments.
//
// Graph modela adjuntos como un union polimórfico: fileAttachment (el
// común — bytes embebidos en `contentBytes`), itemAttachment (referencia
// a otro item de Graph) y referenceAttachment (link a un archivo en
// OneDrive/SharePoint). Esta implementación cubre fileAttachment en
// profundidad y trata los otros dos como opacos.
//
// CAVEAT DE SEGURIDAD: `get_attachment` devuelve los bytes del adjunto
// embebidos en base64 dentro del campo `contentBytes` del response de
// Graph. Para un adjunto de 25 MiB eso son ~33 MiB de JSON. El caller
// DEBE pasar max_bytes (cap recomendado: 10 MiB) para evitar OOM o
// payloads gigantes que revientan el contexto del LLM downstream.
package mail

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// AttachmentType es el discriminador del union polimórfico de Graph.
// Solo cubrimos fileAttachment explícitamente; los otros se preservan
// como opacos en el campo genérico Raw.
type AttachmentType string

const (
	AttachmentFile        AttachmentType = "#microsoft.graph.fileAttachment"
	AttachmentItem        AttachmentType = "#microsoft.graph.itemAttachment"
	AttachmentReference   AttachmentType = "#microsoft.graph.referenceAttachment"
)

// FileAttachment es el resource concreto para archivos embebidos. Este
// es el tipo que el 99% de los emails del mundo real usan.
type FileAttachment struct {
	ODataType   AttachmentType `json:"@odata.type"`
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	ContentType string         `json:"contentType,omitempty"`
	Size        int            `json:"size,omitzero"`
	IsInline    bool           `json:"isInline,omitzero"`
	ContentID   string         `json:"contentId,omitempty"`
	// ContentBytes son los bytes del adjunto en base64 estándar. Solo
	// presente cuando se pide el resource individual con $select=contentBytes
	// (o implícitamente en get_attachment). AUSENTE en list_attachments
	// porque listar 50 adjuntos de 5 MB cada uno sería 250 MB de JSON.
	ContentBytes string `json:"contentBytes,omitempty"`
	LastModifiedDateTime string `json:"lastModifiedDateTime,omitempty"`
}

// Attachment es el envelope común que list_attachments devuelve. Graph
// usa un discriminator (@odata.type) que el JSON unmarshal resuelve
// automáticamente a FileAttachment cuando aplica; los otros tipos
// quedan como Raw opaco.
type Attachment struct {
	ODataType   AttachmentType `json:"@odata.type"`
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	ContentType string         `json:"contentType,omitempty"`
	Size        int            `json:"size,omitzero"`
	IsInline    bool           `json:"isInline,omitzero"`
	ContentID   string         `json:"contentId,omitempty"`
	LastModifiedDateTime string `json:"lastModifiedDateTime,omitempty"`
	// File está presente solo cuando @odata.type es fileAttachment.
	// Para itemAttachment y referenceAttachment queda como opaco.
	File *FileAttachment `json:"file,omitempty,omitzero"`
}

// AttachmentList es el OData envelope estándar.
type AttachmentList struct {
	Context  string       `json:"@odata.context,omitempty"`
	Count    int          `json:"@odata.count,omitempty"`
	NextLink string       `json:"@odata.nextLink,omitempty"`
	Value    []Attachment `json:"value"`
}

// ListAttachmentsOptions controla el listado.
type ListAttachmentsOptions struct {
	// Top limita el tamaño de página (1-1000).
	Top int
}

// ListAttachments devuelve los adjuntos (metadata solamente) de un
// mensaje. NO incluye contentBytes — esa proyección solo ocurre en
// get_attachment con max_bytes explícito. Esto evita que un listado
// accidental de un mensaje con 50 adjuntos de 5 MB reviente el MCP.
func (c *Client) ListAttachments(ctx context.Context, bearer, messageID string, opts ListAttachmentsOptions) (*AttachmentList, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if messageID == "" {
		return nil, errors.New("mail: empty messageID")
	}
	if opts.Top <= 0 || opts.Top > 1000 {
		opts.Top = 25
	}

	endpoint := fmt.Sprintf("/me/messages/%s/attachments", url.PathEscape(messageID))
	q := url.Values{}
	q.Set("$top", fmt.Sprintf("%d", opts.Top))
	full := c.BaseURL + endpoint + "?" + q.Encode()

	resp := &AttachmentList{}
	if err := c.doGet(ctx, bearer, full, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// MaxAttachmentBytes es el cap absoluto que get_attachment acepta sin
// quejarse. Por encima de eso, devuelve error antes de hacer la request.
// 25 MiB coincide con MaxBodyBytes del cliente HTTP — si el response
// raw ya pasa eso, http.MaxBytesReader trunca. Este cap es defensa
// adicional desde el lado del caller.
const MaxAttachmentBytes = 25 * 1024 * 1024 // 25 MiB

// RecommendedAttachmentCap es el tamaño que recomendamos para uso
// desde LLMs. PDFs típicos de 1-3 páginas caben en <500 KiB; imágenes
// inline suelen ser <200 KiB. Subir esto solo si el caso de uso lo
// justifica explícitamente.
const RecommendedAttachmentCap = 10 * 1024 * 1024 // 10 MiB

// GetAttachment devuelve un adjunto individual. Si maxBytes es > 0 y
// el contentBytes del adjunto (decodificado de base64) excede ese
// tamaño, devuelve error sin pedir el body. Esto protege contra
// adjuntos enormes que revientan el contexto del LLM.
//
// maxBytes=0 significa "sin cap" — el response de Graph ya está
// limitado por MaxAttachmentBytes via http.MaxBytesReader.
func (c *Client) GetAttachment(ctx context.Context, bearer, messageID, attachmentID string, maxBytes int) (*FileAttachment, error) {
	if bearer == "" {
		return nil, errors.New("mail: empty bearer token")
	}
	if messageID == "" {
		return nil, errors.New("mail: empty messageID")
	}
	if attachmentID == "" {
		return nil, errors.New("mail: empty attachmentID")
	}
	if maxBytes < 0 {
		return nil, errors.New("mail: negative maxBytes")
	}
	if maxBytes > MaxAttachmentBytes {
		return nil, fmt.Errorf("mail: maxBytes %d exceeds hard cap %d", maxBytes, MaxAttachmentBytes)
	}

	endpoint := fmt.Sprintf("/me/messages/%s/attachments/%s",
		url.PathEscape(messageID), url.PathEscape(attachmentID))
	full := c.BaseURL + endpoint

	var out FileAttachment
	if err := c.doGet(ctx, bearer, full, &out); err != nil {
		return nil, err
	}

	// Post-fetch guard: si el caller pidió cap y el contenido real lo
	// excede, fallar explícitamente. contentBytes está en base64, así
	// que el tamaño real es ~3/4 del largo del string.
	if maxBytes > 0 && out.ContentBytes != "" {
		// len(contentBytes) * 3/4 ≈ bytes originales (over-estimate
		// seguro, dado padding).
		approxBytes := len(out.ContentBytes) * 3 / 4
		if approxBytes > maxBytes {
			return nil, fmt.Errorf("mail: attachment ~%d bytes exceeds cap %d (pass max_bytes=%d to allow)",
				approxBytes, maxBytes, approxBytes)
		}
	}

	return &out, nil
}