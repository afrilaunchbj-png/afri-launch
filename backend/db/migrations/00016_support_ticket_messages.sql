-- +goose Up
-- Fil de discussion des tickets de support : réponses utilisateur et support.
-- Le message initial du ticket reste dans support_tickets.message ;
-- support_ticket_messages porte uniquement les échanges suivants.

CREATE TABLE support_ticket_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id  UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    is_admin   BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_ticket_messages_ticket ON support_ticket_messages (ticket_id, created_at ASC);
CREATE INDEX idx_support_ticket_messages_author ON support_ticket_messages (author_id);

-- +goose Down
DROP TABLE IF EXISTS support_ticket_messages;
