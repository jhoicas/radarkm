pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "RadarKm"

include(":app")
include(":core:finance")
include(":core:security")
include(":core:network")
include(":feature:capture")
include(":feature:overlay")
include(":feature:billing")
