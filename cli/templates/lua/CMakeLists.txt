cmake_minimum_required(VERSION 3.16)
project({{PROJECT_NAME}})

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Find Lua
find_package(Lua REQUIRED)

# Set up Emscripten-specific settings
if(EMSCRIPTEN)
    # WebAssembly build settings
    set(CMAKE_EXECUTABLE_SUFFIX ".wasm")

    # Emscripten flags for AO compatibility
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -O3 -s WASM=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORTED_FUNCTIONS='[\"_init_lua_process\",\"_handle_message\",\"_ao_handle\",\"_eval_lua\",\"_get_lua_state\",\"_cleanup_lua_process\",\"_main\"]'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORTED_RUNTIME_METHODS='[\"ccall\",\"cwrap\",\"UTF8ToString\",\"stringToUTF8\",\"allocate\",\"ALLOC_NORMAL\"]'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ALLOW_MEMORY_GROWTH=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s INITIAL_MEMORY=33554432")  # 32MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s MAXIMUM_MEMORY=134217728") # 128MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s STACK_SIZE=16777216")      # 16MB
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s MODULARIZE=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s EXPORT_NAME='AOLuaProcess'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ENVIRONMENT='web,worker'")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s NO_EXIT_RUNTIME=1")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s ASSERTIONS=0")

    # Embed Lua files
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} --embed-file main.lua")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} --embed-file handlers/")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} --embed-file utils/")

    # Link Lua statically for Emscripten
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -s USE_LUA=1")

    # Debug flags (uncomment for debugging)
    # set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -g -s ASSERTIONS=1 -s SAFE_HEAP=1")
else()
    # Native build for testing
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra -g")
endif()

# Source files
set(SOURCES
    src/trampoline.c
)

# Add executable
add_executable(${PROJECT_NAME} ${SOURCES})

# Include directories
target_include_directories(${PROJECT_NAME} PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${LUA_INCLUDE_DIR}
)

# Link libraries
if(NOT EMSCRIPTEN)
    target_link_libraries(${PROJECT_NAME} ${LUA_LIBRARIES})
endif()

# Custom target for LuaRocks setup
add_custom_target(luarocks_setup
    COMMAND luarocks install --only-deps ${PROJECT_NAME}-1.0.0-1.rockspec
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Installing LuaRocks dependencies"
)

# Custom target for building with Harlequin
add_custom_target(harlequin_build
    COMMAND harlequin build --entrypoint main.lua
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Building with Harlequin CLI"
    DEPENDS luarocks_setup
)

# Custom target for testing Lua code
add_custom_target(test_lua
    COMMAND busted test/
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Running Lua tests"
    DEPENDS luarocks_setup
)

# Custom target for testing C trampoline
add_custom_target(test_trampoline
    COMMAND ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}
    DEPENDS ${PROJECT_NAME}
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Testing C trampoline"
)

# Custom target for bundling Lua files
add_custom_target(bundle_lua
    COMMAND harlequin lua-utils bundle --entrypoint main.lua --output dist/bundle.lua
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Bundling Lua files"
)

# Install rules
if(EMSCRIPTEN)
    install(FILES
        ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.wasm
        ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.js
        DESTINATION dist
    )

    # Install Lua files
    install(FILES
        main.lua
        DESTINATION dist
    )

    install(DIRECTORY
        handlers/
        utils/
        DESTINATION dist
    )
else()
    install(TARGETS ${PROJECT_NAME}
        RUNTIME DESTINATION bin
    )
endif()

# Custom target for cleaning build artifacts
add_custom_target(clean_all
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}
    COMMAND ${CMAKE_COMMAND} -E remove_directory dist
    COMMAND ${CMAKE_COMMAND} -E remove *.rock
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Cleaning all build artifacts"
)

