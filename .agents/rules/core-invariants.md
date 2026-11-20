---
description: Invariantes centrales de RadarKm (dinero COP, Zero-Trust de cuota, specs)
globs:
  - "**/*"
alwaysApply: true
---

# Core invariants

1. **Dinero en COP:** todo monto monetario usa `Long` (Kotlin), `int64` (Go) o `BIGINT` (PostgreSQL). Prohibido `Float`, `Double`, `float32`, `float64` o tipos decimales flotantes para dinero.
2. **Cuota / premium:** la autoridad de cuota gratuita y estado premium es **solo el backend**. Prohibido lógica premium o de cuota autoritativa local en el APK (UI puede cachear respuestas firmadas, nunca decidir sola).
3. **Specs:** prohibido modificar archivos bajo `/specs` sin aprobación explícita del humano responsable.
4. **Privacidad:** capturas y OCR permanecen en RAM del dispositivo; no persistir ni subir bitmaps de ofertas.
