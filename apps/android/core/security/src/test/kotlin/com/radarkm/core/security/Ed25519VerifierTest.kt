package com.radarkm.core.security

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import java.security.KeyPairGenerator
import java.security.Signature
import java.time.Instant
import java.util.Base64

/**
 * TDD — verificación Ed25519 de envelopes backend (issued_at / expires_at / key_id).
 */
class Ed25519VerifierTest {

    private val fixedNow = Instant.parse("2026-06-15T12:00:00Z")
    private val clock = SecurityClock { fixedNow }

    @Test
    fun `valid envelope passes verify`() {
        val keys = generateEd25519()
        val payload = canonicalJson(
            issuedAt = fixedNow.minusSeconds(30),
            expiresAt = fixedNow.plusSeconds(300),
            keyId = "test-key",
            bodyJson = """{"remaining":1}""",
        )
        val envelope = signEnvelope(keys, payload)
        val verifier = DefaultEd25519Verifier(keys.rawPublic, clock)

        val outcome = verifier.verifyEnvelope(
            payloadB64 = envelope.payloadB64,
            signatureB64 = envelope.signatureB64,
            expectedKeyId = "test-key",
        )
        assertEquals(VerifyOutcome.Ok, outcome)
        assertTrue(verifier.verifyRaw(payload.toByteArray(Charsets.UTF_8), envelope.signatureBytes))
    }

    @Test
    fun `tampered payload fails verify`() {
        val keys = generateEd25519()
        val payload = canonicalJson(
            issuedAt = fixedNow.minusSeconds(30),
            expiresAt = fixedNow.plusSeconds(300),
            keyId = "test-key",
            bodyJson = """{"remaining":1}""",
        )
        val envelope = signEnvelope(keys, payload)
        val tampered = payload.replace("\"remaining\":1", "\"remaining\":99")
        val tamperedB64 = Base64.getEncoder().encodeToString(tampered.toByteArray(Charsets.UTF_8))
        val verifier = DefaultEd25519Verifier(keys.rawPublic, clock)

        val outcome = verifier.verifyEnvelope(tamperedB64, envelope.signatureB64, "test-key")
        assertTrue(outcome is VerifyOutcome.Rejected)
        assertEquals("invalid_signature", (outcome as VerifyOutcome.Rejected).reason)
    }

    @Test
    fun `one byte altered signature fails verify`() {
        val keys = generateEd25519()
        val payload = canonicalJson(
            issuedAt = fixedNow.minusSeconds(10),
            expiresAt = fixedNow.plusSeconds(60),
            keyId = "k1",
            bodyJson = """{"ok":true}""",
        )
        val envelope = signEnvelope(keys, payload)
        val badSig = envelope.signatureBytes.copyOf().also { it[0] = (it[0].toInt() xor 0x01).toByte() }
        val verifier = DefaultEd25519Verifier(keys.rawPublic, clock)

        assertFalse(verifier.verifyRaw(payload.toByteArray(Charsets.UTF_8), badSig))
        val outcome = verifier.verifyEnvelope(
            envelope.payloadB64,
            Base64.getEncoder().encodeToString(badSig),
            "k1",
        )
        assertTrue(outcome is VerifyOutcome.Rejected)
    }

    @Test
    fun `expired envelope is rejected`() {
        val keys = generateEd25519()
        val payload = canonicalJson(
            issuedAt = fixedNow.minusSeconds(600),
            expiresAt = fixedNow.minusSeconds(1),
            keyId = "test-key",
            bodyJson = """{"remaining":0}""",
        )
        val envelope = signEnvelope(keys, payload)
        val verifier = DefaultEd25519Verifier(keys.rawPublic, clock)

        val outcome = verifier.verifyEnvelope(envelope.payloadB64, envelope.signatureB64, "test-key")
        assertEquals(VerifyOutcome.Rejected("expired"), outcome)
    }

    private data class EdKeys(
        val rawPublic: ByteArray,
        val privateKey: java.security.PrivateKey,
        val publicKey: java.security.PublicKey,
    )

    private data class Signed(
        val payloadB64: String,
        val signatureB64: String,
        val signatureBytes: ByteArray,
    )

    private fun generateEd25519(): EdKeys {
        val kpg = KeyPairGenerator.getInstance("Ed25519")
        val pair = kpg.generateKeyPair()
        val encoded = pair.public.encoded
        val raw = encoded.copyOfRange(encoded.size - 32, encoded.size)
        return EdKeys(raw, pair.private, pair.public)
    }

    private fun signEnvelope(keys: EdKeys, payload: String): Signed {
        val bytes = payload.toByteArray(Charsets.UTF_8)
        val sig = Signature.getInstance("Ed25519")
        sig.initSign(keys.privateKey)
        sig.update(bytes)
        val signature = sig.sign()
        return Signed(
            payloadB64 = Base64.getEncoder().encodeToString(bytes),
            signatureB64 = Base64.getEncoder().encodeToString(signature),
            signatureBytes = signature,
        )
    }

    private fun canonicalJson(
        issuedAt: Instant,
        expiresAt: Instant,
        keyId: String,
        bodyJson: String,
    ): String =
        """{"issued_at":"$issuedAt","expires_at":"$expiresAt","key_id":"$keyId","body":$bodyJson}"""
}
