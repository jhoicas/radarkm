---
name: verify-task
description: Flujo para dar una tarea por terminada — tests, lint y revisión contra /specs antes del commit.
---

# Skill: verify-task

Usar **antes** de marcar una tarea como hecha y **antes** del commit único.

## Pasos

1. Identificar la tarea activa en `specs/03-tasks.md` y su línea `Check:`.
2. Revisar diff contra invariantes en `specs/00-constitution.md` y `.agents/rules/*`.
3. Ejecutar el `Check:` de la tarea (comando exacto).
4. Si toca Android JVM puro: `./gradlew :core:finance:test` desde `apps/android`.
5. Si toca API: `cd services/api && go test ./... && go build -o bin/server ./cmd/server`.
6. Si toca Admin: `pnpm exec tsc --noEmit` o `npx tsc --noEmit` (según gestor disponible).
7. Si toca `.so` / NDK: `bash scripts/check_16kb_alignment.sh <ruta-intermedios-cxx>`.
8. Confirmar que no se modificó `/specs` sin aprobación.
9. Solo si todo está verde: un commit atómico. Si falla: corregir y repetir desde el paso 3.

## Criterio de salida

La tarea está terminada únicamente cuando el `Check:` reporta éxito y las reglas de `.agents/rules/` no se violan.
