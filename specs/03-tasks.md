# 03 — Tasks (scaffolding + cimientos)

Tareas atómicas. Regla: **1 tarea = 1 commit + Check en verde**.

## T-01 — SDD y reglas de agente
Crear `AGENTS.md`, `.agents/rules/*.md` (plano + YAML), skills, `specs/00–02`.
**Check:** `test -f AGENTS.md && test -f specs/00-constitution.md && test -f .agents/rules/core-invariants.md`

## T-02 — Gradle root + Version Catalog
Configurar `apps/android` con `rootProject.name=RadarKm`, AGP 8.9.1, SDK 36, catalog.
**Check:** `cd apps/android && ./gradlew projects`

## T-03 — :core:finance compilable + test
Motor esqueleto COP `Long` + 1 test unitario verde.
**Check:** `cd apps/android && ./gradlew :core:finance:test`

## T-04 — :core:security + NDK 16 KB
CMake `radarkm_security`, flags 16 KB, interfaces Integrity/Widevine/Ed25519/path.
**Check:** `cd apps/android && ./gradlew :core:security:assembleDebug` (y script 16 KB tras `.so`)

## T-05 — :core:network Ktor OkHttp
Cliente con `ktor-client-okhttp` + stub CertificatePinner + DTOs.
**Check:** `cd apps/android && ./gradlew :core:network:assembleDebug`

## T-06 — Features + :app Manifest compliant
Esqueletos capture/overlay/billing + Manifest sin flags prohibidos.
**Check:** `cd apps/android && ./gradlew :app:assembleDebug`

## T-07 — API Go + Docker
`radarkm/api`, `/health`, compose Postgres+Redis, migrations base.
**Check:** `cd services/api && go test ./... && go build -o bin/server ./cmd/server`

## T-08 — Admin Web typecheck
Next.js RadarKm Admin con 4 secciones base.
**Check:** `cd apps/admin-web && (pnpm exec tsc --noEmit || npx tsc --noEmit)`

## T-09 — Script 16 KB + .env.example
`scripts/check_16kb_alignment.sh` y ejemplos de env sin secretos.
**Check:** `test -f scripts/check_16kb_alignment.sh && test -f .env.example`
