package com.radarkm.core.finance

/**
 * Monto monetario en pesos colombianos (COP) como entero.
 * Prohibido Float/Double para dinero (specs/00-constitution.md).
 */
@JvmInline
value class MoneyCop(val centavosOrPesos: Long) {
    init {
        require(centavosOrPesos >= 0L) { "MoneyCop must be non-negative" }
    }

    operator fun plus(other: MoneyCop): MoneyCop = MoneyCop(centavosOrPesos + other.centavosOrPesos)

    operator fun minus(other: MoneyCop): MoneyCop =
        MoneyCop((centavosOrPesos - other.centavosOrPesos).coerceAtLeast(0L))
}
