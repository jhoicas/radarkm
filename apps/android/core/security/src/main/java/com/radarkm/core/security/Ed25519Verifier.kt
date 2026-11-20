package com.radarkm.core.security

/**
 * Verificación de firmas Ed25519 de respuestas autoritativas del backend.
 * Implementación completa en tareas posteriores; aquí solo el contrato.
 */
interface Ed25519Verifier {
    fun verify(message: ByteArray, signature: ByteArray, publicKey: ByteArray): Boolean
}

class Ed25519VerifierStub : Ed25519Verifier {
    override fun verify(message: ByteArray, signature: ByteArray, publicKey: ByteArray): Boolean {
        require(publicKey.size == 32) { "Ed25519 public key must be 32 bytes" }
        require(signature.size == 64) { "Ed25519 signature must be 64 bytes" }
        return false
    }
}
