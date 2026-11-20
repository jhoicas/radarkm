-- RadarKm T-01 schema
-- Money: BIGINT COP only. Timestamps: TIMESTAMPTZ (UTC storage).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- users (Ley 1581 deletion flag)
CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT,
    display_name      TEXT,
    role              TEXT NOT NULL CHECK (role IN ('driver', 'admin', 'super_admin')),
    status            TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'suspended', 'deleted')),
    deleted_ley_1581  BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- devices (aligned with specs/02-architecture-plan.md + antipiratería)
CREATE TABLE devices (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID REFERENCES users(id),
    widevine_hash           TEXT NOT NULL,
    play_integrity_verdict  TEXT,
    device_recall_bits      BIGINT NOT NULL DEFAULT 0,
    status                  TEXT NOT NULL DEFAULT 'active'
                            CHECK (status IN ('active', 'blocked', 'revoked')),
    blocked                 BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT devices_widevine_hash_unique UNIQUE (widevine_hash)
);

-- system_config (cuota diaria gratuita + overrides)
CREATE TABLE system_config (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    free_daily_quota     BIGINT NOT NULL DEFAULT 2
                         CHECK (free_daily_quota >= 0),
    quota_counting_mode  TEXT NOT NULL DEFAULT 'per_device'
                         CHECK (quota_counting_mode IN ('per_device', 'per_user')),
    min_version_code     BIGINT NOT NULL DEFAULT 1
                         CHECK (min_version_code >= 0),
    city_overrides       JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- tariffs (nombre exacto del architecture plan)
CREATE TABLE tariffs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    free_daily_quota  BIGINT NOT NULL DEFAULT 2
                      CHECK (free_daily_quota >= 0),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_json      JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- plans
CREATE TABLE plans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    plan_type   TEXT NOT NULL CHECK (plan_type IN ('pass', 'subscription', 'credits')),
    price_cop   BIGINT NOT NULL CHECK (price_cop >= 0),
    play_sku    TEXT,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    city_code   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- subscriptions (architecture plan)
CREATE TABLE subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id            UUID NOT NULL REFERENCES devices(id),
    play_purchase_token  TEXT NOT NULL,
    active               BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at           TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- entitlements
CREATE TABLE entitlements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),
    user_id     UUID REFERENCES users(id),
    plan_id     UUID REFERENCES plans(id),
    status      TEXT NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'expired', 'revoked')),
    starts_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- transactions (play_billing | wompi | bre_b)
CREATE TABLE transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id            UUID REFERENCES devices(id),
    user_id              UUID REFERENCES users(id),
    plan_id              UUID REFERENCES plans(id),
    channel              TEXT NOT NULL
                         CHECK (channel IN ('play_billing', 'wompi', 'bre_b')),
    status               TEXT NOT NULL
                         CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    external_reference   TEXT NOT NULL,
    amount_cop           BIGINT NOT NULL CHECK (amount_cop >= 0),
    dian_invoice_status  TEXT NOT NULL DEFAULT 'not_required'
                         CHECK (dian_invoice_status IN ('not_required', 'pending', 'issued', 'failed')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transactions_external_reference_unique UNIQUE (external_reference)
);

-- ota_selectors (architecture: selector_json; status OTA lifecycle)
CREATE TABLE ota_selectors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform        TEXT NOT NULL
                    CHECK (platform IN ('uber', 'didi', 'indrive', 'cabify', 'picap')),
    version         BIGINT NOT NULL CHECK (version >= 0),
    selector_json   JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'canary', 'active', 'rollback')),
    active          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- integrity_events (architecture plan)
CREATE TABLE integrity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID REFERENCES devices(id),
    verdict     TEXT NOT NULL,
    raw_claims  JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- audit_logs (inmutable Super Admin)
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id   UUID REFERENCES users(id),
    action          TEXT NOT NULL,
    entity_type     TEXT NOT NULL,
    entity_id       UUID,
    payload_before  JSONB,
    payload_after   JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION audit_logs_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE PROCEDURE audit_logs_immutable();

CREATE TRIGGER audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW EXECUTE PROCEDURE audit_logs_immutable();

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES ('000001_t01_schema')
ON CONFLICT (version) DO NOTHING;
