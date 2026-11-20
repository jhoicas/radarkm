package com.radarkm.core.security

/**
 * Reglas puras (JVM-testables) para rutas de datos legítimas vs clonadores / perfiles.
 * El NDK aplica la misma política en [NativeSecurity.verifyDataPath].
 */
object SecureDataPath {
    const val EXPECTED_PACKAGE: String = "com.radarkm.app"

    private val userPathRegex = Regex("""^/data/user/(\d+)/([^/]+)(/.*)?$""")
    private val dataDataRegex = Regex("""^/data/data/([^/]+)(/.*)?$""")
    private val clonerMarkers = listOf("/virtual/", "/parallel/", "/clone/", "/dual/", "/island/")

    fun isLegitimateDataPath(
        absolutePath: String,
        packageName: String = EXPECTED_PACKAGE,
    ): Boolean {
        val path = absolutePath.trim()
        if (path.isEmpty()) return false
        if (clonerMarkers.any { path.contains(it, ignoreCase = true) }) return false

        userPathRegex.matchEntire(path)?.let { match ->
            val userId = match.groupValues[1].toIntOrNull() ?: return false
            val pkg = match.groupValues[2]
            return userId == 0 && pkg == packageName
        }

        dataDataRegex.matchEntire(path)?.let { match ->
            return match.groupValues[1] == packageName
        }

        return false
    }

    fun userIdFromPath(absolutePath: String): Int? {
        val match = userPathRegex.matchEntire(absolutePath.trim()) ?: return null
        return match.groupValues[1].toIntOrNull()
    }
}
