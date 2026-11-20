package com.radarkm.core.finance

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class ProfitabilityEngineTest {

    private val engine = ProfitabilityEngine()

    @Test
    fun netProfit_subtractsCostsInLongCop() {
        val fare = MoneyCop(25_000L)
        val costs = MoneyCop(8_000L)
        val net = engine.netProfit(fare, costs)
        assertEquals(17_000L, net.centavosOrPesos)
    }

    @Test
    fun semaphore_requiresAllThresholdsAnd() {
        val thresholds = ProfitThresholds(
            minNetPerKm = 1_000L,
            minNetPerHour = 15_000L,
            minNetAbsolute = MoneyCop(10_000L),
        )
        assertTrue(
            SemaphoreLogic.isGreen(
                netPerKm = 1_200L,
                netPerHour = 18_000L,
                netAbsolute = MoneyCop(12_000L),
                thresholds = thresholds,
            ),
        )
    }
}
