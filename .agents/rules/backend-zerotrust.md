---
description: Zero-Trust backend — cuota Redis, Play Integrity, Ed25519
globs:
  - "services/api/**"
alwaysApply: false
---

# Backend Zero-Trust

1. **Autoridad de cuota:** Redis es la fuente de verdad de la cuota diaria gratuita, con zona horaria **`America/Bogota`** (día civil Bogotá, no UTC del cliente).
2. **Play Integrity:** toda solicitud sensible debe validar tokens de Play Integrity en el servidor antes de conceder cuota o premium.
3. **Respuestas firmadas:** payloads autoritativos (cuota, selectores OTA, flags antipiratería) se firman con **Ed25519**; el cliente verifica la firma antes de confiar.
4. No confiar en claims del cliente sobre device id, reloj local o contadores de uso.
5. Secretos solo vía variables de entorno / secret manager; nunca hardcodeados.
