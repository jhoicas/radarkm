package com.radarkm.feature.billing

import android.app.Activity
import android.content.Context
import com.android.billingclient.api.BillingClient
import com.android.billingclient.api.BillingClientStateListener
import com.android.billingclient.api.BillingResult
import com.android.billingclient.api.PendingPurchasesParams
import com.android.billingclient.api.PurchasesUpdatedListener

/**
 * Integración base con Play Billing Library 8+.
 * Prohibido enlazar pasarelas externas en el APK.
 */
class PlayBillingGateway(
    context: Context,
    purchasesUpdatedListener: PurchasesUpdatedListener,
) {
    val client: BillingClient = BillingClient.newBuilder(context)
        .setListener(purchasesUpdatedListener)
        .enablePendingPurchases(
            PendingPurchasesParams.newBuilder()
                .enableOneTimeProducts()
                .build(),
        )
        .build()

    fun startConnection(onReady: (BillingResult) -> Unit) {
        client.startConnection(
            object : BillingClientStateListener {
                override fun onBillingSetupFinished(billingResult: BillingResult) {
                    onReady(billingResult)
                }

                override fun onBillingServiceDisconnected() {
                    // Scaffolding: reconnect policy in later tasks.
                }
            },
        )
    }

    @Suppress("UNUSED_PARAMETER")
    fun launchPurchaseFlowPlaceholder(activity: Activity) {
        // ProductDetails query + launchBillingFlow in later tasks (Billing 8 API).
    }
}
