package com.radarkm.feature.billing

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.radarkm.core.finance.MoneyCop

/**
 * Pantalla esqueleto de configuración de vehículo / costos (enteros COP).
 */
@Composable
fun VehicleConfigScreen(
    fuelCostPerKm: MoneyCop = MoneyCop(450L),
    platformCommissionPerTrip: MoneyCop = MoneyCop(3_000L),
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("Configuración de rentabilidad")
        Text("Costo combustible/km (COP): ${fuelCostPerKm.centavosOrPesos}")
        Text("Comisión por servicio (COP): ${platformCommissionPerTrip.centavosOrPesos}")
        Text("Los montos se editan solo como enteros COP.")
    }
}
