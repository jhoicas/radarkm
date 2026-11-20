# 01 — Product Spec RadarKm

## Resumen

Copiloto para conductores de Uber, DiDi, inDrive, Cabify y Picap en Colombia: lee ofertas localmente, calcula rentabilidad neta y contraoferta (inDrive), con cuota gratuita backend y dashboard Super Admin.

## Requisitos funcionales y criterios de aceptación

### CA-1 — Captura local de ofertas
El sistema lee ofertas vía Accessibility; si falla, usa `takeScreenshot()` + ML Kit Text Recognition v2 en RAM sin subir la captura.
**Verificable:** prueba de instrumentación/manual que confirma parseo y ausencia de upload de bitmap.

### CA-2 — Rentabilidad neta COP
Calcula neto por km y por hora con costos configurables (combustible, comisión, tiempo) usando solo enteros COP.
**Verificable:** tests unitarios en `:core:finance` con fixtures en `Long`.

### CA-3 — Contraoferta inDrive
Sugiere contraoferta cuando el semáforo (AND de umbrales) lo permite, en enteros COP.
**Verificable:** test unitario de `SemaphoreLogic` + motor de contraoferta.

### CA-4 — Cuota gratuita 2/día
Máximo 2 servicios gratis por celular/día (`America/Bogota`), autoridad 100% backend Redis.
**Verificable:** test de API de cuota + intento local no incrementa ni autoriza solo.

### CA-5 — Antipiratería
Play Integrity, Device Recall, Widevine MediaDrm hash, detección de clonadores y chequeos NDK C++.
**Verificable:** contratos/interfaces + endpoint integrity rechaza tokens inválidos en tests.

### CA-6 — Selectores OTA
Selectores JSON de lectura de pantalla configurables desde Admin y consumidos por la app firmados.
**Verificable:** round-trip Admin → API → cliente con firma Ed25519 válida.

### CA-7 — RadarKm Admin
Dashboard Super Admin: tarifas/cuotas, selectores OTA, consola antipiratería, tesorería.
**Verificable:** rutas Admin renderizan y typecheck pasa (`tsc --noEmit`).
