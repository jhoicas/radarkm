package com.radarkm.core.security

import android.media.MediaDrm
import android.media.UnsupportedSchemeException
import java.util.UUID

/**
 * Obtiene un hash/identificador Widevine vía MediaDrm (esqueleto).
 */
class WidevineIdProvider {

    fun widevineDeviceHash(): Result<ByteArray> =
        runCatching {
            val drm = MediaDrm(WIDEVINE_UUID)
            try {
                drm.getPropertyByteArray(MediaDrm.PROPERTY_DEVICE_UNIQUE_ID)
            } finally {
                drm.close()
            }
        }.recoverCatching { error ->
            if (error is UnsupportedSchemeException) throw error
            throw error
        }

    companion object {
        private val WIDEVINE_UUID: UUID = UUID(-0x121074568629b532L, -0x5c37d8232ae2de13L)
    }
}
