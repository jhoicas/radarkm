package com.radarkm.core.finance

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

/**
 * TDD — Motor financiero (CA-2 / CA-3).
 * Todos los montos en Long COP; distancias en metros; tiempos en segundos.
 */
class ProfitabilityEngineTest {

    private val engine = ProfitabilityEngine()

    private val ownedCar = CostProfile(
        vehicleProfile = VehicleProfile.OWNED_CAR,
        fuelCopPerKm = 450L,
        maintenanceCopPerKm = 80L,
        depreciationOrRentCopPerHour = 3_000L,
        platformCommissionBps = 2_500L, // 25%
    )

    private val rentedCar = CostProfile(
        vehicleProfile = VehicleProfile.RENTED_CAR,
        fuelCopPerKm = 450L,
        maintenanceCopPerKm = 0L,
        depreciationOrRentCopPerHour = 12_000L, // cuota alquiler prorrateada
        platformCommissionBps = 2_500L,
    )

    private val motorcycle = CostProfile(
        vehicleProfile = VehicleProfile.MOTORCYCLE,
        fuelCopPerKm = 180L,
        maintenanceCopPerKm = 40L,
        depreciationOrRentCopPerHour = 1_500L,
        platformCommissionBps = 2_000L, // 20%
    )

    private val thresholds = ProfitThresholds(
        minNetPerKmCop = 1_000L,
        minNetPerHourCop = 15_000L,
        maxPickupDistanceMeters = 5_000L,
    )

    @Nested
    inner class Totals {
        @Test
        fun sumsPickupAndTripDistanceAndAllTimes() {
            val offer = TripOffer(
                fareCop = MoneyCop(40_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 1_200L,
                waitDurationSeconds = 180L,
                tollsCop = MoneyCop(0L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(10_000L, result.totalDistanceMeters) // d_rec + d_viaje
            assertEquals(1_980L, result.totalDurationSeconds) // t_rec + t_viaje + t_esp
        }
    }

    @Nested
    inner class NetUtilities {
        @Test
        fun grossFare_discountsFuelMaintenanceDepreciationTollsAndCommission() {
            // d_total=10km, t_total=1h (3600s)
            // fuel=450*10=4500, maint=80*10=800, depr=3000*1=3000, tolls=2000
            // commission=25% of 40000=10000
            // costs=4500+800+3000+2000+10000=20300
            // net=40000-20300=19700
            // U_km=19700*1000/10000=1970
            // U_h=19700*3600/3600=19700
            val offer = TripOffer(
                fareCop = MoneyCop(40_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(2_000L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(19_700L, result.netUtility.centavosOrPesos)
            assertEquals(1_970L, result.netPerKmCop)
            assertEquals(19_700L, result.netPerHourCop)
            assertEquals(10_000L, result.commission.centavosOrPesos)
        }

        @Test
        fun netFare_doesNotApplyPlatformCommissionAgain() {
            val offer = TripOffer(
                fareCop = MoneyCop(30_000L),
                fareKind = FareKind.NET,
                pickupDistanceMeters = 1_000L,
                tripDistanceMeters = 9_000L,
                pickupDurationSeconds = 300L,
                tripDurationSeconds = 3_000L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(0L),
            )
            // fuel=4500, maint=800, depr=3000 → costs=8300; commission=0; net=21700
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(0L, result.commission.centavosOrPesos)
            assertEquals(21_700L, result.netUtility.centavosOrPesos)
            assertEquals(2_170L, result.netPerKmCop)
            assertEquals(21_700L, result.netPerHourCop)
        }

        @Test
        fun rentedCar_usesHourlyRentInsteadOfOwnershipDepreciation() {
            val offer = TripOffer(
                fareCop = MoneyCop(50_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 1_000L,
                tripDistanceMeters = 9_000L,
                pickupDurationSeconds = 300L,
                tripDurationSeconds = 3_000L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(0L),
            )
            // fuel=4500, maint=0, rent=12000, commission=12500 → costs=29000; net=21000
            val result = engine.evaluate(offer, rentedCar, thresholds)
            assertEquals(VehicleProfile.RENTED_CAR, result.vehicleProfile)
            assertEquals(21_000L, result.netUtility.centavosOrPesos)
            assertEquals(2_100L, result.netPerKmCop)
            assertEquals(21_000L, result.netPerHourCop)
        }

        @Test
        fun motorcycle_usesLowerFuelAndCommissionProfile() {
            val offer = TripOffer(
                fareCop = MoneyCop(25_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 500L,
                tripDistanceMeters = 4_500L,
                pickupDurationSeconds = 180L,
                tripDurationSeconds = 1_440L,
                waitDurationSeconds = 180L,
                tollsCop = MoneyCop(0L),
            )
            // d=5km, t=1800s=0.5h
            // fuel=180*5=900, maint=40*5=200, depr=1500/2=750, commission=5000
            // costs=6850; net=18150
            // U_km=18150*1000/5000=3630
            // U_h=18150*3600/1800=36300
            val result = engine.evaluate(offer, motorcycle, thresholds)
            assertEquals(18_150L, result.netUtility.centavosOrPesos)
            assertEquals(3_630L, result.netPerKmCop)
            assertEquals(36_300L, result.netPerHourCop)
        }
    }

    @Nested
    inner class Semaphore {
        @Test
        fun green_onlyWhenBothThresholdsPassAndPickupWithinMax() {
            val offer = TripOffer(
                fareCop = MoneyCop(40_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(2_000L),
            )
            // U_km=1970>=1000, U_h=19700>=15000, pickup=2000<=5000
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(SemaphoreColor.GREEN, result.semaphore)
        }

        @Test
        fun yellow_whenGoodPerKmButLowPerHourDueToCongestion() {
            // Same net economics stretched over many hours → high U_km, low U_h
            val offer = TripOffer(
                fareCop = MoneyCop(40_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 1_800L,
                tripDurationSeconds = 10_800L, // trancón
                waitDurationSeconds = 1_800L,
                tollsCop = MoneyCop(2_000L),
            )
            // t=14400s=4h; depr=3000*4=12000; costs=4500+800+12000+2000+10000=29300
            // net=10700; U_km=1070; U_h=10700*3600/14400=2675 < 15000
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(1_070L, result.netPerKmCop)
            assertEquals(2_675L, result.netPerHourCop)
            assertEquals(SemaphoreColor.YELLOW, result.semaphore)
        }

        @Test
        fun red_whenBothThresholdsFail() {
            val offer = TripOffer(
                fareCop = MoneyCop(12_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(2_000L),
            )
            // commission=3000; costs=4500+800+3000+2000+3000=13300; net=-1300
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertTrue(result.netUtility.centavosOrPesos < 0L)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }

        @Test
        fun red_whenPickupDistanceExceedsMaxEvenIfUtilitiesAreGood() {
            val offer = TripOffer(
                fareCop = MoneyCop(40_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 6_000L, // > 5000
                tripDistanceMeters = 4_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(2_000L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }
    }

    @Nested
    inner class CounterOffer {
        @Test
        fun roundsUpToNearest500Cop() {
            // desired net 15000 + operating costs (fuel+maint+depr+tolls) on 10km/1h owned = 4500+800+3000+0=8300
            // raw=23300 → ceil 500 → 23500; then add commission for GROSS:
            // P * (1 - 0.25) = 23300 → P = 23300 * 10000 / 7500 = 31066.66… → 31067 → ceil500 → 31500
            val offer = TripOffer(
                fareCop = MoneyCop(20_000L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(0L),
            )
            val contra = engine.suggestCounterOffer(
                offer = offer,
                costs = ownedCar,
                targetNet = MoneyCop(15_000L),
            )
            assertEquals(0L, contra.centavosOrPesos % 500L)
            assertEquals(31_500L, contra.centavosOrPesos)
        }

        @Test
        fun netFare_counterOfferIsTargetPlusOperatingCostsCeiledTo500() {
            val offer = TripOffer(
                fareCop = MoneyCop(10_000L),
                fareKind = FareKind.NET,
                pickupDistanceMeters = 2_000L,
                tripDistanceMeters = 8_000L,
                pickupDurationSeconds = 600L,
                tripDurationSeconds = 2_700L,
                waitDurationSeconds = 300L,
                tollsCop = MoneyCop(0L),
            )
            // operating=8300 + target=15000 = 23300 → 23500
            val contra = engine.suggestCounterOffer(
                offer = offer,
                costs = ownedCar,
                targetNet = MoneyCop(15_000L),
            )
            assertEquals(23_500L, contra.centavosOrPesos)
        }
    }

    @Nested
    inner class EdgeCases {
        @Test
        fun zeroTotalDistance_marksRedAndZeroPerKm() {
            val offer = TripOffer(
                fareCop = MoneyCop(20_000L),
                fareKind = FareKind.NET,
                pickupDistanceMeters = 0L,
                tripDistanceMeters = 0L,
                pickupDurationSeconds = 60L,
                tripDurationSeconds = 60L,
                waitDurationSeconds = 0L,
                tollsCop = MoneyCop(0L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(0L, result.totalDistanceMeters)
            assertEquals(0L, result.netPerKmCop)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }

        @Test
        fun zeroTotalTime_marksRedAndZeroPerHour() {
            val offer = TripOffer(
                fareCop = MoneyCop(20_000L),
                fareKind = FareKind.NET,
                pickupDistanceMeters = 1_000L,
                tripDistanceMeters = 1_000L,
                pickupDurationSeconds = 0L,
                tripDurationSeconds = 0L,
                waitDurationSeconds = 0L,
                tollsCop = MoneyCop(0L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertEquals(0L, result.totalDurationSeconds)
            assertEquals(0L, result.netPerHourCop)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }

        @Test
        fun zeroFare_yieldsNonPositiveNet() {
            val offer = TripOffer(
                fareCop = MoneyCop(0L),
                fareKind = FareKind.GROSS,
                pickupDistanceMeters = 1_000L,
                tripDistanceMeters = 4_000L,
                pickupDurationSeconds = 300L,
                tripDurationSeconds = 900L,
                waitDurationSeconds = 0L,
                tollsCop = MoneyCop(0L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertTrue(result.netUtility.centavosOrPesos <= 0L)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }

        @Test
        fun commissionAtOrAbove100Percent_isRejected() {
            val bad = ownedCar.copy(platformCommissionBps = 10_000L)
            assertThrows<IllegalArgumentException> {
                engine.evaluate(
                    TripOffer(
                        fareCop = MoneyCop(10_000L),
                        fareKind = FareKind.GROSS,
                        pickupDistanceMeters = 1_000L,
                        tripDistanceMeters = 1_000L,
                        pickupDurationSeconds = 60L,
                        tripDurationSeconds = 60L,
                        waitDurationSeconds = 0L,
                        tollsCop = MoneyCop(0L),
                    ),
                    bad,
                    thresholds,
                )
            }
        }

        @Test
        fun tollsGreaterThanFare_canProduceNegativeNet() {
            val offer = TripOffer(
                fareCop = MoneyCop(10_000L),
                fareKind = FareKind.NET,
                pickupDistanceMeters = 500L,
                tripDistanceMeters = 500L,
                pickupDurationSeconds = 120L,
                tripDurationSeconds = 120L,
                waitDurationSeconds = 0L,
                tollsCop = MoneyCop(15_000L),
            )
            val result = engine.evaluate(offer, ownedCar, thresholds)
            assertTrue(result.netUtility.centavosOrPesos < 0L)
            assertEquals(SemaphoreColor.RED, result.semaphore)
        }
    }
}
