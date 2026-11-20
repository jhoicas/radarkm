---
name: add-ota-selector
description: Plantilla y procedimiento para agregar/probar selectores JSON OTA de una app de transporte en RadarKm.
---

# Skill: add-ota-selector

## Objetivo

Añadir o actualizar el selector JSON OTA de una app (Uber, DiDi, inDrive, Cabify, Picap) sin hardcodear selectores solo en el cliente.

## Plantilla JSON

```json
{
  "platform": "uber",
  "version": 1,
  "packageName": "com.ubercab.driver",
  "fields": {
    "fare": { "viewId": "...", "textRegex": "\\$?[0-9.,]+" },
    "distanceKm": { "viewId": "...", "textRegex": "[0-9.,]+\\s*km" },
    "durationMin": { "viewId": "...", "textRegex": "[0-9]+\\s*min" },
    "pickup": { "viewId": "..." },
    "dropoff": { "viewId": "..." }
  },
  "ocrHints": ["COP", "km", "min"]
}
```

## Procedimiento

1. Confirmar plataforma y `packageName` reales.
2. Crear/actualizar selector en admin (`ota-selectors`) → API `internal/ota` (fuente de verdad).
3. Cliente: descargar vía `:core:network` DTO `OtaSelectorDto`, verificar firma Ed25519.
4. Parser OTA en `:feature:capture` consume el JSON; Accessibility primero, OCR v2 como fallback en RAM.
5. Probar en dispositivo/emulador con la app de transporte objetivo.
6. Documentar versión del selector y criterios de aceptación en el commit (sin tocar `/specs` salvo aprobación).

## Checklist

- [ ] JSON válido y versionado
- [ ] Firma Ed25519 verificada en cliente
- [ ] Sin persistir bitmaps
- [ ] `Check:` de la tarea OTA en verde
