package com.radarkm.feature.overlay

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

/**
 * Pantalla de Prominent Disclosure (requisito Play para Accessibility).
 */
@Composable
fun ProminentDisclosureScreen(onContinue: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Text(text = "RadarKm — Divulgación prominente")
        Text(
            text = "RadarKm usa el servicio de accesibilidad únicamente para leer " +
                "ofertas de viaje en tu dispositivo y calcular rentabilidad. " +
                "Las capturas se procesan en memoria RAM y no se suben a servidores.",
        )
        Button(onClick = onContinue) {
            Text("Entiendo y deseo continuar")
        }
    }
}
