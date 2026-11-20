package com.radarkm.core.security

/**
 * Verifica que la ruta de datos de la app viva bajo /data/user/0/ (chequeo NDK).
 */
class PathVerifier {
    fun isExpectedDataPath(absolutePath: String): Boolean =
        NativeSecurity.verifyDataPath(absolutePath)

    fun tamperProbeOk(): Boolean = NativeSecurity.quickTamperProbe()
}
