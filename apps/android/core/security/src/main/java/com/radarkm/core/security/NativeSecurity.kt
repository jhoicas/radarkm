package com.radarkm.core.security

internal object NativeSecurity {
    init {
        System.loadLibrary("radarkm_security")
    }

    external fun verifyDataPath(absolutePath: String): Boolean

    external fun quickTamperProbe(): Boolean
}
