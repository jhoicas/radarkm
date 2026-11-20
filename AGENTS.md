# RadarKm — AGENTS.md

Principios de arquitectura y operación para agentes que trabajan en este monorepo.

## Producto

**RadarKm** es un copiloto Android para conductores de plataformas (Uber, DiDi, inDrive, Cabify, Picap) en Colombia, con backend Zero-Trust y dashboard **RadarKm Admin**.

## Arquitectura (mapa rápido)

| Área | Ubicación | Rol |
|---|---|---|
| App Android modular | `apps/android/` | Captura local, overlay, billing, seguridad NDK |
| Motor financiero | `:core:finance` | JVM puro; dinero COP en `Long` |
| Seguridad | `:core:security` | Play Integrity, Widevine, Ed25519, NDK 16 KB |
| Red | `:core:network` | Ktor + OkHttp + Certificate Pinning |
| API | `services/api/` | Go; autoridad de cuota en Redis (`America/Bogota`) |
| Admin | `apps/admin-web/` | Next.js — tarifas, OTA, antipiratería, tesorería |
| Specs SDD | `/specs` | Constitución, producto, arquitectura, tareas |

## Principios innegociables

1. **Dinero en enteros COP** — `Long` / `int64` / `BIGINT`. Prohibido `Float`/`Double` para dinero.
2. **Zero-Trust de cuota** — la cuota gratuita (2/día) es autoritativa solo en backend Redis. Nunca lógica premium/cuota autoritativa local.
3. **Privacidad de captura** — Accessibility + OCR en RAM; no subir capturas; reciclar `Bitmap`.
4. **Play Store compliance** — sin `isAccessibilityTool="true"`, sin `REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`, sin pasarelas externas en el APK.
5. **Specs primero** — no modificar `/specs` sin aprobación explícita del humano.
6. **Secretos** — solo `.env.example`; nunca llaves reales en el repo.

## Convenciones de código

- Kotlin + Jetpack Compose + Hilt + Coroutines/Flow en Android.
- Paquete base: `com.radarkm.*`. `applicationId`: `com.radarkm.app`.
- Go limpio: `cmd/`, `internal/{quota,integrity,billing,ota}`, `migrations/`.
- Admin: App Router + TypeScript estricto + Tailwind.
- Commits atómicos, mensajes en imperativo corto (qué/por qué).
- Un módulo = una responsabilidad. Features no dependen entre sí; solo de `core/*`.

## Regla estricta: 1 tarea = 1 commit + verificación en verde

1. Tomar **una** tarea de `specs/03-tasks.md` (o subtarea acordada).
2. Implementar solo ese alcance.
3. Ejecutar el skill `verify-task` (tests/lint/revisión contra `/specs`).
4. El comando `Check:` de la tarea debe quedar **en verde**.
5. Crear **un** commit. No mezclar tareas. No commit sin check verde.
6. Si el check falla: corregir → re-verificar → luego commit.

## Antes de editar

- Leer `specs/00-constitution.md` y la regla aplicable en `.agents/rules/`.
- Para selectores OTA: skill `add-ota-selector`.
- Para cerrar tarea: skill `verify-task`.

## Qué no hacer

- No implementar lógica de negocio completa fuera de la tarea activa.
- No añadir dependencias no listadas en el Version Catalog / `go.mod` / `package.json` sin justificación en el commit.
- No debilitar Certificate Pinning, Integrity ni firmas Ed25519 “por comodidad”.
