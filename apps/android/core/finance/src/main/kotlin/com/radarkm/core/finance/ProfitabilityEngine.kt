package com.radarkm.core.finance

/**
 * Esqueleto del motor financiero. Cálculos solo con Long/MoneyCop.
 */
class ProfitabilityEngine {

    /**
     * Neto = tarifa - costos configurables (todos en COP enteros).
     */
    fun netProfit(fare: MoneyCop, costs: MoneyCop): MoneyCop = fare - costs

    /**
     * COP netos por kilómetro (enteros; división truncada hacia abajo).
     */
    fun netPerKm(net: MoneyCop, distanceMeters: Long): Long {
        require(distanceMeters > 0L) { "distanceMeters must be > 0" }
        val kmTimes1000 = distanceMeters // meters as proxy scale for scaffolding
        return (net.centavosOrPesos * 1000L) / kmTimes1000
    }

    /**
     * COP netos por hora estimada (enteros).
     */
    fun netPerHour(net: MoneyCop, durationMinutes: Long): Long {
        require(durationMinutes > 0L) { "durationMinutes must be > 0" }
        return (net.centavosOrPesos * 60L) / durationMinutes
    }
}
