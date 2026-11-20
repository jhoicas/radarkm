cmake_minimum_required(VERSION 3.22.1)
project(radarkm_security)

# Play Store 16 KB page-size compliance (NDK r28+ / ELF max-page-size).
set(CMAKE_SHARED_LINKER_FLAGS "${CMAKE_SHARED_LINKER_FLAGS} -Wl,-z,max-page-size=16384")

add_library(radarkm_security SHARED
    src/main/cpp/radarkm_security.cpp
)

target_include_directories(radarkm_security PRIVATE
    src/main/cpp/include
)

find_library(log-lib log)

target_link_options(radarkm_security PRIVATE
    "-Wl,-z,max-page-size=16384"
)

target_link_libraries(radarkm_security ${log-lib})
