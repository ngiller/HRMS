-- +goose Up
-- ============================================================
-- Migration 00045: Push Notification Subscriptions
-- ============================================================
-- Menyimpan subscription endpoint dari browser untuk Web Push API.
-- ============================================================

CREATE TABLE push_subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    endpoint        TEXT NOT NULL,
    p256dh_key      TEXT NOT NULL,
    auth_key        TEXT NOT NULL,
    user_agent      TEXT,
    device_name     VARCHAR(255),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Satu device per endpoint
    UNIQUE(user_id, endpoint)
);

CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id, is_active);
CREATE INDEX idx_push_subscriptions_endpoint ON push_subscriptions(endpoint);

CREATE TRIGGER set_push_subscriptions_updated_at
    BEFORE UPDATE ON push_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS set_push_subscriptions_updated_at ON push_subscriptions;
DROP TABLE IF EXISTS push_subscriptions;
