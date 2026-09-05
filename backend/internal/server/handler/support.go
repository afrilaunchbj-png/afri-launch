package handler

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"afrilaunch/backend/internal/application/support"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/server/authctx"
)

// SupportHandler expose les demandes d'assistance utilisateur.
type SupportHandler struct {
	svc *support.Service
}

// NewSupportHandler construit le handler de support.
func NewSupportHandler(svc *support.Service) *SupportHandler {
	return &SupportHandler{svc: svc}
}

type ticketDTO struct {
	ID          string          `json:"id"`
	Subject     string          `json:"subject"`
	Message     string          `json:"message"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	Attachments []attachmentDTO `json:"attachments,omitempty"`
}

func toTicketDTO(t domain.SupportTicket) ticketDTO {
	return ticketDTO{ID: t.ID, Subject: t.Subject, Message: t.Message, Status: t.Status, CreatedAt: t.CreatedAt}
}

// attachmentDTO sérialise une pièce jointe (jamais la clé de stockage).
type attachmentDTO struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	IsImage     bool   `json:"is_image"`
}

func toAttachmentDTO(a domain.SupportAttachment) attachmentDTO {
	return attachmentDTO{
		ID: a.ID, Filename: a.Filename, ContentType: a.ContentType,
		SizeBytes: a.SizeBytes, IsImage: isImageType(a.ContentType),
	}
}

func isImageType(ct string) bool {
	return ct == "image/png" || ct == "image/jpeg" || ct == "image/webp" || ct == "image/gif"
}

func toAttachmentDTOs(items []domain.SupportAttachment) []attachmentDTO {
	out := make([]attachmentDTO, 0, len(items))
	for _, a := range items {
		out = append(out, toAttachmentDTO(a))
	}
	return out
}

type ticketMessageDTO struct {
	ID          string          `json:"id"`
	AuthorID    string          `json:"author_id"`
	AuthorName  string          `json:"author_name"`
	IsAdmin     bool            `json:"is_admin"`
	Content     string          `json:"content"`
	CreatedAt   time.Time       `json:"created_at"`
	Attachments []attachmentDTO `json:"attachments,omitempty"`
}

func toTicketMessageDTO(m domain.TicketMessageView) ticketMessageDTO {
	return ticketMessageDTO{
		ID: m.ID, AuthorID: m.AuthorID, AuthorName: m.AuthorName,
		IsAdmin: m.IsAdmin, Content: m.Content, CreatedAt: m.CreatedAt,
	}
}

type ticketDetailDTO struct {
	Ticket   ticketDTO          `json:"ticket"`
	Messages []ticketMessageDTO `json:"messages"`
}

type createTicketRequest struct {
	Subject       string   `json:"subject"`
	Message       string   `json:"message"`
	AttachmentIDs []string `json:"attachment_ids"`
}

// Create gère POST /support/tickets.
func (h *SupportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in createTicketRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticket, err := h.svc.Create(r.Context(), authctx.UserID(r.Context()), in.Subject, in.Message, in.AttachmentIDs)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toTicketDTO(ticket))
}

// ListMine gère GET /support/tickets.
func (h *SupportHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListMine(r.Context(), authctx.UserID(r.Context()))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]ticketDTO, 0, len(items))
	for _, t := range items {
		out = append(out, toTicketDTO(t))
	}
	writeData(w, http.StatusOK, out)
}

// GetTicket gère GET /support/tickets/{id} : détail + fil + pièces jointes.
func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	ticketID := chi.URLParam(r, "id")
	ticket, messages, err := h.svc.Detail(r.Context(), userID, ticketID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticketFiles, messageFiles, err := h.svc.DetailAttachments(r.Context(), userID, ticketID)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	out := make([]ticketMessageDTO, 0, len(messages))
	for _, m := range messages {
		dto := toTicketMessageDTO(m)
		dto.Attachments = toAttachmentDTOs(messageFiles[m.ID])
		out = append(out, dto)
	}
	detail := ticketDetailDTO{Ticket: toTicketDTO(ticket), Messages: out}
	detail.Ticket.Attachments = toAttachmentDTOs(ticketFiles)
	writeData(w, http.StatusOK, detail)
}

type replyRequest struct {
	Content       string   `json:"content"`
	AttachmentIDs []string `json:"attachment_ids"`
}

// Reply gère POST /support/tickets/{id}/messages : réponse de l'utilisateur.
func (h *SupportHandler) Reply(w http.ResponseWriter, r *http.Request) {
	var in replyRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeAPIError(w, r, err)
		return
	}
	ticket, msg, err := h.svc.Reply(r.Context(), authctx.UserID(r.Context()), chi.URLParam(r, "id"), in.Content, in.AttachmentIDs)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	dto := toTicketMessageDTO(msg)
	writeData(w, http.StatusCreated, ticketDetailDTO{
		Ticket: toTicketDTO(ticket), Messages: []ticketMessageDTO{dto},
	})
}

// UploadAttachment gère POST /support/attachments (multipart/form-data,
// champ "file", max 4 fichiers de 5 Mo : images ou PDF).
func (h *SupportHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if err := r.ParseMultipartForm(domain.AttachmentMaxSize); err != nil {
		writeAPIError(w, r, domain.ErrInvalidInput)
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 || len(files) > domain.AttachmentMaxPerSubmit {
		writeAPIError(w, r, domain.ErrInvalidInput)
		return
	}
	out := make([]attachmentDTO, 0, len(files))
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			writeAPIError(w, r, domain.ErrInvalidInput)
			return
		}
		data, err := readAllLimited(src, domain.AttachmentMaxSize)
		src.Close()
		if err != nil {
			writeAPIError(w, r, domain.ErrInvalidInput)
			return
		}
		// Un type absent est déduit de l'extension (captures d'écran mobiles).
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = guessContentType(fh.Filename)
		}
		attachment, err := h.svc.UploadAttachment(r.Context(), userID, fh.Filename, contentType, data)
		if err != nil {
			writeAPIError(w, r, err)
			return
		}
		out = append(out, toAttachmentDTO(attachment))
	}
	writeData(w, http.StatusCreated, out)
}

// DownloadAttachment gère GET /support/attachments/{id}/download (propriétaire).
func (h *SupportHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	attachment, data, err := h.svc.DownloadAttachment(r.Context(), userID, chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+attachment.Filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

// AdminDownloadAttachment gère GET /admin/support/attachments/{id}/download
// (superadmin : sans contrainte d'appartenance).
func (h *SupportHandler) AdminDownloadAttachment(w http.ResponseWriter, r *http.Request) {
	attachment, data, err := h.svc.DownloadAttachmentForAdmin(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+attachment.Filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

// readAllLimited lit un flux en bornant la taille.
func readAllLimited(src io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(src, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, domain.ErrInvalidInput
	}
	return data, nil
}

// guessContentType déduit le type MIME d'une extension de fichier.
func guessContentType(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
