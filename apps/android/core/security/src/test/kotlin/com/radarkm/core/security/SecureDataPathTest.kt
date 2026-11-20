package com.radarkm.core.security

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test

/**
 * TDD — rutas de datos legítimas vs clonadores (CA-5), sin cargar NDK.
 */
class SecureDataPathTest {

    private val pkg = SecureDataPath.EXPECTED_PACKAGE

    @Nested
    inner class AcceptsLegitimatePaths {
        @Test
        fun `data user 0 package root`() {
            assertTrue(SecureDataPath.isLegitimateDataPath("/data/user/0/$pkg", pkg))
        }

        @Test
        fun `data user 0 filesDir`() {
            assertTrue(SecureDataPath.isLegitimateDataPath("/data/user/0/$pkg/files", pkg))
        }

        @Test
        fun `legacy data data package`() {
            assertTrue(SecureDataPath.isLegitimateDataPath("/data/data/$pkg", pkg))
            assertTrue(SecureDataPath.isLegitimateDataPath("/data/data/$pkg/files", pkg))
        }
    }

    @Nested
    inner class RejectsClonersAndTampering {
        @Test
        fun `work profile or clone user 999`() {
            assertFalse(SecureDataPath.isLegitimateDataPath("/data/user/999/$pkg/files", pkg))
        }

        @Test
        fun `non-zero user id`() {
            assertFalse(SecureDataPath.isLegitimateDataPath("/data/user/10/$pkg/files", pkg))
            assertFalse(SecureDataPath.isLegitimateDataPath("/data/user/1/$pkg", pkg))
        }

        @Test
        fun `nested virtual cloner directories`() {
            assertFalse(
                SecureDataPath.isLegitimateDataPath(
                    "/data/user/0/com.lbe.parallel/virtual/data/user/0/$pkg/files",
                    pkg,
                ),
            )
            assertFalse(
                SecureDataPath.isLegitimateDataPath(
                    "/data/data/com.dual.space/parallel/$pkg/files",
                    pkg,
                ),
            )
        }

        @Test
        fun `wrong package under user 0`() {
            assertFalse(SecureDataPath.isLegitimateDataPath("/data/user/0/com.other.app/files", pkg))
        }

        @Test
        fun `empty or blank`() {
            assertFalse(SecureDataPath.isLegitimateDataPath("", pkg))
            assertFalse(SecureDataPath.isLegitimateDataPath("   ", pkg))
        }
    }

    @Nested
    inner class UserIdAndCloneDetector {
        @Test
        fun `extracts user id from path`() {
            assertEquals(0, SecureDataPath.userIdFromPath("/data/user/0/$pkg/files"))
            assertEquals(999, SecureDataPath.userIdFromPath("/data/user/999/$pkg/files"))
            assertEquals(null, SecureDataPath.userIdFromPath("/data/data/$pkg/files"))
        }

        @Test
        fun `clone detector rejects non-zero userId even if path looks ok`() {
            val detector = CloneDetector()
            val verdict = detector.evaluate(
                filesDirPath = "/data/user/0/$pkg/files",
                userId = 10,
            )
            assertFalse(verdict.isSecure)
            assertEquals(EnvironmentRejectReason.NON_ZERO_USER_ID, verdict.reason)
        }

        @Test
        fun `clone detector accepts primary user and legitimate path`() {
            val verdict = CloneDetector().evaluate(
                filesDirPath = "/data/user/0/$pkg/files",
                userId = 0,
            )
            assertTrue(verdict.isSecure)
            assertEquals(null, verdict.reason)
        }

        @Test
        fun `clone detector rejects cloner path`() {
            val verdict = CloneDetector().evaluate(
                filesDirPath = "/data/user/999/$pkg/files",
                userId = 999,
            )
            assertFalse(verdict.isSecure)
            assertEquals(EnvironmentRejectReason.CLONE_OR_INVALID_PATH, verdict.reason)
        }
    }
}
