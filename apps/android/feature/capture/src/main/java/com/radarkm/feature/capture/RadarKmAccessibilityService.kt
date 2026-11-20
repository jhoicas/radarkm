package com.radarkm.feature.capture

import android.accessibilityservice.AccessibilityService
import android.view.accessibility.AccessibilityEvent

/**
 * Esqueleto del servicio de accesibilidad.
 * Sin `isAccessibilityTool="true"`. Captura solo local.
 */
class RadarKmAccessibilityService : AccessibilityService() {

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        // Scaffolding: parsing OTA will be wired in later tasks.
    }

    override fun onInterrupt() {
        // No-op scaffolding
    }
}
