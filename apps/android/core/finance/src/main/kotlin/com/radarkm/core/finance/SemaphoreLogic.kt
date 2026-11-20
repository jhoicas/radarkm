package com.radarkm.core.finance

/**
 * Semáforo de rentabilidad: todos los umbrales deben cumplirse (AND).
 */
data class ProfitThresholds(
    val minNetPerKm: Long,
    val minNetPerHour: Long,
    val minNetAbsolute: MoneyCop,
)

object SemaphoreLogic {
    fun isGreen(
        netPerKm: Long,
        netPerHour: Long,
        netAbsolute: MoneyCop,
        thresholds: ProfitThresholds,
    ): Boolean =
        netPerKm >= thresholds.minNetPerKm &&
            netPerHour >= thresholds.minNetPerHour &&
            netAbsolute.centavosOrPesos >= thresholds.minNetAbsolute.centavosOrPesos

    /**
     * Contraoferta inDrive esqueleto: redondeo hacia arriba al múltiplo de 100 COP.
     */
    fun suggestCounterOffer(targetNet: MoneyCop, costs: MoneyCop): MoneyCop {
        val raw = targetNet.centavosOrPesos + costs.centavosOrPesos
        val rounded = ((raw + 99L) / 100L) * 100L
        return MoneyCop(rounded)
    }
}
