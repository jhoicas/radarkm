package com.radarkm.core.security

import java.security.KeyFactory
import java.security.PublicKey
import java.security.Signature
import java.security.spec.X509EncodedKeySpec
import java.time.Instant
import java.util.Base64

fun interface SecurityClock {
    fun now(): Instant
}

sealed class VerifyOutcome {
    data object Ok : VerifyOutcome()
    data class Rejected(val reason: String) : VerifyOutcome()
}

/**
 * Verificación de firmas Ed25519 de respuestas autoritativas del backend.
 * Nunca autoriza cuota/premium por sí solo: solo valida integridad y frescura.
 */
interface Ed25519Verifier {
    fun verifyRaw(message: ByteArray, signature: ByteArray): Boolean
    fun verifyEnvelope(
        payloadB64: String,
        signatureB64: String,
        expectedKeyId: String? = null,
    ): VerifyOutcome
}

class DefaultEd25519Verifier(
    private val publicKeyRaw32: ByteArray,
    private val clock: SecurityClock = SecurityClock { Instant.now() },
) : Ed25519Verifier {

    init {
        require(publicKeyRaw32.size == 32) { "Ed25519 public key must be 32 bytes" }
    }

    private val publicKey: PublicKey by lazy { decodeEd25519PublicKey(publicKeyRaw32) }

    override fun verifyRaw(message: ByteArray, signature: ByteArray): Boolean {
        if (signature.size != 64) return false
        return runCatching {
            val sig = Signature.getInstance("Ed25519")
            sig.initVerify(publicKey)
            sig.update(message)
            sig.verify(signature)
        }.getOrDefault(false)
    }

    override fun verifyEnvelope(
        payloadB64: String,
        signatureB64: String,
        expectedKeyId: String?,
    ): VerifyOutcome {
        val payload = runCatching { Base64.getDecoder().decode(payloadB64) }
            .getOrElse { return VerifyOutcome.Rejected("invalid_payload_b64") }
        val signature = runCatching { Base64.getDecoder().decode(signatureB64) }
            .getOrElse { return VerifyOutcome.Rejected("invalid_signature_b64") }

        if (!verifyRaw(payload, signature)) {
            return VerifyOutcome.Rejected("invalid_signature")
        }

        val json = payload.toString(Charsets.UTF_8)
        val keyId = extractJsonString(json, "key_id")
        if (expectedKeyId != null && keyId != null && keyId != expectedKeyId) {
            return VerifyOutcome.Rejected("key_id_mismatch")
        }

        val expiresAtRaw = extractJsonString(json, "expires_at")
            ?: return VerifyOutcome.Rejected("missing_expires_at")
        val expiresAt = runCatching { Instant.parse(expiresAtRaw) }
            .getOrElse { return VerifyOutcome.Rejected("invalid_expires_at") }

        if (!expiresAt.isAfter(clock.now())) {
            return VerifyOutcome.Rejected("expired")
        }
        return VerifyOutcome.Ok
    }

    companion object {
        /** Prefijo X.509 SubjectPublicKeyInfo para clave pública Ed25519 cruda (32 bytes). */
        private val ED25519_SPKI_PREFIX = byteArrayOf(
            0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00,
        )

        fun decodeEd25519PublicKey(raw32: ByteArray): PublicKey {
            require(raw32.size == 32)
            val encoded = ByteArray(ED25519_SPKI_PREFIX.size + raw32.size)
            System.arraycopy(ED25519_SPKI_PREFIX, 0, encoded, 0, ED25519_SPKI_PREFIX.size)
            System.arraycopy(raw32, 0, encoded, ED25519_SPKI_PREFIX.size, raw32.size)
            return KeyFactory.getInstance("Ed25519")
                .generatePublic(X509EncodedKeySpec(encoded))
        }

        internal fun extractJsonString(json: String, key: String): String? {
            val pattern = Regex(""""$key"\s*:\s*"([^"]*)"""")
            return pattern.find(json)?.groupValues?.get(1)
        }
    }
}
