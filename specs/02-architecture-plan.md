# 02 — Architecture Plan RadarKm

## Android (módulos)

```text
:app                  → Application, Nav, DI raíz, Manifest
:core:finance         → JVM puro — MoneyCop Long, rentabilidad, semáforo
:core:security        → Play Integrity, Widevine, path /data/user/0/, Ed25519, NDK radarkm_security
:core:network         → Ktor + OkHttp engine, CertificatePinner, DTOs cuota/OTA
:feature:capture      → AccessibilityService, OtaParser, OCR fallback
:feature:overlay      → TYPE_ACCESSIBILITY_OVERLAY, TTS, Prominent Disclosure
:feature:billing      → Play Billing 8+, config vehículo/rentabilidad
```

Dependencias: features → core; `:app` → features + core. Features no se acoplan entre sí.

## Contratos REST principales (API Go)

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/health` | Liveness |
| POST | `/v1/quota/consume` | Consume 1 unidad de cuota (Integrity + Redis Bogotá) |
| GET | `/v1/quota/status` | Estado de cuota del device (firmado Ed25519) |
| POST | `/v1/integrity/verify` | Valida Play Integrity / señales device |
| GET | `/v1/ota/selectors` | Selectores OTA vigentes (firmados) |
| PUT | `/v1/admin/ota/selectors` | Admin actualiza selectores |
| PUT | `/v1/admin/tariffs` | Admin tarifas/cuota diaria |
| GET | `/v1/admin/treasury` | Resumen tesorería |

## Esquema base PostgreSQL

- `devices(id, widevine_hash, created_at, status)`
- `subscriptions(id, device_id, play_purchase_token, active, expires_at)`
- `tariffs(id, free_daily_quota, updated_at, payload_json)`
- `ota_selectors(id, platform, version, selector_json, active)`
- `integrity_events(id, device_id, verdict, raw_claims, created_at)`
- Montos: columnas `BIGINT` en COP.

## Redis

- Clave cuota: `quota:{deviceId}:{yyyy-MM-dd}` con TZ `America/Bogota`.
- TTL hasta fin del día Bogotá + margen.
- Contador atómico; nunca confiar en contador cliente.

## Admin Web

Next.js App Router — secciones: Tarifas y Cuotas, Selectores OTA, Consola Anti-Piratería, Tesorería.
