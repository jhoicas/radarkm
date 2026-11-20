-- RadarKm T-01 schema rollback (clean)

DROP TRIGGER IF EXISTS audit_logs_no_delete ON audit_logs;
DROP TRIGGER IF EXISTS audit_logs_no_update ON audit_logs;
DROP FUNCTION IF EXISTS audit_logs_immutable();

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS integrity_events;
DROP TABLE IF EXISTS ota_selectors;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS entitlements;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS tariffs;
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;
