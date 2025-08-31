cmake_minimum_required(VERSION 3.16)
project({{PROJECT_NAME}})

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Conan integration
find_package(nlohmann_json REQUIRED)
find_package(fmt REQUIRED)
find_package(spdlog REQUIRED)

# Optional: GTest for testing
find_package(GTest)

# Set up Emscripten-specific settings
if(EMSCRIPTEN)
    # WebAssembly build settings
    set(CMAKE_EXECUTABLE_SUFFIX ".wasm")

    # Emscripten flags for AO compatibility
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -O3 -s WASM=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORTED_FUNCTIONS='[\"_handle_message\",\"_init_process\",\"_main\"]'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORTED_RUNTIME_METHODS='[\"ccall\",\"cwrap\",\"UTF8ToString\",\"stringToUTF8\"]'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ALLOW_MEMORY_GROWTH=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s INITIAL_MEMORY=16777216")  # 16MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s MAXIMUM_MEMORY=67108864")  # 64MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s STACK_SIZE=8388608")       # 8MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s MODULARIZE=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORT_NAME='AOProcess'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ENVIRONMENT='web,worker'")

    # Optimization flags
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s NO_EXIT_RUNTIME=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ASSERTIONS=0")

    # Debug flags (uncomment for debugging)
    # set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -g -s ASSERTIONS=1 -s SAFE_HEAP=1")
else()
    # Native build for testing
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra -g")
endif()

# Source files
set(SOURCES
    main.c
)

# Add executable
add_executable(${PROJECT_NAME} ${SOURCES})

# Include directories
target_include_directories(${PROJECT_NAME} PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)

# Link Conan dependencies
target_link_libraries(${PROJECT_NAME}
    nlohmann_json::nlohmann_json
    fmt::fmt
    spdlog::spdlog
)

# Custom target for Conan install
add_custom_target(conan_install
    COMMAND conan install . --output-folder=build --build=missing --profile:build=default --profile:host=default
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Installing Conan dependencies"
)

# Custom target for building with Harlequin
add_custom_target(harlequin_build
    COMMAND harlequin build --entrypoint main.c
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Building with Harlequin CLI"
    DEPENDS conan_install
)

# Custom target for testing
add_custom_target(run_tests
    COMMAND ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}
    DEPENDS ${PROJECT_NAME}
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Running tests"
)

# Optional: GTest integration
if(GTest_FOUND)
    # Test executable
    add_executable(${PROJECT_NAME}_tests
        test/test_main.c
        main.c  # Include main.c for testing functions
    )

    target_link_libraries(${PROJECT_NAME}_tests
        GTest::gtest
        GTest::gtest_main
        nlohmann_json::nlohmann_json
        fmt::fmt
        spdlog::spdlog
    )

    # Define TEST_MODE for conditional compilation
    target_compile_definitions(${PROJECT_NAME}_tests PRIVATE TEST_MODE=1)

    # Add test target
    add_custom_target(gtest
        COMMAND ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}_tests
        DEPENDS ${PROJECT_NAME}_tests
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Running GTest suite"
    )
endif()

# Install rules
if(EMSCRIPTEN)
    install(FILES
        ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.wasm
        ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.js
        DESTINATION dist
    )
else()
    install(TARGETS ${PROJECT_NAME}
        RUNTIME DESTINATION bin
    )
endif()
