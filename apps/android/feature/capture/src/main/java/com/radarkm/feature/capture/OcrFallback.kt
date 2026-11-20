package com.radarkm.feature.capture

import android.graphics.Bitmap
import com.google.mlkit.vision.common.InputImage
import com.google.mlkit.vision.text.TextRecognition
import com.google.mlkit.vision.text.latin.TextRecognizerOptions
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlin.coroutines.resume

/**
 * Fallback OCR v2 en RAM. El caller DEBE reciclar el [Bitmap] y nunca escribirlo a disco.
 */
class OcrFallback(
    private val recognizer: com.google.mlkit.vision.text.TextRecognizer =
        TextRecognition.getClient(TextRecognizerOptions.DEFAULT_OPTIONS),
) {
    suspend fun recognizeInMemory(bitmap: Bitmap): String =
        suspendCancellableCoroutine { cont ->
            val image = InputImage.fromBitmap(bitmap, 0)
            recognizer.process(image)
                .addOnSuccessListener { result ->
                    if (cont.isActive) cont.resume(result.text)
                }
                .addOnFailureListener { error ->
                    if (cont.isActive) cont.resume("")
                    // Scaffolding: surface error via logging in later tasks.
                    error.printStackTrace()
                }
        }
}
