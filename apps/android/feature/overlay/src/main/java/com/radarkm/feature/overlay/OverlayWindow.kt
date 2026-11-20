package com.radarkm.feature.overlay

import android.content.Context
import android.graphics.PixelFormat
import android.view.WindowManager
import android.accessibilityservice.AccessibilityService

/**
 * Esqueleto de ventana flotante con TYPE_ACCESSIBILITY_OVERLAY.
 */
class OverlayWindow(
    private val context: Context,
) {
    fun buildLayoutParams(): WindowManager.LayoutParams =
        WindowManager.LayoutParams(
            WindowManager.LayoutParams.WRAP_CONTENT,
            WindowManager.LayoutParams.WRAP_CONTENT,
            WindowManager.LayoutParams.TYPE_ACCESSIBILITY_OVERLAY,
            WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE,
            PixelFormat.TRANSLUCENT,
        )

    fun requireAccessibilityContext(): Boolean = context is AccessibilityService
}
