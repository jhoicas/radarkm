# 00 — Constitución RadarKm

Invariantes innegociables. Toda implementación debe cumplirlas.

## Dinero

- Todos los montos en COP se representan como enteros: `Long` / `int64` / `BIGINT`.
- Prohibido `Float`/`Double` (y equivalentes) para cálculos o persistencia de dinero.
- Redondeos se definen explícitamente en el motor financiero (hacia abajo/arriba según regla de negocio documentada).

## Zero-Trust

- La cuota gratuita diaria y el estado premium son **autoritativos solo en el backend**.
- Redis + zona `America/Bogota` gobiernan el día de cuota.
- Play Integrity + Device Recall + señales Widevine/NDK informan decisiones servidor.
- Respuestas autoritativas firmadas Ed25519; el cliente verifica antes de confiar.

## Privacidad

- Lectura de ofertas: Accessibility tree + fallback `takeScreenshot()` + ML Kit OCR v2 **en RAM**.
- Prohibido subir capturas de pantalla de ofertas a servidores.
- Reciclar `Bitmap`; no escribir capturas de ofertas a disco.

## Play Store

- `targetSdk`/`compileSdk` 36; Billing 8+; NDK 16 KB page size.
- Sin `android:isAccessibilityTool="true"`.
- Sin `REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`.
- Sin enlaces a pasarelas de pago externas en el APK.
- Prominent Disclosure obligatorio antes de activar Accessibility.

## Specs y agente

- `/specs` no se modifica sin aprobación humana.
- 1 tarea = 1 commit + verificación en verde (`AGENTS.md`).
