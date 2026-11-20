package com.radarkm.core.security

enum class EnvironmentRejectReason {
    CLONE_OR_INVALID_PATH,
    NON_ZERO_USER_ID,
}

data class EnvironmentVerdict(
    val isSecure: Boolean,
    val reason: EnvironmentRejectReason? = null,
)

/**
 * Valida entorno de ejecución (ruta filesDir + userId) sin decidir cuota/premium.
 * Las señales se envían al backend; este detector solo clasifica el entorno local.
 */
class CloneDetector(
    private val packageName: String = SecureDataPath.EXPECTED_PACKAGE,
) {
    fun evaluate(filesDirPath: String, userId: Int?): EnvironmentVerdict {
        if (!SecureDataPath.isLegitimateDataPath(filesDirPath, packageName)) {
            return EnvironmentVerdict(isSecure = false, reason = EnvironmentRejectReason.CLONE_OR_INVALID_PATH)
        }
        if (userId != null && userId != 0) {
            return EnvironmentVerdict(isSecure = false, reason = EnvironmentRejectReason.NON_ZERO_USER_ID)
        }
        val pathUser = SecureDataPath.userIdFromPath(filesDirPath)
        if (pathUser != null && pathUser != 0) {
            return EnvironmentVerdict(isSecure = false, reason = EnvironmentRejectReason.NON_ZERO_USER_ID)
        }
        return EnvironmentVerdict(isSecure = true)
    }
}
