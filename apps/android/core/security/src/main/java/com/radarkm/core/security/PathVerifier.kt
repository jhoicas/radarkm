package com.radarkm.core.security

/**
 * Verificación de ruta de datos: reglas Kotlin + puente NDK cuando está disponible.
 */
class PathVerifier(
    private val packageName: String = SecureDataPath.EXPECTED_PACKAGE,
    private val nativeBridge: NativePathBridge? = null,
) {
    fun isExpectedDataPath(absolutePath: String): Boolean {
        if (!SecureDataPath.isLegitimateDataPath(absolutePath, packageName)) {
            return false
        }
        val native = nativeBridge ?: return true
        return runCatching { native.verifyDataPath(absolutePath) }.getOrDefault(true)
    }

    fun evaluateEnvironment(filesDirPath: String, userId: Int?): EnvironmentVerdict =
        CloneDetector(packageName).evaluate(filesDirPath, userId)

    fun tamperProbeOk(): Boolean {
        val native = nativeBridge ?: return true
        return runCatching { native.quickTamperProbe() }.getOrDefault(false)
    }

    companion object {
        /** Usa la librería NDK real en dispositivo/proceso Android. */
        fun withNative(): PathVerifier = PathVerifier(nativeBridge = JniNativePathBridge)
    }
}

interface NativePathBridge {
    fun verifyDataPath(absolutePath: String): Boolean
    fun quickTamperProbe(): Boolean
}

internal object JniNativePathBridge : NativePathBridge {
    override fun verifyDataPath(absolutePath: String): Boolean =
        NativeSecurity.verifyDataPath(absolutePath)

    override fun quickTamperProbe(): Boolean = NativeSecurity.quickTamperProbe()
}
