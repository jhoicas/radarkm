package com.radarkm.core.finance

enum class VehicleProfile {
    OWNED_CAR,
    RENTED_CAR,
    MOTORCYCLE,
}

/** GROSS: la oferta incluye comisión de plataforma. NET: ya viene neta. */
enum class FareKind {
    GROSS,
    NET,
}

enum class SemaphoreColor {
    GREEN,
    YELLOW,
    RED,
}

data class TripOffer(
    val fareCop: MoneyCop,
    val fareKind: FareKind,
    val pickupDistanceMeters: Long,
    val tripDistanceMeters: Long,
    val pickupDurationSeconds: Long,
    val tripDurationSeconds: Long,
    val waitDurationSeconds: Long,
    val tollsCop: MoneyCop,
) {
    init {
        require(pickupDistanceMeters >= 0L)
        require(tripDistanceMeters >= 0L)
        require(pickupDurationSeconds >= 0L)
        require(tripDurationSeconds >= 0L)
        require(waitDurationSeconds >= 0L)
    }

    val totalDistanceMeters: Long get() = pickupDistanceMeters + tripDistanceMeters
    val totalDurationSeconds: Long
        get() = pickupDurationSeconds + tripDurationSeconds + waitDurationSeconds
}

/**
 * Costos configurables del conductor.
 * [platformCommissionBps]: basis points (2500 = 25%). Debe ser < 10000.
 */
data class CostProfile(
    val vehicleProfile: VehicleProfile,
    val fuelCopPerKm: Long,
    val maintenanceCopPerKm: Long,
    val depreciationOrRentCopPerHour: Long,
    val platformCommissionBps: Long,
) {
    init {
        require(fuelCopPerKm >= 0L)
        require(maintenanceCopPerKm >= 0L)
        require(depreciationOrRentCopPerHour >= 0L)
        require(platformCommissionBps >= 0L)
    }
}

data class ProfitThresholds(
    val minNetPerKmCop: Long,
    val minNetPerHourCop: Long,
    val maxPickupDistanceMeters: Long,
) {
    init {
        require(maxPickupDistanceMeters >= 0L)
    }
}

data class ProfitEvaluation(
    val totalDistanceMeters: Long,
    val totalDurationSeconds: Long,
    val commission: MoneyCop,
    val netUtility: MoneyCop,
    val netPerKmCop: Long,
    val netPerHourCop: Long,
    val semaphore: SemaphoreColor,
    val vehicleProfile: VehicleProfile,
)
