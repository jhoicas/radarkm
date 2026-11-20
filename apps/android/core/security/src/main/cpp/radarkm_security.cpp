#include "include/radarkm_security.h"

#include <cstdio>
#include <cstring>
#include <fstream>
#include <string>

namespace {

constexpr const char *kPackage = "com.radarkm.app";
constexpr const char *kUser0Prefix = "/data/user/0/";
constexpr const char *kDataDataPrefix = "/data/data/";

bool starts_with(const std::string &value, const char *prefix) {
    const size_t n = std::strlen(prefix);
    return value.size() >= n && value.compare(0, n, prefix) == 0;
}

bool contains_ci(const std::string &haystack, const char *needle) {
    const size_t n = std::strlen(needle);
    if (n == 0 || haystack.size() < n) {
        return false;
    }
    for (size_t i = 0; i + n <= haystack.size(); ++i) {
        bool ok = true;
        for (size_t j = 0; j < n; ++j) {
            const char a = haystack[i + j];
            const char b = needle[j];
            const char al = (a >= 'A' && a <= 'Z') ? static_cast<char>(a - 'A' + 'a') : a;
            const char bl = (b >= 'A' && b <= 'Z') ? static_cast<char>(b - 'A' + 'a') : b;
            if (al != bl) {
                ok = false;
                break;
            }
        }
        if (ok) {
            return true;
        }
    }
    return false;
}

bool has_cloner_marker(const std::string &path) {
    return contains_ci(path, "/virtual/") ||
           contains_ci(path, "/parallel/") ||
           contains_ci(path, "/clone/") ||
           contains_ci(path, "/dual/") ||
           contains_ci(path, "/island/");
}

bool package_segment_matches(const std::string &path, size_t start) {
    const size_t pkgLen = std::strlen(kPackage);
    if (path.size() < start + pkgLen) {
        return false;
    }
    if (path.compare(start, pkgLen, kPackage) != 0) {
        return false;
    }
    if (path.size() == start + pkgLen) {
        return true;
    }
    return path[start + pkgLen] == '/';
}

bool is_legitimate_data_path(const std::string &path) {
    if (path.empty() || has_cloner_marker(path)) {
        return false;
    }

    // Reject non-zero Android user ids: /data/user/<id>/...
    if (starts_with(path, "/data/user/")) {
        // Must be exactly /data/user/0/<package>...
        const std::string expected = std::string(kUser0Prefix) + kPackage;
        if (!starts_with(path, expected.c_str())) {
            return false;
        }
        if (path.size() > expected.size() && path[expected.size()] != '/') {
            return false;
        }
        // Explicitly reject /data/user/999/, /data/user/10/, etc. (already covered).
        return true;
    }

    if (starts_with(path, kDataDataPrefix)) {
        return package_segment_matches(path, std::strlen(kDataDataPrefix));
    }

    return false;
}

bool tracer_pid_is_zero() {
    std::ifstream status("/proc/self/status");
    if (!status) {
        // If we cannot read status, treat as suspicious.
        return false;
    }
    std::string line;
    while (std::getline(status, line)) {
        if (starts_with(line, "TracerPid:")) {
            const char *p = line.c_str() + std::strlen("TracerPid:");
            while (*p == ' ' || *p == '\t') {
                ++p;
            }
            return std::strcmp(p, "0") == 0 || std::strcmp(p, "0\r") == 0;
        }
    }
    return false;
}

bool maps_look_clean() {
    std::ifstream maps("/proc/self/maps");
    if (!maps) {
        return false;
    }
    std::string line;
    while (std::getline(maps, line)) {
        if (contains_ci(line, "frida") ||
            contains_ci(line, "xposed") ||
            contains_ci(line, "substrate") ||
            contains_ci(line, "libhook")) {
            return false;
        }
    }
    return true;
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
    return is_legitimate_data_path(path) ? JNI_TRUE : JNI_FALSE;
}

extern "C" JNIEXPORT jboolean JNICALL
Java_com_radarkm_core_security_NativeSecurity_quickTamperProbe(
    JNIEnv * /* env */,
    jobject /* thiz */
) {
    if (!tracer_pid_is_zero()) {
        return JNI_FALSE;
    }
    if (!maps_look_clean()) {
        return JNI_FALSE;
    }
    return JNI_TRUE;
}
