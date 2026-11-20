package com.radarkm.core.security

import android.media.MediaDrm
import android.media.UnsupportedSchemeException
import java.security.MessageDigest
import java.util.UUID

/**
 * Hash SHA-256 del identificador Widevine vía MediaDrm.
 * El valor se envía al backend; no autoriza cuota ni premium en el cliente.
 */
class WidevineIdProvider {

    fun widevineDeviceHashSha256(): Result<ByteArray> =
        runCatching {
            val drm = MediaDrm(WIDEVINE_UUID)
            try {
                val uniqueId = drm.getPropertyByteArray(MediaDrm.PROPERTY_DEVICE_UNIQUE_ID)
                MessageDigest.getInstance("SHA-256").digest(uniqueId)
            } finally {
                runCatching { drm.close() }
            }
        }.recoverCatching { error ->
            when (error) {
                is UnsupportedSchemeException ->
                    throw WidevineUnavailableException("Widevine MediaDrm not supported", error)
                else -> throw WidevineUnavailableException("Widevine id unavailable", error)
            }
        }

    companion object {
        private val WIDEVINE_UUID: UUID = UUID(-0x121074568629b532L, -0x5c37d8232ae2de13L)
    }
}

class WidevineUnavailableException(
    message: String,
    cause: Throwable? = null,
) : Exception(message, cause)
