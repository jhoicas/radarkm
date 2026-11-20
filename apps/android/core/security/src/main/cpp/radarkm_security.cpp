#include "include/radarkm_security.h"

#include <cstring>
#include <string>

namespace {

constexpr const char *kExpectedPrefix = "/data/user/0/";

bool starts_with(const std::string &value, const char *prefix) {
    const size_t n = std::strlen(prefix);
    return value.size() >= n && value.compare(0, n, prefix) == 0;
}

} // namespace

extern "C" JNIEXPORT jboolean JNICALL
Java_com_radarkm_core_security_NativeSecurity_verifyDataPath(
    JNIEnv *env,
    jobject /* thiz */,
    jstring absolutePath
) {
    if (absolutePath == nullptr) {
        return JNI_FALSE;
    }
    const char *raw = env->GetStringUTFChars(absolutePath, nullptr);
    if (raw == nullptr) {
        return JNI_FALSE;
    }
    const std::string path(raw);
    env->ReleaseStringUTFChars(absolutePath, raw);
    return starts_with(path, kExpectedPrefix) ? JNI_TRUE : JNI_FALSE;
}

extern "C" JNIEXPORT jboolean JNICALL
Java_com_radarkm_core_security_NativeSecurity_quickTamperProbe(
    JNIEnv * /* env */,
    jobject /* thiz */
) {
    // Scaffolding stub: real anti-clone / frida probes come in later tasks.
    return JNI_TRUE;
}
