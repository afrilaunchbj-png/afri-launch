-- +goose Up
-- Pièces jointes des tickets de support (captures d'écran, PDF).
-- Un fichier est d'abord uploadé (message_id/ticket_id NULL), puis rattaché
-- au ticket ou au message au moment de la soumission — orphelins purgés.
CREATE TABLE support_attachments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ticket_id    UUID REFERENCES support_tickets(id) ON DELETE CASCADE,
    message_id   UUID REFERENCES support_ticket_messages(id) ON DELETE CASCADE,
    filename     TEXT NOT NULL,
    storage_key  TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 5242880),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_attachments_ticket ON support_attachments (ticket_id);
CREATE INDEX idx_support_attachments_message ON support_attachments (message_id);
CREATE INDEX idx_support_attachments_user ON support_attachments (user_id);

-- +goose Down
DROP TABLE IF EXISTS support_attachments;
