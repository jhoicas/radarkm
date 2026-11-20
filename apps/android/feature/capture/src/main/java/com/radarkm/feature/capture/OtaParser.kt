package com.radarkm.feature.capture

import android.view.accessibility.AccessibilityNodeInfo
import com.radarkm.core.network.dto.OtaSelectorDto

/**
 * Contrato del parser OTA impulsado por selectores JSON firmados.
 */
interface OtaParser {
    fun parse(root: AccessibilityNodeInfo, selector: OtaSelectorDto): ParsedOffer?
}

data class ParsedOffer(
    val fareCop: Long,
    val distanceMeters: Long,
    val durationMinutes: Long,
    val platform: String,
)

class OtaParserStub : OtaParser {
    override fun parse(root: AccessibilityNodeInfo, selector: OtaSelectorDto): ParsedOffer? = null
}
