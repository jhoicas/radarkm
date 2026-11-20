package com.radarkm.feature.overlay

import android.content.Context
import android.speech.tts.TextToSpeech
import java.util.Locale

/**
 * Esqueleto TTS para anunciar semáforo / contraoferta.
 */
class TtsSpeaker(context: Context) : TextToSpeech.OnInitListener {

    private val tts = TextToSpeech(context.applicationContext, this)
    private var ready = false

    override fun onInit(status: Int) {
        ready = status == TextToSpeech.SUCCESS
        if (ready) {
            tts.language = Locale("es", "CO")
        }
    }

    fun speak(text: String) {
        if (!ready) return
        tts.speak(text, TextToSpeech.QUEUE_FLUSH, null, "radarkm-tts")
    }

    fun shutdown() {
        tts.stop()
        tts.shutdown()
    }
}
