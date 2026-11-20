-- RadarKm initial schema. Monetary columns are BIGINT (COP integers).

CREATE TABLE IF NOT EXISTS devices (
    id              UUID PRIMARY KEY,
    widevine_hash   TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id                  UUID PRIMARY KEY,
    device_id           UUID NOT NULL REFERENCES devices(id),
    play_purchase_token TEXT NOT NULL,
    active              BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tariffs (
    id                  UUID PRIMARY KEY,
    free_daily_quota    BIGINT NOT NULL DEFAULT 2,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_json        JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS ota_selectors (
    id              UUID PRIMARY KEY,
    platform        TEXT NOT NULL,
    version         BIGINT NOT NULL,
    selector_json   JSONB NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integrity_events (
    id          UUID PRIMARY KEY,
    device_id   UUID REFERENCES devices(id),
    verdict     TEXT NOT NULL,
    raw_claims  JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
