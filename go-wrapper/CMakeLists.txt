# CMake build configuration for Go wrapper
cmake_minimum_required(VERSION 3.16)

project(shaka-packager-go-wrapper)

# Find the packager library
find_library(PACKAGER_LIB packager PATHS ../build REQUIRED)

# Build the C++ wrapper as a shared library
add_library(packager_wrapper SHARED packager_wrapper.cpp)

# Include directories
target_include_directories(packager_wrapper PRIVATE 
    ../include
    ../packager
)

# Link against the packager library
target_link_libraries(packager_wrapper ${PACKAGER_LIB})

# Set C++ standard
set_property(TARGET packager_wrapper PROPERTY CXX_STANDARD 17)

# Install the library
install(TARGETS packager_wrapper 
        LIBRARY DESTINATION lib
        ARCHIVE DESTINATION lib
        RUNTIME DESTINATION bin)