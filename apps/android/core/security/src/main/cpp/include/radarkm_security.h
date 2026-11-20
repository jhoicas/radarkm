#pragma once

#include <jni.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

JNIEXPORT jboolean JNICALL
Java_com_radarkm_core_security_NativeSecurity_verifyDataPath(
    JNIEnv *env,
    jobject thiz,
    jstring absolutePath
);

JNIEXPORT jboolean JNICALL
Java_com_radarkm_core_security_NativeSecurity_quickTamperProbe(
    JNIEnv *env,
    jobject thiz
);

#ifdef __cplusplus
}
#endif
