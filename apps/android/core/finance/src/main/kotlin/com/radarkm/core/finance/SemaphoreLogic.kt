package com.radarkm.core.finance

/**
 * Semáforo de rentabilidad:
 * - GREEN: U_km y U_h cumplen umbrales (AND) y recogida dentro del máximo.
 * - YELLOW: cumple exactamente uno de los umbrales U_km / U_h.
 * - RED: falla ambos, métricas inválidas (d=0 o t=0), o supera distancia máxima de recogida.
 */
object SemaphoreLogic {
    fun color(
        netPerKmCop: Long,
        netPerHourCop: Long,
        pickupDistanceMeters: Long,
        thresholds: ProfitThresholds,
        distanceOrTimeInvalid: Boolean,
    ): SemaphoreColor {
        if (distanceOrTimeInvalid || pickupDistanceMeters > thresholds.maxPickupDistanceMeters) {
            return SemaphoreColor.RED
        }
        val kmOk = netPerKmCop >= thresholds.minNetPerKmCop
        val hourOk = netPerHourCop >= thresholds.minNetPerHourCop
        return when {
            kmOk && hourOk -> SemaphoreColor.GREEN
            kmOk || hourOk -> SemaphoreColor.YELLOW
            else -> SemaphoreColor.RED
        }
    }
}
