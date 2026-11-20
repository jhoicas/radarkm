package com.radarkm.core.network.dto

import kotlinx.serialization.Serializable

@Serializable
data class QuotaStatusDto(
    val deviceId: String,
    val dayBogota: String,
    val used: Long,
    val limit: Long,
    val remaining: Long,
    val signatureEd25519Base64: String,
)

@Serializable
data class QuotaConsumeRequestDto(
    val deviceId: String,
    val integrityToken: String,
)
