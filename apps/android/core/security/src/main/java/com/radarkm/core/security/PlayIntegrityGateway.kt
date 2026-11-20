package com.radarkm.core.security

/**
 * Gateway esqueleto para Play Integrity. La autoridad de veredicto es el backend.
 */
interface PlayIntegrityGateway {
    suspend fun requestIntegrityToken(nonce: String): Result<String>
}

class PlayIntegrityGatewayStub : PlayIntegrityGateway {
    override suspend fun requestIntegrityToken(nonce: String): Result<String> =
        Result.failure(UnsupportedOperationException("Play Integrity wiring is deferred past scaffolding"))
}
