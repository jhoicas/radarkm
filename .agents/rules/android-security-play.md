---
description: Cumplimiento Play Store y seguridad Android de RadarKm
globs:
  - "apps/android/**"
alwaysApply: false
---

# Android security & Play

1. `compileSdk` / `targetSdk` = **36** (Android 16). `minSdk` = 26.
2. Play Billing Library **8+** únicamente para pagos in-app. Prohibido incluir enlaces a pasarelas externas dentro del APK.
3. Librerías NDK (`.so`) alineadas a páginas de **16 KB** (`-Wl,-z,max-page-size=16384`; NDK r28+ preferido; en r27 añadir `-DANDROID_SUPPORT_FLEXIBLE_PAGE_SIZES=ON`).
4. Prohibido `android:isAccessibilityTool="true"` en el manifiesto del AccessibilityService.
5. Prohibido declarar o solicitar `REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`.
6. Foreground services de captura/overlay usan `foregroundServiceType="specialUse"` y bandera de entorno seguro cuando aplique (`REQUIRE_SECURE_ENV`).
7. Obligatorio reciclar `Bitmap` en memoria; nunca escribir capturas de ofertas a disco.
