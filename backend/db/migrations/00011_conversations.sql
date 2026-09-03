-- +goose Up
-- Chat conversationnel : conversations multi-tours + messages, liaison des idées.

CREATE TABLE conversations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title          TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','archived')),
    opportunity_id UUID REFERENCES opportunities(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_conversations_user ON conversations (user_id, updated_at DESC);

-- Messages du chat (role user|assistant) ; payload = métadonnées JSON
-- (ids d'idées proposées, sources de recherche…).
CREATE TABLE conversation_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK (role IN ('user','assistant')),
    content         TEXT NOT NULL,
    payload         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_conversation_messages_conv ON conversation_messages (conversation_id, created_at);

-- Idées rattachées à la conversation qui les a produites.
ALTER TABLE product_ideas
    ADD COLUMN conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL;

CREATE INDEX idx_product_ideas_conversation ON product_ideas (conversation_id) WHERE conversation_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_product_ideas_conversation;
ALTER TABLE product_ideas DROP COLUMN IF EXISTS conversation_id;
DROP TABLE IF EXISTS conversation_messages;
DROP TABLE IF EXISTS conversations;
