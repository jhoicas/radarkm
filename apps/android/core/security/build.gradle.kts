plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.android)
}

android {
    namespace = "com.radarkm.core.security"
    compileSdk = libs.versions.compileSdk.get().toInt()
    // Pinned to NDK r28c for 16 KB page-size compliance (keep in sync with libs.versions.toml).
    ndkVersion = "28.2.13676358"

    defaultConfig {
        minSdk = libs.versions.minSdk.get().toInt()
        consumerProguardFiles("consumer-rules.pro")

        externalNativeBuild {
            cmake {
                // NDK r28+ preferred; keep FLEXIBLE_PAGE_SIZES for r27 compatibility.
                arguments += listOf(
                    "-DANDROID_SUPPORT_FLEXIBLE_PAGE_SIZES=ON",
                    "-DANDROID_STL=c++_shared",
                )
                cppFlags += listOf("-std=c++17")

            }
        }

        ndk {
            abiFilters += listOf("arm64-v8a", "armeabi-v7a", "x86_64")
        }
    }

    externalNativeBuild {
        cmake {
            path = file("CMakeLists.txt")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.play.integrity)
    implementation(libs.kotlinx.coroutines.android)
}
