# Plan de implementación — Scaffolding RadarKm (SDD)

**Estado:** Aprobado (con 3 ajustes menores).  
**Alcance:** cimientos SDD + proyectos compilables.  
**Fuera de alcance:** lógica premium/OCR real, pantallas finales, billing productivo, integridad Play en producción.

## Decisiones fijadas

| Decisión | Elección |
|---|---|
| Cliente HTTP | **Ktor Client** + motor **`ktor-client-okhttp`** + Kotlinx Serialization (pinning vía `okhttp3.CertificatePinner`) |
| Admin Web | **Next.js 15** App Router + React 19 + TypeScript + Tailwind CSS 3.4; **pnpm si está instalado, si no npm** |
| API | **Go 1.23**, `chi` router, `pgx` + Redis, migraciones SQL en `migrations/` |
| NDK | Preferir **NDK r28+** (o el instalado local si es r28+). Si se usa r27: `-DANDROID_SUPPORT_FLEXIBLE_PAGE_SIZES=ON` + `-Wl,-z,max-page-size=16384` |
| Dinero | Solo `Long` / `int64` / `BIGINT` (COP) |

## Ajustes de aprobación

1. **NDK 16 KB:** NDK r28+ preferido; fallback r27 con `ANDROID_SUPPORT_FLEXIBLE_PAGE_SIZES=ON` + linker `max-page-size=16384`.
2. **Network:** `ktor-client-okhttp` explícito para `CertificatePinner`.
3. **Admin:** detectar `pnpm`; fallback a `npm`.

## Árbol de directorios exacto

```text
radarkm/
├── AGENTS.md
├── implementation_plan.md
├── README.md
├── .gitignore
├── .env.example
├── .agents/
│   ├── rules/
│   │   ├── core-invariants.md
│   │   ├── android-security-play.md
│   │   └── backend-zerotrust.md
│   └── skills/
│       ├── verify-task/
│       │   └── SKILL.md
│       └── add-ota-selector/
│           └── SKILL.md
├── specs/
│   ├── 00-constitution.md
│   ├── 01-product-spec.md
│   ├── 02-architecture-plan.md
│   └── 03-tasks.md
├── scripts/
│   └── check_16kb_alignment.sh
├── apps/
│   ├── android/
│   │   ├── settings.gradle.kts
│   │   ├── build.gradle.kts
│   │   ├── gradle.properties
│   │   ├── gradle/
│   │   │   ├── libs.versions.toml
│   │   │   └── wrapper/
│   │   │       └── gradle-wrapper.properties
│   │   ├── app/
│   │   ├── core/finance/
│   │   ├── core/security/
│   │   ├── core/network/
│   │   ├── feature/capture/
│   │   ├── feature/overlay/
│   │   └── feature/billing/
│   └── admin-web/
└── services/
    └── api/
```

## Version Catalog Android

| Artefacto | Versión |
|---|---|
| AGP | `8.9.1` |
| Gradle Wrapper | `8.11.1` |
| Kotlin | `2.1.20` |
| KSP | `2.1.20-2.0.0` |
| compileSdk / targetSdk | `36` |
| minSdk | `26` |
| JDK toolchain | `17` |
| Compose BOM | `2025.04.00` |
| Navigation Compose | `2.8.9` |
| Lifecycle | `2.8.7` |
| Activity Compose | `1.10.1` |
| Hilt | `2.56.2` |
| Coroutines | `1.10.2` |
| Kotlinx Serialization | `1.8.1` |
| Ktor Client | `3.1.2` (engine OkHttp) |
| OkHttp | `4.12.0` |
| Play Billing | `8.0.0` |
| Play Integrity | `1.4.0` |
| ML Kit Text Recognition | `16.0.1` |
| NDK | `28.2.13676358` (r28+ local; fallback r27 + `ANDROID_SUPPORT_FLEXIBLE_PAGE_SIZES`) |
| JUnit 5 | `5.12.2` |

`rootProject.name = "RadarKm"` · `applicationId = "com.radarkm.app"` · paquete base `com.radarkm` · `app_name = "RadarKm"`.

## Backend / Admin

| Artefacto | Versión |
|---|---|
| Go | `1.23` |
| chi | `v5` |
| pgx | `v5` |
| go-redis | `v9` |
| Next.js | `15.x` |
| React | `19.x` |
| TypeScript | `5.x` |
| Tailwind CSS | `3.4.x` |
| PostgreSQL (compose) | `16` |
| Redis (compose) | `7` |

## Comandos de verificación

```bash
# Android — motor financiero
cd apps/android && ./gradlew :core:finance:test

# Android — alineación 16 KB (tras build NDK)
bash scripts/check_16kb_alignment.sh apps/android/core/security/build/intermediates/cxx

# API Go
cd services/api && go test ./... && go build -o bin/server ./cmd/server

# Docker (opcional)
cd services/api && docker compose up -d --build
curl -sf http://localhost:8080/health

# Admin Web (pnpm si existe; si no npm)
cd apps/admin-web
(command -v pnpm >/dev/null && pnpm install && pnpm exec tsc --noEmit) || (npm install && npx tsc --noEmit)
```

## Orden de ejecución post-aprobación

1. Escribir este artefacto (`implementation_plan.md`).
2. SDD: `AGENTS.md`, `.agents/rules/` (plano + YAML frontmatter), skills, `specs/`.
3. Scaffolding Android modular + NDK + script 16 KB.
4. Scaffolding `services/api` + Docker.
5. Scaffolding `apps/admin-web`.
6. Ejecutar checks disponibles y corregir hasta verde.
7. Walkthrough con salida real.
