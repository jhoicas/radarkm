package com.radarkm.core.network

import io.ktor.client.HttpClient
import io.ktor.client.engine.okhttp.OkHttp
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.request.url
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient

/**
 * Cliente HTTP con motor **ktor-client-okhttp** para soportar [okhttp3.CertificatePinner].
 */
object HttpClientFactory {

    fun create(
        baseUrl: String,
        pinning: CertificatePinningConfig,
        enableLogging: Boolean = false,
    ): HttpClient {
        val okHttpClient = OkHttpClient.Builder()
            .certificatePinner(pinning.toCertificatePinner())
            .build()

        return HttpClient(OkHttp) {
            expectSuccess = false
            engine {
                preconfigured = okHttpClient
            }
            install(ContentNegotiation) {
                json(
                    Json {
                        ignoreUnknownKeys = true
                        isLenient = false
                        explicitNulls = false
                    },
                )
            }
            if (enableLogging) {
                install(Logging) {
                    level = LogLevel.INFO
                }
            }
            defaultRequest {
                url(baseUrl)
            }
        }
    }
}
