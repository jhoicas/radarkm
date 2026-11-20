package com.radarkm.core.finance

/**
 * Monto monetario en pesos colombianos (COP) como entero firmado.
 * Prohibido Float/Double para dinero (specs/00-constitution.md).
 * Puede ser negativo para utilidad neta.
 */
@JvmInline
value class MoneyCop(val centavosOrPesos: Long) {
    val pesos: Long get() = centavosOrPesos

    operator fun plus(other: MoneyCop): MoneyCop = MoneyCop(centavosOrPesos + other.centavosOrPesos)

    operator fun minus(other: MoneyCop): MoneyCop = MoneyCop(centavosOrPesos - other.centavosOrPesos)

    operator fun unaryMinus(): MoneyCop = MoneyCop(-centavosOrPesos)
}
