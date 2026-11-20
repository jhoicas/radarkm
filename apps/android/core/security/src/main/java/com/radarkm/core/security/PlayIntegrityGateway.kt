package com.radarkm.core.security

/**
 * Gateway Play Integrity. El veredicto autoritativo lo decide el backend (T-03 API).
 * El cliente solo solicita tokens Standard/Classic y adjunta señales Device Recall / Key Attestation.
 */
enum class IntegrityRequestMode {
    STANDARD,
    CLASSIC,
}

data class IntegrityTokenRequest(
    val mode: IntegrityRequestMode,
    /** Hash de la petición (API Standard). */
    val requestHash: String? = null,
    /** Nonce (API Classic). */
    val nonce: String? = null,
    val includeDeviceRecall: Boolean = true,
    val includeKeyAttestationChallenge: ByteArray? = null,
) {
    init {
        when (mode) {
            IntegrityRequestMode.STANDARD ->
                require(!requestHash.isNullOrBlank()) { "Standard mode requires requestHash" }
            IntegrityRequestMode.CLASSIC ->
                require(!nonce.isNullOrBlank()) { "Classic mode requires nonce" }
        }
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is IntegrityTokenRequest) return false
        return mode == other.mode &&
            requestHash == other.requestHash &&
            nonce == other.nonce &&
            includeDeviceRecall == other.includeDeviceRecall &&
            includeKeyAttestationChallenge.contentEquals(other.includeKeyAttestationChallenge)
    }

    override fun hashCode(): Int {
        var result = mode.hashCode()
        result = 31 * result + (requestHash?.hashCode() ?: 0)
        result = 31 * result + (nonce?.hashCode() ?: 0)
        result = 31 * result + includeDeviceRecall.hashCode()
        result = 31 * result + (includeKeyAttestationChallenge?.contentHashCode() ?: 0)
        return result
    }
}

data class IntegrityTokenBundle(
    val integrityToken: String,
    val mode: IntegrityRequestMode,
    val requestHash: String? = null,
    val nonce: String? = null,
    val deviceRecallRequested: Boolean,
    val keyAttestationChallenge: ByteArray? = null,
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is IntegrityTokenBundle) return false
        return integrityToken == other.integrityToken &&
            mode == other.mode &&
            requestHash == other.requestHash &&
            nonce == other.nonce &&
            deviceRecallRequested == other.deviceRecallRequested &&
            keyAttestationChallenge.contentEquals(other.keyAttestationChallenge)
    }

    override fun hashCode(): Int {
        var result = integrityToken.hashCode()
        result = 31 * result + mode.hashCode()
        result = 31 * result + (requestHash?.hashCode() ?: 0)
        result = 31 * result + (nonce?.hashCode() ?: 0)
        result = 31 * result + deviceRecallRequested.hashCode()
        result = 31 * result + (keyAttestationChallenge?.contentHashCode() ?: 0)
        return result
    }
}

interface PlayIntegrityGateway {
    suspend fun requestIntegrityToken(request: IntegrityTokenRequest): Result<IntegrityTokenBundle>

    /** Compatibilidad con llamadas Classic simples. */
    suspend fun requestIntegrityToken(nonce: String): Result<String> =
        requestIntegrityToken(
            IntegrityTokenRequest(mode = IntegrityRequestMode.CLASSIC, nonce = nonce),
        ).map { it.integrityToken }
}

/**
 * Stub determinista para tests JVM / scaffolding sin Play Services.
 * No inventa veredictos de seguridad: solo empaqueta el token solicitado.
 */
class PlayIntegrityGatewayStub(
    private val tokenFactory: (IntegrityTokenRequest) -> String = { req ->
        "stub-token-${req.mode}-${req.requestHash ?: req.nonce}"
    },
) : PlayIntegrityGateway {
    override suspend fun requestIntegrityToken(request: IntegrityTokenRequest): Result<IntegrityTokenBundle> =
        Result.success(
            IntegrityTokenBundle(
                integrityToken = tokenFactory(request),
                mode = request.mode,
                requestHash = request.requestHash,
                nonce = request.nonce,
                deviceRecallRequested = request.includeDeviceRecall,
                keyAttestationChallenge = request.includeKeyAttestationChallenge,
            ),
        )
}
