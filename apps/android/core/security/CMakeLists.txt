cmake_minimum_required(VERSION 3.22.1)
project(radarkm_security)

add_library(radarkm_security SHARED
    src/main/cpp/radarkm_security.cpp
)

target_include_directories(radarkm_security PRIVATE
    src/main/cpp/include
)

find_library(log-lib log)

# Enforce 16 KB ELF max page size for Play compliance.
target_link_options(radarkm_security PRIVATE
    "-Wl,-z,max-page-size=16384"
)

target_link_libraries(radarkm_security ${log-lib})
