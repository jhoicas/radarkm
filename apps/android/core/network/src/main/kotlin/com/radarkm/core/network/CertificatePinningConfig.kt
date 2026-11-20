package com.radarkm.core.network

import okhttp3.CertificatePinner

/**
 * Configuración de Certificate Pinning lista para pins reales (sin secretos en repo).
 * Sustituir `sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=` por pins de producción vía build config / remote config firmada.
 */
data class CertificatePinningConfig(
    val hostname: String,
    val pinsSha256: List<String>,
) {
    fun toCertificatePinner(): CertificatePinner {
        require(pinsSha256.isNotEmpty()) { "At least one pin is required" }
        val builder = CertificatePinner.Builder()
        pinsSha256.forEach { pin ->
            builder.add(hostname, pin)
        }
        return builder.build()
    }

    companion object {
        fun scaffoldingPlaceholder(hostname: String = "api.radarkm.local"): CertificatePinningConfig =
            CertificatePinningConfig(
                hostname = hostname,
                pinsSha256 = listOf("sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="),
            )
    }
}
