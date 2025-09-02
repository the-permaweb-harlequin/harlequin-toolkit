package templates

// C template definitions

const cProcessTemplate = `// {{.ProjectName}} C Trampoline
// Author: {{.AuthorName}}

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>

// Initialize Lua state
lua_State* init_lua() {
    lua_State *L = luaL_newstate();
    luaL_openlibs(L);

    // Load the main Lua process file
    if (luaL_dofile(L, "process.lua") != LUA_OK) {
        fprintf(stderr, "Error loading process.lua: %s\n", lua_tostring(L, -1));
        lua_close(L);
        return NULL;
    }

    return L;
}

// Handle function - called from WebAssembly
char* handle(char* message) {
    lua_State *L = init_lua();
    if (!L) {
        return strdup("{\"error\":\"Failed to initialize Lua\"}");
    }

    // Get the handle function
    lua_getglobal(L, "handle");
    if (!lua_isfunction(L, -1)) {
        lua_close(L);
        return strdup("{\"error\":\"No handle function found\"}");
    }

    // Push message as argument
    lua_pushstring(L, message);

    // Call the function
    if (lua_pcall(L, 1, 1, 0) != LUA_OK) {
        const char* error = lua_tostring(L, -1);
        char* result = malloc(strlen(error) + 50);
        sprintf(result, "{\"error\":\"%s\"}", error);
        lua_close(L);
        return result;
    }

    // Get result
    const char* result = lua_tostring(L, -1);
    char* output = strdup(result ? result : "null");

    lua_close(L);
    return output;
}
`

const cHandlersTemplate = `// {{.ProjectName}} C Handlers
// Author: {{.AuthorName}}

#include "handlers.h"
#include <string.h>
#include <stdio.h>

// Example handler functions that can be called from Lua
int ping_handler(lua_State *L) {
    lua_pushstring(L, "pong from C!");
    return 1;
}

int info_handler(lua_State *L) {
    lua_newtable(L);
    lua_pushstring(L, "{{.ProjectName}}");
    lua_setfield(L, -2, "name");
    lua_pushstring(L, "{{.AuthorName}}");
    lua_setfield(L, -2, "author");
    lua_pushstring(L, "1.0.0");
    lua_setfield(L, -2, "version");
    return 1;
}

// Register C functions with Lua
void register_handlers(lua_State *L) {
    lua_register(L, "c_ping", ping_handler);
    lua_register(L, "c_info", info_handler);
}
`

const cMakefileTemplate = `# {{.ProjectName}} C Trampoline Makefile
# Author: {{.AuthorName}}

CC = gcc
CFLAGS = -Wall -Wextra -std=c99 -fPIC
LDFLAGS = -shared
LUA_INCLUDE = /usr/include/lua5.3
LUA_LIB = -llua5.3

# Detect platform
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	LDFLAGS += -undefined dynamic_lookup
endif

TARGET = process.so
SOURCES = process.c handlers.c
OBJECTS = $(SOURCES:.c=.o)

all: $(TARGET)

$(TARGET): $(OBJECTS)
	$(CC) $(LDFLAGS) -o $@ $^ $(LUA_LIB)

%.o: %.c
	$(CC) $(CFLAGS) -I$(LUA_INCLUDE) -c $< -o $@

clean:
	rm -f $(OBJECTS) $(TARGET)

install:
	@echo "Building C trampoline for {{.ProjectName}}"
	@$(MAKE) all

.PHONY: all clean install
`

const cReadmeTemplate = `# {{.ProjectName}}

An AO process built with C.

## Author

{{.AuthorName}}{{if .GitHubUser}} ([@{{.GitHubUser}}](https://github.com/{{.GitHubUser}})){{end}}

## Description

This is a C-based AO process optimized for performance and memory efficiency.

## Project Structure

- ` + "`src/`" + ` - Source files
- ` + "`include/`" + ` - Header files
- ` + "`test/`" + ` - Test files
- ` + "`docs/`" + ` - Documentation
- ` + "`CMakeLists.txt`" + ` - CMake build configuration
- ` + "`conanfile.txt`" + ` - Conan dependencies

## Building

### Prerequisites

- CMake 3.10+
- Conan (for dependency management)
- C compiler (GCC or Clang)

### Build Instructions

` + "```bash" + `
# Install dependencies
conan install . --build=missing

# Configure and build
cmake -B build
cmake --build build
` + "```" + `

### Testing

` + "```bash" + `
cd build
ctest
` + "```" + `

## Deployment

Deploy your process to AO using the Harlequin CLI:

` + "```bash" + `
harlequin build
harlequin upload-module
` + "```" + `
`

const cCMakeListsTemplate = `cmake_minimum_required(VERSION 3.10)
project({{.ProjectName}} VERSION 1.0.0)

# Set C standard
set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Add include directories
include_directories(include)

# Add executable
add_executable({{.ProjectName}}
    src/process.c
    src/handlers.c
)

# Add libraries if needed
# find_package(PkgConfig REQUIRED)
# pkg_check_modules(DEPS REQUIRED ...)

# Set compiler flags
target_compile_options({{.ProjectName}} PRIVATE
    -Wall
    -Wextra
    -O2
)

# Install targets
install(TARGETS {{.ProjectName}} DESTINATION bin)

# Enable testing
enable_testing()

# Add tests
# add_subdirectory(test)
`

const cConanFileTemplate = `[requires]

[generators]
cmake

[options]

[imports]
`

const cSrcProcessTemplate = `// {{.ProjectName}} Main Process
// Author: {{.AuthorName}}

#include "process.h"
#include "handlers.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char *argv[]) {
    if (argc < 2) {
        fprintf(stderr, "Usage: %s <message>\n", argv[0]);
        return 1;
    }

    char *result = handle_message(argv[1]);
    printf("%s\n", result);
    free(result);

    return 0;
}

char* handle_message(const char* message) {
    // Parse message and route to appropriate handler
    if (strstr(message, "ping")) {
        return handle_ping(message);
    } else if (strstr(message, "info")) {
        return handle_info(message);
    } else {
        char* error = malloc(100);
        snprintf(error, 100, "{\"error\":\"Unknown action\"}");
        return error;
    }
}
`

const cSrcHandlersTemplate = `// {{.ProjectName}} Message Handlers
// Author: {{.AuthorName}}

#include "handlers.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char* handle_ping(const char* message) {
    (void)message; // Unused parameter
    char* response = malloc(50);
    snprintf(response, 50, "{\"output\":\"pong\",\"data\":\"Hello from {{.ProjectName}}!\"}");
    return response;
}

char* handle_info(const char* message) {
    (void)message; // Unused parameter
    char* response = malloc(200);
    snprintf(response, 200,
        "{\"output\":\"info\",\"data\":{\"name\":\"{{.ProjectName}}\",\"author\":\"{{.AuthorName}}\",\"version\":\"1.0.0\"}}");
    return response;
}
`

const cProcessHeaderTemplate = `#ifndef PROCESS_H
#define PROCESS_H

// {{.ProjectName}} Process Header
// Author: {{.AuthorName}}

/**
 * Main message handler
 * @param message Input message to process
 * @return Allocated string with response (caller must free)
 */
char* handle_message(const char* message);

#endif // PROCESS_H
`

const cHandlersHeaderTemplate = `#ifndef HANDLERS_H
#define HANDLERS_H

// {{.ProjectName}} Handlers Header
// Author: {{.AuthorName}}

/**
 * Handle ping messages
 * @param message Input message
 * @return Allocated string with response (caller must free)
 */
char* handle_ping(const char* message);

/**
 * Handle info messages
 * @param message Input message
 * @return Allocated string with response (caller must free)
 */
char* handle_info(const char* message);

#endif // HANDLERS_H
`
