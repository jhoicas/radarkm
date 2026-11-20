package com.radarkm.core.network.dto

import kotlinx.serialization.Serializable

@Serializable
data class OtaSelectorDto(
    val platform: String,
    val version: Long,
    val packageName: String,
    val selectorJson: String,
    val signatureEd25519Base64: String,
)

@Serializable
data class OtaSelectorsResponseDto(
    val selectors: List<OtaSelectorDto>,
)
