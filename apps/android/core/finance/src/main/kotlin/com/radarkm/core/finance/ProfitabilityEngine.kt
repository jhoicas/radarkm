package com.radarkm.core.finance

/**
 * Motor financiero en enteros COP (CA-2 / CA-3).
 *
 * Invariantes:
 * - Dinero solo en [Long] / [MoneyCop] (sin Float/Double).
 * - Distancia total: d_rec + d_viaje (metros).
 * - Tiempo total: t_rec + t_viaje + t_esp (segundos).
 */
class ProfitabilityEngine {

    fun evaluate(
        offer: TripOffer,
        costs: CostProfile,
        thresholds: ProfitThresholds,
    ): ProfitEvaluation {
        requireCommissionBelow100Percent(costs)

        val totalDistanceMeters = offer.totalDistanceMeters
        val totalDurationSeconds = offer.totalDurationSeconds

        val operatingCosts = operatingCostsCop(offer, costs)
        val commission = platformCommission(offer.fareCop, offer.fareKind, costs.platformCommissionBps)
        val net = offer.fareCop.centavosOrPesos - operatingCosts - commission.centavosOrPesos

        val distanceInvalid = totalDistanceMeters == 0L
        val timeInvalid = totalDurationSeconds == 0L

        val netPerKm = if (distanceInvalid) {
            0L
        } else {
            (net * 1_000L) / totalDistanceMeters
        }
        val netPerHour = if (timeInvalid) {
            0L
        } else {
            (net * 3_600L) / totalDurationSeconds
        }

        val semaphore = SemaphoreLogic.color(
            netPerKmCop = netPerKm,
            netPerHourCop = netPerHour,
            pickupDistanceMeters = offer.pickupDistanceMeters,
            thresholds = thresholds,
            distanceOrTimeInvalid = distanceInvalid || timeInvalid,
        )

        return ProfitEvaluation(
            totalDistanceMeters = totalDistanceMeters,
            totalDurationSeconds = totalDurationSeconds,
            commission = commission,
            netUtility = MoneyCop(net),
            netPerKmCop = netPerKm,
            netPerHourCop = netPerHour,
            semaphore = semaphore,
            vehicleProfile = costs.vehicleProfile,
        )
    }

    /**
     * Contraoferta inDrive: precio bruto/neto que deja [targetNet] tras costos,
     * redondeado **hacia arriba** al múltiplo de 500 COP.
     */
    fun suggestCounterOffer(
        offer: TripOffer,
        costs: CostProfile,
        targetNet: MoneyCop,
    ): MoneyCop {
        requireCommissionBelow100Percent(costs)
        require(targetNet.centavosOrPesos >= 0L) { "targetNet must be non-negative" }

        val operating = operatingCostsCop(offer, costs)
        val neededBeforeCommission = targetNet.centavosOrPesos + operating

        val rawPrice = when (offer.fareKind) {
            FareKind.NET -> neededBeforeCommission
            FareKind.GROSS -> {
                val keepBps = 10_000L - costs.platformCommissionBps
                ceilDiv(neededBeforeCommission * 10_000L, keepBps)
            }
        }
        return MoneyCop(ceilToMultiple(rawPrice, COUNTER_OFFER_STEP_COP))
    }

    private fun operatingCostsCop(offer: TripOffer, costs: CostProfile): Long {
        val meters = offer.totalDistanceMeters
        val seconds = offer.totalDurationSeconds
        val fuel = perKmCost(costs.fuelCopPerKm, meters)
        val maintenance = perKmCost(costs.maintenanceCopPerKm, meters)
        val depreciationOrRent = perHourCost(costs.depreciationOrRentCopPerHour, seconds)
        return fuel + maintenance + depreciationOrRent + offer.tollsCop.centavosOrPesos
    }

    private fun platformCommission(
        fare: MoneyCop,
        fareKind: FareKind,
        commissionBps: Long,
    ): MoneyCop {
        if (fareKind == FareKind.NET) return MoneyCop(0L)
        return MoneyCop((fare.centavosOrPesos * commissionBps) / 10_000L)
    }

    private fun perKmCost(copPerKm: Long, meters: Long): Long {
        if (meters == 0L || copPerKm == 0L) return 0L
        return (copPerKm * meters) / 1_000L
    }

    private fun perHourCost(copPerHour: Long, seconds: Long): Long {
        if (seconds == 0L || copPerHour == 0L) return 0L
        return (copPerHour * seconds) / 3_600L
    }

    private fun requireCommissionBelow100Percent(costs: CostProfile) {
        require(costs.platformCommissionBps < 10_000L) {
            "platformCommissionBps must be < 10000 (100%)"
        }
    }

    private fun ceilDiv(numerator: Long, denominator: Long): Long {
        require(denominator > 0L)
        return (numerator + denominator - 1L) / denominator
    }

    private fun ceilToMultiple(value: Long, multiple: Long): Long {
        require(multiple > 0L)
        if (value <= 0L) return 0L
        return ((value + multiple - 1L) / multiple) * multiple
    }

    companion object {
        const val COUNTER_OFFER_STEP_COP: Long = 500L
    }
}
