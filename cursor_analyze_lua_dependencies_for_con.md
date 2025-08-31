# Analyze lua dependencies for conan package

_Exported on 8/30/2025 at 18:38:52 CST from Cursor (1.4.5)_

---

**User**

Analyze the dependencies for just lua here.

Is it possible to make this a package that is installable with conan to be included in a build system?

---

**Cursor**

I'll analyze the Lua dependencies in the current codebase and assess the feasibility of creating a Conan package. Let me start by examining the current file you have open and then explore the broader Lua usage in the project.

Let me examine the build system more closely to understand the Lua dependencies and compilation process.

Now let me examine the loader and build system more closely to understand the complete Lua integration:

Let me also check for any configuration files that might define additional dependencies:

Based on my analysis of the ao project's Lua dependencies and build system, here's a comprehensive assessment:

## Current Lua Dependencies Analysis

The ao project uses Lua in a specialized way for WebAssembly compilation. Here are the key dependencies I identified:

### Core Lua Dependencies:

1. **Lua 5.3.4** - The specific version used in the project
2. **Standard Lua libraries** (`lua.h`, `lualib.h`, `lauxlib.h`)
3. **Custom JSON library** - A pure Lua JSON implementation (308 lines)
4. **LuaRocks 2.4.4** - For package management
5. **Emscripten SDK 3.1.59** - For WASM compilation

### Build System Dependencies:

1. **Emscripten compiler** (`emcc`) with specific flags:

   - WASM=1, MEMORY64=1 (for 64-bit), SUPPORT_LONGJMP=1
   - ASYNCIFY=1 for async operations
   - Custom memory settings and optimizations

2. **Custom C wrapper** (`main.c`) that:

   - Embeds Lua runtime in WASM
   - Handles message passing via JSON
   - Manages Lua state lifecycle
   - Provides error handling and debugging

3. **Build toolchain**:
   - Python 3.6.6 for build scripts
   - Custom aolibc (patched libc for ao-specific functions)
   - Both 32-bit and 64-bit Lua builds

## Conan Package Feasibility

**Yes, it's possible to create a Conan package**, but with important considerations:

### ✅ **What Can Be Packaged:**

1. **Lua Runtime**: The core Lua 5.3.4 interpreter and libraries
2. **JSON Library**: The pure Lua JSON implementation
3. **Build Scripts**: The Python-based build system
4. **C Wrapper Code**: The main.c and supporting C files

### ⚠️ **Challenges and Considerations:**

1. **Emscripten Dependency**: The package would need Emscripten as a build requirement, which adds complexity
2. **Custom Compilation Flags**: The specific WASM compilation flags are tightly coupled to ao's requirements
3. **Multiple Architectures**: Supporting both 32-bit and 64-bit WASM builds
4. **Custom libc**: The aolibc patches would need to be included

### 📦 **Recommended Conan Package Structure:**

```python
# conanfile.py
class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    settings = "os", "compiler", "build_type", "arch"

    requires = (
        "lua/5.3.4",  # Base Lua from ConanCenter
        # Emscripten would need to be a build_requires
    )

    build_requires = (
        "emscripten/3.1.59@",  # If available
    )

    def build(self):
        # Custom build logic for WASM compilation
        # Include the ao-specific patches and flags
        pass

    def package(self):
        # Package the compiled WASM libraries
        # Include headers and build tools
        pass
```

### 🎯 **Recommended Approach:**

1. **Create a specialized Conan package** specifically for ao-lua that:

   - Extends the standard Lua package
   - Includes Emscripten build configuration
   - Packages the custom C wrapper and build tools
   - Provides both 32-bit and 64-bit WASM outputs

2. **Package Components:**

   ```
   ao-lua/
   ├── include/           # Lua headers + ao extensions
   ├── lib/
   │   ├── lua53-32.a    # 32-bit WASM Lua
   │   ├── lua53-64.a    # 64-bit WASM Lua
   │   └── aolibc.a      # Custom libc patches
   ├── src/
   │   ├── main.c        # WASM wrapper
   │   ├── json.lua      # JSON library
   │   └── loader.lua    # Lua loader
   └── tools/
       └── ao-build-module  # Build script
   ```

3. **Usage in consuming projects:**
   ```cmake
   find_package(ao-lua REQUIRED)
   target_link_libraries(my_target ao-lua::ao-lua)
   ```

### 🔧 **Implementation Steps:**

1. **Start with base Lua**: Use existing `lua/5.3.4` from ConanCenter
2. **Add Emscripten support**: Create build logic for WASM compilation
3. **Include ao-specific patches**: Bundle the custom C code and libraries
4. **Test cross-platform**: Ensure it works on different development environments
5. **Document usage**: Provide clear examples for integration

This would make the ao Lua runtime much more accessible for other projects wanting to build similar WASM-based Lua environments, while maintaining the specific optimizations and patches that ao requires.

```python
# conanfile.py
class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    settings = "os", "compiler", "build_type", "arch"

    requires = (
        "lua/5.3.4",  # Base Lua from ConanCenter
        # Emscripten would need to be a build_requires
    )

    build_requires = (
        "emscripten/3.1.59@",  # If available
    )

    def build(self):
        # Custom build logic for WASM compilation
        # Include the ao-specific patches and flags
        pass

    def package(self):
        # Package the compiled WASM libraries
        # Include headers and build tools
        pass
```

```plaintext
   ao-lua/
   ├── include/           # Lua headers + ao extensions
   ├── lib/
   │   ├── lua53-32.a    # 32-bit WASM Lua
   │   ├── lua53-64.a    # 64-bit WASM Lua
   │   └── aolibc.a      # Custom libc patches
   ├── src/
   │   ├── main.c        # WASM wrapper
   │   ├── json.lua      # JSON library
   │   └── loader.lua    # Lua loader
   └── tools/
       └── ao-build-module  # Build script
```

```cmake
   find_package(ao-lua REQUIRED)
   target_link_libraries(my_target ao-lua::ao-lua)
```

---

**User**

Right - i think i would make aolibc a conan package itself, and maybe extend that a bit to be fully ao spec compatible, then make the lua wrapper a package.

Regarding the build tools, fine to leave them out. However for the initial build, if I have a luarocks project and am intending on using it in conjunction with this - would you still recommend injection of the lua files, or is there a more elegant way to handle that?

---

**Cursor**

Great approach! Separating `aolibc` as its own Conan package and then building the Lua wrapper on top makes a lot of sense architecturally. Let me analyze the current Lua file injection mechanism and suggest more elegant alternatives for LuaRocks integration.

Perfect! Now I understand the current injection mechanism. Let me analyze this and provide better alternatives for LuaRocks integration.

## Current Lua File Injection Analysis

The current system does the following:

1. **Installs LuaRocks dependencies** to `/src/modules`
2. **Scans for all Lua files** recursively in `/src/` and LuaRocks directories
3. **Converts each Lua file to C hex literals** and embeds them directly into the C program
4. **Creates static arrays** for each Lua module and registers them in the Lua state
5. **Links binary modules** (.so files) from LuaRocks

This approach works but has several drawbacks:

- **Large binary size** (every Lua file becomes embedded C data)
- **Build-time coupling** (all dependencies must be known at compile time)
- **No runtime flexibility** (can't load modules dynamically)

## Better Alternatives for LuaRocks Integration

### 🎯 **Recommended Approach: Hybrid Module System**

Instead of pure injection, I'd suggest a **hybrid approach** that maintains ao's sandboxing while being more elegant:

#### **1. Virtual File System (VFS) Approach**

```c
// In your ao-lua Conan package
typedef struct {
    const char* name;
    const char* content;
    size_t size;
} lua_module_t;

// Pre-compiled module registry
static lua_module_t lua_modules[] = {
    {"json", json_lua_content, sizeof(json_lua_content)},
    {"myproject.utils", utils_lua_content, sizeof(utils_lua_content)},
    // LuaRocks modules get added here
    {NULL, NULL, 0}
};

// Custom require() implementation
int ao_require(lua_State* L) {
    const char* module_name = luaL_checkstring(L, 1);

    // Look up in embedded modules first
    for (int i = 0; lua_modules[i].name; i++) {
        if (strcmp(lua_modules[i].name, module_name) == 0) {
            luaL_loadbuffer(L, lua_modules[i].content,
                          lua_modules[i].size, module_name);
            lua_call(L, 0, 1);
            return 1;
        }
    }

    return luaL_error(L, "module '%s' not found", module_name);
}
```

#### **2. Build-Time Module Bundling**

Create a **separate build step** that processes your LuaRocks project:

```yaml
# conanfile.txt
[requires]
aolibc/1.0.0
ao-lua/5.3.4

[generators]
CMakeDeps
CMakeToolchain

[options]
ao-lua:bundle_luarocks=True
ao-lua:luarocks_manifest=luarocks-manifest.txt
```

```bash
# luarocks-manifest.txt (in your project)
json
lpeg
lua-cjson
```

The ao-lua Conan package would:

1. **Install specified LuaRocks modules** during build
2. **Generate a module registry** header file
3. **Create bundled source** with all dependencies

#### **3. CMake Integration Example**

```cmake
# Your project's CMakeLists.txt
find_package(ao-lua REQUIRED)

# Define your Lua modules
set(LUA_MODULES
    src/main.lua
    src/utils.lua
)

# Bundle with LuaRocks dependencies
ao_lua_bundle_modules(
    TARGET my_ao_process
    MODULES ${LUA_MODULES}
    LUAROCKS_DEPS json lpeg
    OUTPUT_NAME process.wasm
)
```

### 🔧 **Implementation Strategy**

#### **Package Structure:**

```
ao-lua/
├── conanfile.py
├── include/
│   ├── ao_lua.h          # Main API
│   └── ao_module_registry.h  # Generated module registry
├── src/
│   ├── ao_lua_runtime.c  # Your main.c equivalent
│   ├── ao_require.c      # Custom require() implementation
│   └── json.lua          # Bundled JSON library
├── tools/
│   ├── bundle_modules.py # Module bundling script
│   └── cmake/
│       └── ao-lua-config.cmake
└── templates/
    └── module_registry.h.in
```

#### **Conanfile.py:**

```python
from conan import ConanFile
from conan.tools.cmake import CMake, cmake_layout
from conan.tools.files import copy
import os

class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    settings = "os", "compiler", "build_type", "arch"
    options = {
        "bundle_luarocks": [True, False],
        "target_bits": [32, 64]
    }
    default_options = {
        "bundle_luarocks": True,
        "target_bits": 64
    }

    requires = "aolibc/1.0.0"

    def build_requirements(self):
        self.tool_requires("emscripten/3.1.59")
        if self.options.bundle_luarocks:
            self.tool_requires("luarocks/2.4.4")

    def build(self):
        # Build Lua with Emscripten
        # Generate module registry from LuaRocks deps
        # Compile to WASM
        pass

    def package(self):
        # Package headers, libraries, and tools
        copy(self, "*.h", src=self.build_folder, dst=os.path.join(self.package_folder, "include"))
        copy(self, "*.a", src=self.build_folder, dst=os.path.join(self.package_folder, "lib"))
        copy(self, "bundle_modules.py", src=self.source_folder, dst=os.path.join(self.package_folder, "tools"))
```

### 🚀 **Usage in Your Project**

#### **Option 1: Declarative (Recommended)**

```yaml
# ao-project.yml
dependencies:
  - json
  - lpeg
  - my-custom-lib

modules:
  - src/main.lua
  - src/handlers/
  - lib/utils.lua

build:
  target: wasm64
  optimize: true
```

#### **Option 2: Programmatic**

```python
# build.py
from ao_lua import ModuleBundler

bundler = ModuleBundler()
bundler.add_luarocks_deps(['json', 'lpeg'])
bundler.add_local_modules(['src/main.lua', 'src/handlers/'])
bundler.build('process.wasm', target_bits=64)
```

### 📊 **Benefits of This Approach**

1. **Clean separation**: aolibc and ao-lua as separate, reusable packages
2. **Standard LuaRocks workflow**: Use normal `luarocks install` in development
3. **Build-time optimization**: Dependencies resolved and bundled at build time
4. **Smaller binaries**: Only include what you actually use
5. **Better debugging**: Source maps and module names preserved
6. **Incremental builds**: Only rebuild when dependencies change

### 🎯 **Migration Path**

1. **Phase 1**: Extract aolibc as Conan package
2. **Phase 2**: Create ao-lua package with current injection method
3. **Phase 3**: Add VFS-based module system
4. **Phase 4**: Add CMake/build tool integration
5. **Phase 5**: Optimize for incremental builds

This approach gives you the best of both worlds: the sandboxing and determinism of the current system, but with much better developer experience and maintainability for LuaRocks projects.

```c
// In your ao-lua Conan package
typedef struct {
    const char* name;
    const char* content;
    size_t size;
} lua_module_t;

// Pre-compiled module registry
static lua_module_t lua_modules[] = {
    {"json", json_lua_content, sizeof(json_lua_content)},
    {"myproject.utils", utils_lua_content, sizeof(utils_lua_content)},
    // LuaRocks modules get added here
    {NULL, NULL, 0}
};

// Custom require() implementation
int ao_require(lua_State* L) {
    const char* module_name = luaL_checkstring(L, 1);

    // Look up in embedded modules first
    for (int i = 0; lua_modules[i].name; i++) {
        if (strcmp(lua_modules[i].name, module_name) == 0) {
            luaL_loadbuffer(L, lua_modules[i].content,
                          lua_modules[i].size, module_name);
            lua_call(L, 0, 1);
            return 1;
        }
    }

    return luaL_error(L, "module '%s' not found", module_name);
}
```

```yaml
# conanfile.txt
[requires]
aolibc/1.0.0
ao-lua/5.3.4

[generators]
CMakeDeps
CMakeToolchain

[options]
ao-lua:bundle_luarocks=True
ao-lua:luarocks_manifest=luarocks-manifest.txt
```

```shellscript
# luarocks-manifest.txt (in your project)
json
lpeg
lua-cjson
```

```cmake
# Your project's CMakeLists.txt
find_package(ao-lua REQUIRED)

# Define your Lua modules
set(LUA_MODULES
    src/main.lua
    src/utils.lua
)

# Bundle with LuaRocks dependencies
ao_lua_bundle_modules(
    TARGET my_ao_process
    MODULES ${LUA_MODULES}
    LUAROCKS_DEPS json lpeg
    OUTPUT_NAME process.wasm
)
```

```plaintext
ao-lua/
├── conanfile.py
├── include/
│   ├── ao_lua.h          # Main API
│   └── ao_module_registry.h  # Generated module registry
├── src/
│   ├── ao_lua_runtime.c  # Your main.c equivalent
│   ├── ao_require.c      # Custom require() implementation
│   └── json.lua          # Bundled JSON library
├── tools/
│   ├── bundle_modules.py # Module bundling script
│   └── cmake/
│       └── ao-lua-config.cmake
└── templates/
    └── module_registry.h.in
```

```python
from conan import ConanFile
from conan.tools.cmake import CMake, cmake_layout
from conan.tools.files import copy
import os

class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    settings = "os", "compiler", "build_type", "arch"
    options = {
        "bundle_luarocks": [True, False],
        "target_bits": [32, 64]
    }
    default_options = {
        "bundle_luarocks": True,
        "target_bits": 64
    }

    requires = "aolibc/1.0.0"

    def build_requirements(self):
        self.tool_requires("emscripten/3.1.59")
        if self.options.bundle_luarocks:
            self.tool_requires("luarocks/2.4.4")

    def build(self):
        # Build Lua with Emscripten
        # Generate module registry from LuaRocks deps
        # Compile to WASM
        pass

    def package(self):
        # Package headers, libraries, and tools
        copy(self, "*.h", src=self.build_folder, dst=os.path.join(self.package_folder, "include"))
        copy(self, "*.a", src=self.build_folder, dst=os.path.join(self.package_folder, "lib"))
        copy(self, "bundle_modules.py", src=self.source_folder, dst=os.path.join(self.package_folder, "tools"))
```

```yaml
# ao-project.yml
dependencies:
  - json
  - lpeg
  - my-custom-lib

modules:
  - src/main.lua
  - src/handlers/
  - lib/utils.lua

build:
  target: wasm64
  optimize: true
```

```python
# build.py
from ao_lua import ModuleBundler

bundler = ModuleBundler()
bundler.add_luarocks_deps(['json', 'lpeg'])
bundler.add_local_modules(['src/main.lua', 'src/handlers/'])
bundler.build('process.wasm', target_bits=64)
```

---

**User**

Can we make that a more dynamic workflow i wonder?

This in general is a "C program that runs an interpreter calling a specific function global in the interpreter" right? So in theory it should be able to abstract it to any interpreted language?

---

**Cursor**

Absolutely! You've identified the core abstraction perfectly. This is indeed a **generic "C program that runs an interpreter calling a specific function"** pattern, and it can definitely be abstracted to support multiple interpreted languages dynamically.

Let me analyze the current architecture and show how to make it truly dynamic and language-agnostic.

Perfect! Now I can see the complete architecture. You're absolutely right - this is a **generic interpreter runtime pattern** that can be abstracted. Let me design a dynamic, language-agnostic system.

## 🎯 **Generic Interpreter Runtime Architecture**

The current system has this flow:

1. **C main()** → boots interpreter
2. **Interpreter** loads modules and calls `handle(msgJSON, aoJSON)`
3. **Returns JSON string** back to C/WASM

This can be generalized to **any interpreted language**!

## 🚀 **Dynamic Multi-Language Runtime**

### **Core Abstraction: Language Runtime Interface**

```c
// ao_runtime.h - Generic runtime interface
typedef struct ao_runtime ao_runtime_t;

typedef struct {
    const char* name;           // "lua", "python", "javascript", "wasm"

    // Lifecycle
    ao_runtime_t* (*create)(void);
    int (*destroy)(ao_runtime_t* runtime);

    // Module loading
    int (*load_module)(ao_runtime_t* runtime, const char* name,
                      const char* source, size_t size);
    int (*load_binary)(ao_runtime_t* runtime, const char* name,
                      const void* data, size_t size);

    // Function calling
    char* (*call_function)(ao_runtime_t* runtime, const char* function_name,
                          const char* msg_json, const char* ao_json);

    // Memory management
    void (*free_result)(char* result);

} ao_language_vtable_t;

// Registry of supported languages
extern const ao_language_vtable_t* ao_get_language(const char* name);
```

### **Dynamic Runtime Implementation**

```c
// ao_runtime.c
#include "ao_runtime.h"

// Global runtime instance
static ao_runtime_t* g_runtime = NULL;
static const ao_language_vtable_t* g_language = NULL;

// WASM exported function - language agnostic!
EMSCRIPTEN_KEEPALIVE
char* ao_handle(const char* msg_json, const char* ao_json) {
    if (!g_runtime || !g_language) {
        return strdup("{\"ok\":false,\"error\":\"Runtime not initialized\"}");
    }

    return g_language->call_function(g_runtime, "handle", msg_json, ao_json);
}

// Boot any language runtime
EMSCRIPTEN_KEEPALIVE
int ao_boot(const char* language_name, const char* entry_module,
           const char* source, size_t source_size) {

    // Get language implementation
    g_language = ao_get_language(language_name);
    if (!g_language) {
        printf("Unsupported language: %s\n", language_name);
        return 1;
    }

    // Create runtime
    g_runtime = g_language->create();
    if (!g_runtime) {
        printf("Failed to create %s runtime\n", language_name);
        return 1;
    }

    // Load entry module
    if (g_language->load_module(g_runtime, entry_module, source, source_size)) {
        printf("Failed to load entry module\n");
        g_language->destroy(g_runtime);
        return 1;
    }

    return 0;
}
```

### **Language Implementations**

#### **Lua Implementation**

```c
// ao_lua.c
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

typedef struct {
    lua_State* L;
} ao_lua_runtime_t;

static ao_runtime_t* lua_create(void) {
    ao_lua_runtime_t* runtime = malloc(sizeof(ao_lua_runtime_t));
    runtime->L = luaL_newstate();
    luaL_openlibs(runtime->L);
    return (ao_runtime_t*)runtime;
}

static char* lua_call_function(ao_runtime_t* rt, const char* func_name,
                              const char* msg_json, const char* ao_json) {
    ao_lua_runtime_t* runtime = (ao_lua_runtime_t*)rt;
    lua_State* L = runtime->L;

    // Get global function
    lua_getglobal(L, func_name);
    if (!lua_isfunction(L, -1)) {
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Push arguments
    lua_pushstring(L, msg_json);
    lua_pushstring(L, ao_json);

    // Call function
    if (lua_pcall(L, 2, 1, 0) != LUA_OK) {
        const char* error = lua_tostring(L, -1);
        char* result = malloc(strlen(error) + 100);
        sprintf(result, "{\"ok\":false,\"error\":\"%s\"}", error);
        lua_pop(L, 1);
        return result;
    }

    // Get result
    const char* result_json = lua_tostring(L, -1);
    char* result = strdup(result_json ? result_json : "null");
    lua_pop(L, 1);

    return result;
}

static const ao_language_vtable_t lua_vtable = {
    .name = "lua",
    .create = lua_create,
    .destroy = lua_destroy,
    .load_module = lua_load_module,
    .call_function = lua_call_function,
    .free_result = free
};
```

#### **Python Implementation**

```c
// ao_python.c
#include <Python.h>

typedef struct {
    PyObject* globals;
    PyObject* handle_func;
} ao_python_runtime_t;

static ao_runtime_t* python_create(void) {
    if (!Py_IsInitialized()) {
        Py_Initialize();
    }

    ao_python_runtime_t* runtime = malloc(sizeof(ao_python_runtime_t));
    runtime->globals = PyDict_New();

    // Add json module
    PyRun_String("import json", Py_single_input, runtime->globals, NULL);

    return (ao_runtime_t*)runtime;
}

static char* python_call_function(ao_runtime_t* rt, const char* func_name,
                                 const char* msg_json, const char* ao_json) {
    ao_python_runtime_t* runtime = (ao_python_runtime_t*)rt;

    // Get function
    PyObject* func = PyDict_GetItemString(runtime->globals, func_name);
    if (!func || !PyCallable_Check(func)) {
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Create arguments
    PyObject* args = PyTuple_New(2);
    PyTuple_SetItem(args, 0, PyUnicode_FromString(msg_json));
    PyTuple_SetItem(args, 1, PyUnicode_FromString(ao_json));

    // Call function
    PyObject* result = PyObject_CallObject(func, args);
    Py_DECREF(args);

    if (!result) {
        PyErr_Print();
        return strdup("{\"ok\":false,\"error\":\"Python execution error\"}");
    }

    // Convert result to string
    const char* result_str = PyUnicode_AsUTF8(result);
    char* output = strdup(result_str ? result_str : "null");
    Py_DECREF(result);

    return output;
}

static const ao_language_vtable_t python_vtable = {
    .name = "python",
    .create = python_create,
    .destroy = python_destroy,
    .load_module = python_load_module,
    .call_function = python_call_function,
    .free_result = free
};
```

#### **JavaScript (QuickJS) Implementation**

```c
// ao_javascript.c
#include "quickjs.h"

typedef struct {
    JSRuntime* rt;
    JSContext* ctx;
} ao_js_runtime_t;

static ao_runtime_t* js_create(void) {
    ao_js_runtime_t* runtime = malloc(sizeof(ao_js_runtime_t));
    runtime->rt = JS_NewRuntime();
    runtime->ctx = JS_NewContext(runtime->rt);

    // Add JSON support
    js_std_add_helpers(runtime->ctx, 0, NULL);

    return (ao_runtime_t*)runtime;
}

static char* js_call_function(ao_runtime_t* rt, const char* func_name,
                             const char* msg_json, const char* ao_json) {
    ao_js_runtime_t* runtime = (ao_js_runtime_t*)rt;
    JSContext* ctx = runtime->ctx;

    // Get global function
    JSValue global = JS_GetGlobalObject(ctx);
    JSValue func = JS_GetPropertyStr(ctx, global, func_name);

    if (!JS_IsFunction(ctx, func)) {
        JS_FreeValue(ctx, func);
        JS_FreeValue(ctx, global);
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Create arguments
    JSValue args[2];
    args[0] = JS_NewString(ctx, msg_json);
    args[1] = JS_NewString(ctx, ao_json);

    // Call function
    JSValue result = JS_Call(ctx, func, global, 2, args);

    // Cleanup args
    JS_FreeValue(ctx, args[0]);
    JS_FreeValue(ctx, args[1]);
    JS_FreeValue(ctx, func);
    JS_FreeValue(ctx, global);

    if (JS_IsException(result)) {
        JS_FreeValue(ctx, result);
        return strdup("{\"ok\":false,\"error\":\"JavaScript execution error\"}");
    }

    // Convert to string
    const char* result_str = JS_ToCString(ctx, result);
    char* output = strdup(result_str ? result_str : "null");
    JS_FreeCString(ctx, result_str);
    JS_FreeValue(ctx, result);

    return output;
}

static const ao_language_vtable_t js_vtable = {
    .name = "javascript",
    .create = js_create,
    .destroy = js_destroy,
    .load_module = js_load_module,
    .call_function = js_call_function,
    .free_result = free
};
```

### **Language Registry**

```c
// ao_registry.c
static const ao_language_vtable_t* languages[] = {
    &lua_vtable,
    &python_vtable,
    &js_vtable,
    NULL
};

const ao_language_vtable_t* ao_get_language(const char* name) {
    for (int i = 0; languages[i]; i++) {
        if (strcmp(languages[i]->name, name) == 0) {
            return languages[i];
        }
    }
    return NULL;
}
```

## 🔧 **Dynamic Build Configuration**

### **Universal Definition Format**

```yaml
# ao-process.yml
language: lua # or python, javascript, etc.
entry_file: src/main.lua # language-specific entry point
output_file: process.wasm

# Language-specific dependencies
dependencies:
  lua:
    - json
    - lpeg
  python:
    - requests
    - numpy
  javascript:
    - lodash

# Functions to export (same interface for all languages)
functions:
  handle:
    args: [string, string]
    return: string

# Build configuration
build:
  target_bits: 64
  optimize: true
  memory: 64MB
```

### **Universal Conan Package**

```python
# conanfile.py for ao-runtime
class AoRuntimeConan(ConanFile):
    name = "ao-runtime"
    version = "1.0.0"

    options = {
        "enable_lua": [True, False],
        "enable_python": [True, False],
        "enable_javascript": [True, False],
        "target_bits": [32, 64]
    }

    default_options = {
        "enable_lua": True,
        "enable_python": False,
        "enable_javascript": False,
        "target_bits": 64
    }

    def requirements(self):
        self.requires("aolibc/1.0.0")
        if self.options.enable_lua:
            self.requires("lua/5.3.4")
        if self.options.enable_python:
            self.requires("cpython/3.11.0")
        if self.options.enable_javascript:
            self.requires("quickjs/2021.03.27")
```

## 🎯 **Usage Examples**

### **Lua Process (Current)**

```lua
-- src/main.lua
local json = require "json"

function handle(msgJSON, aoJSON)
    local msg = json.decode(msgJSON)
    local env = json.decode(aoJSON)

    -- Your logic here
    return json.encode({ok = true, response = "Hello from Lua!"})
end
```

### **Python Process**

```python
# src/main.py
import json

def handle(msg_json, ao_json):
    msg = json.loads(msg_json)
    env = json.loads(ao_json)

    # Your logic here
    return json.dumps({"ok": True, "response": "Hello from Python!"})
```

### **JavaScript Process**

```javascript
// src/main.js
function handle(msgJSON, aoJSON) {
  const msg = JSON.parse(msgJSON)
  const env = JSON.parse(aoJSON)

  // Your logic here
  return JSON.stringify({ ok: true, response: 'Hello from JavaScript!' })
}
```

## 🚀 **Benefits of This Approach**

1. **Language Choice Freedom**: Pick the best language for your use case
2. **Consistent Interface**: Same `handle(msgJSON, aoJSON)` contract across all languages
3. **Polyglot Processes**: Mix languages in different modules
4. **Easy Migration**: Port existing code between languages
5. **Performance Optimization**: Use compiled languages (Rust, Go) for hot paths
6. **Ecosystem Access**: Leverage each language's package ecosystem

## 📦 **Implementation Strategy**

1. **Phase 1**: Extract current Lua system into generic interface
2. **Phase 2**: Add Python support (popular for data processing)
3. **Phase 3**: Add JavaScript support (familiar to web developers)
4. **Phase 4**: Add compiled language support (Rust, Go via WASM)
5. **Phase 5**: Add polyglot module system

This creates a **universal compute platform** where developers can choose their preferred language while maintaining the same security, determinism, and deployment model that ao provides!

```c
// ao_runtime.h - Generic runtime interface
typedef struct ao_runtime ao_runtime_t;

typedef struct {
    const char* name;           // "lua", "python", "javascript", "wasm"

    // Lifecycle
    ao_runtime_t* (*create)(void);
    int (*destroy)(ao_runtime_t* runtime);

    // Module loading
    int (*load_module)(ao_runtime_t* runtime, const char* name,
                      const char* source, size_t size);
    int (*load_binary)(ao_runtime_t* runtime, const char* name,
                      const void* data, size_t size);

    // Function calling
    char* (*call_function)(ao_runtime_t* runtime, const char* function_name,
                          const char* msg_json, const char* ao_json);

    // Memory management
    void (*free_result)(char* result);

} ao_language_vtable_t;

// Registry of supported languages
extern const ao_language_vtable_t* ao_get_language(const char* name);
```

```c
// ao_runtime.c
#include "ao_runtime.h"

// Global runtime instance
static ao_runtime_t* g_runtime = NULL;
static const ao_language_vtable_t* g_language = NULL;

// WASM exported function - language agnostic!
EMSCRIPTEN_KEEPALIVE
char* ao_handle(const char* msg_json, const char* ao_json) {
    if (!g_runtime || !g_language) {
        return strdup("{\"ok\":false,\"error\":\"Runtime not initialized\"}");
    }

    return g_language->call_function(g_runtime, "handle", msg_json, ao_json);
}

// Boot any language runtime
EMSCRIPTEN_KEEPALIVE
int ao_boot(const char* language_name, const char* entry_module,
           const char* source, size_t source_size) {

    // Get language implementation
    g_language = ao_get_language(language_name);
    if (!g_language) {
        printf("Unsupported language: %s\n", language_name);
        return 1;
    }

    // Create runtime
    g_runtime = g_language->create();
    if (!g_runtime) {
        printf("Failed to create %s runtime\n", language_name);
        return 1;
    }

    // Load entry module
    if (g_language->load_module(g_runtime, entry_module, source, source_size)) {
        printf("Failed to load entry module\n");
        g_language->destroy(g_runtime);
        return 1;
    }

    return 0;
}
```

```c
// ao_lua.c
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

typedef struct {
    lua_State* L;
} ao_lua_runtime_t;

static ao_runtime_t* lua_create(void) {
    ao_lua_runtime_t* runtime = malloc(sizeof(ao_lua_runtime_t));
    runtime->L = luaL_newstate();
    luaL_openlibs(runtime->L);
    return (ao_runtime_t*)runtime;
}

static char* lua_call_function(ao_runtime_t* rt, const char* func_name,
                              const char* msg_json, const char* ao_json) {
    ao_lua_runtime_t* runtime = (ao_lua_runtime_t*)rt;
    lua_State* L = runtime->L;

    // Get global function
    lua_getglobal(L, func_name);
    if (!lua_isfunction(L, -1)) {
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Push arguments
    lua_pushstring(L, msg_json);
    lua_pushstring(L, ao_json);

    // Call function
    if (lua_pcall(L, 2, 1, 0) != LUA_OK) {
        const char* error = lua_tostring(L, -1);
        char* result = malloc(strlen(error) + 100);
        sprintf(result, "{\"ok\":false,\"error\":\"%s\"}", error);
        lua_pop(L, 1);
        return result;
    }

    // Get result
    const char* result_json = lua_tostring(L, -1);
    char* result = strdup(result_json ? result_json : "null");
    lua_pop(L, 1);

    return result;
}

static const ao_language_vtable_t lua_vtable = {
    .name = "lua",
    .create = lua_create,
    .destroy = lua_destroy,
    .load_module = lua_load_module,
    .call_function = lua_call_function,
    .free_result = free
};
```

```c
// ao_python.c
#include <Python.h>

typedef struct {
    PyObject* globals;
    PyObject* handle_func;
} ao_python_runtime_t;

static ao_runtime_t* python_create(void) {
    if (!Py_IsInitialized()) {
        Py_Initialize();
    }

    ao_python_runtime_t* runtime = malloc(sizeof(ao_python_runtime_t));
    runtime->globals = PyDict_New();

    // Add json module
    PyRun_String("import json", Py_single_input, runtime->globals, NULL);

    return (ao_runtime_t*)runtime;
}

static char* python_call_function(ao_runtime_t* rt, const char* func_name,
                                 const char* msg_json, const char* ao_json) {
    ao_python_runtime_t* runtime = (ao_python_runtime_t*)rt;

    // Get function
    PyObject* func = PyDict_GetItemString(runtime->globals, func_name);
    if (!func || !PyCallable_Check(func)) {
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Create arguments
    PyObject* args = PyTuple_New(2);
    PyTuple_SetItem(args, 0, PyUnicode_FromString(msg_json));
    PyTuple_SetItem(args, 1, PyUnicode_FromString(ao_json));

    // Call function
    PyObject* result = PyObject_CallObject(func, args);
    Py_DECREF(args);

    if (!result) {
        PyErr_Print();
        return strdup("{\"ok\":false,\"error\":\"Python execution error\"}");
    }

    // Convert result to string
    const char* result_str = PyUnicode_AsUTF8(result);
    char* output = strdup(result_str ? result_str : "null");
    Py_DECREF(result);

    return output;
}

static const ao_language_vtable_t python_vtable = {
    .name = "python",
    .create = python_create,
    .destroy = python_destroy,
    .load_module = python_load_module,
    .call_function = python_call_function,
    .free_result = free
};
```

```c
// ao_javascript.c
#include "quickjs.h"

typedef struct {
    JSRuntime* rt;
    JSContext* ctx;
} ao_js_runtime_t;

static ao_runtime_t* js_create(void) {
    ao_js_runtime_t* runtime = malloc(sizeof(ao_js_runtime_t));
    runtime->rt = JS_NewRuntime();
    runtime->ctx = JS_NewContext(runtime->rt);

    // Add JSON support
    js_std_add_helpers(runtime->ctx, 0, NULL);

    return (ao_runtime_t*)runtime;
}

static char* js_call_function(ao_runtime_t* rt, const char* func_name,
                             const char* msg_json, const char* ao_json) {
    ao_js_runtime_t* runtime = (ao_js_runtime_t*)rt;
    JSContext* ctx = runtime->ctx;

    // Get global function
    JSValue global = JS_GetGlobalObject(ctx);
    JSValue func = JS_GetPropertyStr(ctx, global, func_name);

    if (!JS_IsFunction(ctx, func)) {
        JS_FreeValue(ctx, func);
        JS_FreeValue(ctx, global);
        return strdup("{\"ok\":false,\"error\":\"Function not found\"}");
    }

    // Create arguments
    JSValue args[2];
    args[0] = JS_NewString(ctx, msg_json);
    args[1] = JS_NewString(ctx, ao_json);

    // Call function
    JSValue result = JS_Call(ctx, func, global, 2, args);

    // Cleanup args
    JS_FreeValue(ctx, args[0]);
    JS_FreeValue(ctx, args[1]);
    JS_FreeValue(ctx, func);
    JS_FreeValue(ctx, global);

    if (JS_IsException(result)) {
        JS_FreeValue(ctx, result);
        return strdup("{\"ok\":false,\"error\":\"JavaScript execution error\"}");
    }

    // Convert to string
    const char* result_str = JS_ToCString(ctx, result);
    char* output = strdup(result_str ? result_str : "null");
    JS_FreeCString(ctx, result_str);
    JS_FreeValue(ctx, result);

    return output;
}

static const ao_language_vtable_t js_vtable = {
    .name = "javascript",
    .create = js_create,
    .destroy = js_destroy,
    .load_module = js_load_module,
    .call_function = js_call_function,
    .free_result = free
};
```

```c
// ao_registry.c
static const ao_language_vtable_t* languages[] = {
    &lua_vtable,
    &python_vtable,
    &js_vtable,
    NULL
};

const ao_language_vtable_t* ao_get_language(const char* name) {
    for (int i = 0; languages[i]; i++) {
        if (strcmp(languages[i]->name, name) == 0) {
            return languages[i];
        }
    }
    return NULL;
}
```

```yaml
# ao-process.yml
language: lua # or python, javascript, etc.
entry_file: src/main.lua # language-specific entry point
output_file: process.wasm

# Language-specific dependencies
dependencies:
  lua:
    - json
    - lpeg
  python:
    - requests
    - numpy
  javascript:
    - lodash

# Functions to export (same interface for all languages)
functions:
  handle:
    args: [string, string]
    return: string

# Build configuration
build:
  target_bits: 64
  optimize: true
  memory: 64MB
```

```python
# conanfile.py for ao-runtime
class AoRuntimeConan(ConanFile):
    name = "ao-runtime"
    version = "1.0.0"

    options = {
        "enable_lua": [True, False],
        "enable_python": [True, False],
        "enable_javascript": [True, False],
        "target_bits": [32, 64]
    }

    default_options = {
        "enable_lua": True,
        "enable_python": False,
        "enable_javascript": False,
        "target_bits": 64
    }

    def requirements(self):
        self.requires("aolibc/1.0.0")
        if self.options.enable_lua:
            self.requires("lua/5.3.4")
        if self.options.enable_python:
            self.requires("cpython/3.11.0")
        if self.options.enable_javascript:
            self.requires("quickjs/2021.03.27")
```

```lua
-- src/main.lua
local json = require "json"

function handle(msgJSON, aoJSON)
    local msg = json.decode(msgJSON)
    local env = json.decode(aoJSON)

    -- Your logic here
    return json.encode({ok = true, response = "Hello from Lua!"})
end
```

```python
# src/main.py
import json

def handle(msg_json, ao_json):
    msg = json.loads(msg_json)
    env = json.loads(ao_json)

    # Your logic here
    return json.dumps({"ok": True, "response": "Hello from Python!"})
```

```javascript
// src/main.js
function handle(msgJSON, aoJSON) {
  const msg = JSON.parse(msgJSON)
  const env = JSON.parse(aoJSON)

  // Your logic here
  return JSON.stringify({ ok: true, response: 'Hello from JavaScript!' })
}
```

---

**User**

So I think the way i would package this is have the ao-interpreter C module, then create projects with the specified interpreted matching the header files for the types, and i can use the lua, python, quickjs, etc interpreters to power it. That way the interpreter module is fairly agnostic, it will have the appropriate aolibc in it, and then the interpreter is responsible for all the language bindings.

---

**Cursor**

Excellent architecture! That's a much cleaner separation of concerns. You're essentially creating a **plugin-based interpreter system** where the core ao-interpreter provides the runtime infrastructure, and language-specific interpreters plug into it. This is very similar to how web browsers handle JavaScript engines or how editors handle language servers.

Let me design this modular architecture:

## 🏗️ **Modular Architecture Design**

### **Core Package: `ao-interpreter`**

The foundational runtime that provides:

- WASM bootstrap and lifecycle management
- Memory management and sandboxing
- Message passing infrastructure
- Plugin interface for language interpreters

### **Language Packages: `ao-lua`, `ao-python`, `ao-quickjs`, etc.**

Language-specific interpreters that implement the plugin interface

## 📦 **Package Structure**

### **ao-interpreter (Core Runtime)**

```
ao-interpreter/
├── conanfile.py
├── include/
│   ├── ao_interpreter.h      # Main API
│   ├── ao_plugin.h          # Plugin interface
│   └── ao_types.h           # Common types
├── src/
│   ├── ao_runtime.c         # Core runtime
│   ├── ao_memory.c          # Memory management
│   ├── ao_messaging.c       # Message handling
│   └── ao_plugin_loader.c   # Plugin system
└── cmake/
    └── ao-interpreter-config.cmake
```

### **ao-lua (Lua Plugin)**

```
ao-lua/
├── conanfile.py
├── include/
│   └── ao_lua_plugin.h      # Lua-specific API
├── src/
│   ├── ao_lua_interpreter.c # Lua plugin implementation
│   ├── ao_lua_bindings.c    # ao-specific Lua bindings
│   └── lua_modules/         # Bundled Lua modules
│       ├── json.lua
│       └── ao.lua
└── cmake/
    └── ao-lua-config.cmake
```

## 🔌 **Plugin Interface Definition**

```c
// ao_plugin.h
#ifndef AO_PLUGIN_H
#define AO_PLUGIN_H

#include "ao_types.h"

// Plugin interface version for compatibility checking
#define AO_PLUGIN_API_VERSION 1

// Forward declarations
typedef struct ao_interpreter ao_interpreter_t;
typedef struct ao_plugin ao_plugin_t;
typedef struct ao_message ao_message_t;
typedef struct ao_result ao_result_t;

// Plugin lifecycle callbacks
typedef struct {
    // Plugin metadata
    const char* name;           // "lua", "python", etc.
    const char* version;        // "5.3.4", "3.11", etc.
    int api_version;            // AO_PLUGIN_API_VERSION

    // Lifecycle management
    int (*initialize)(ao_interpreter_t* interpreter, const ao_config_t* config);
    int (*shutdown)(ao_interpreter_t* interpreter);

    // Module management
    int (*load_module)(ao_interpreter_t* interpreter, const char* name,
                      const void* data, size_t size, ao_module_type_t type);
    int (*unload_module)(ao_interpreter_t* interpreter, const char* name);

    // Function execution
    ao_result_t* (*call_function)(ao_interpreter_t* interpreter,
                                 const char* function_name,
                                 const ao_message_t* message);

    // Memory management
    void (*free_result)(ao_result_t* result);

    // Optional: Hot reloading support
    int (*reload_module)(ao_interpreter_t* interpreter, const char* name,
                        const void* data, size_t size);

} ao_plugin_vtable_t;

// Plugin registration function (exported by each plugin)
extern const ao_plugin_vtable_t* ao_plugin_get_vtable(void);

#endif // AO_PLUGIN_H
```

## 🎯 **Core Types Definition**

```c
// ao_types.h
#ifndef AO_TYPES_H
#define AO_TYPES_H

#include <stddef.h>
#include <stdint.h>

// Configuration for interpreter
typedef struct {
    uint32_t memory_limit;      // Memory limit in bytes
    uint32_t execution_timeout; // Timeout in milliseconds
    uint8_t target_bits;        // 32 or 64 bit
    const char* temp_dir;       // Temporary directory for modules
} ao_config_t;

// Message structure (JSON-based communication)
typedef struct {
    char* json_data;            // Raw JSON string
    size_t size;                // Size of JSON data

    // Parsed fields (optional, for performance)
    const char* from;
    const char* target;
    const char* action;
    const void* data;
    size_t data_size;
} ao_message_t;

// Result from function execution
typedef struct {
    int success;                // 0 = success, non-zero = error
    char* json_response;        // JSON response string
    size_t response_size;       // Size of response

    // Error information (if success == 0)
    int error_code;
    char* error_message;
    char* stack_trace;
} ao_result_t;

// Module types
typedef enum {
    AO_MODULE_SOURCE,           // Source code (Lua, Python, JS)
    AO_MODULE_BYTECODE,         // Compiled bytecode
    AO_MODULE_BINARY,           // Native binary (.so, .dll)
    AO_MODULE_WASM              // WebAssembly module
} ao_module_type_t;

// Core interpreter state (opaque to plugins)
typedef struct ao_interpreter {
    const ao_plugin_vtable_t* plugin;
    void* plugin_state;         // Plugin-specific state
    ao_config_t config;

    // Core runtime state (managed by ao-interpreter)
    void* memory_pool;
    void* message_queue;
    void* module_registry;
} ao_interpreter_t;

#endif // AO_TYPES_H
```

## 🔧 **Core Runtime Implementation**

```c
// ao_runtime.c
#include "ao_interpreter.h"
#include "ao_plugin.h"
#include <emscripten.h>

static ao_interpreter_t* g_interpreter = NULL;

// WASM exported functions
EMSCRIPTEN_KEEPALIVE
int ao_initialize(const char* plugin_name, const char* config_json) {
    // Load plugin dynamically
    const ao_plugin_vtable_t* plugin = ao_load_plugin(plugin_name);
    if (!plugin) {
        printf("Failed to load plugin: %s\n", plugin_name);
        return 1;
    }

    // Validate plugin API version
    if (plugin->api_version != AO_PLUGIN_API_VERSION) {
        printf("Plugin API version mismatch: expected %d, got %d\n",
               AO_PLUGIN_API_VERSION, plugin->api_version);
        return 1;
    }

    // Create interpreter instance
    g_interpreter = malloc(sizeof(ao_interpreter_t));
    g_interpreter->plugin = plugin;

    // Parse configuration
    ao_config_t config;
    if (ao_parse_config(config_json, &config) != 0) {
        printf("Failed to parse configuration\n");
        free(g_interpreter);
        return 1;
    }
    g_interpreter->config = config;

    // Initialize core systems
    ao_memory_init(g_interpreter);
    ao_messaging_init(g_interpreter);
    ao_module_registry_init(g_interpreter);

    // Initialize plugin
    if (plugin->initialize(g_interpreter, &config) != 0) {
        printf("Plugin initialization failed\n");
        ao_cleanup(g_interpreter);
        return 1;
    }

    return 0;
}

EMSCRIPTEN_KEEPALIVE
int ao_load_module(const char* name, const void* data, size_t size, int type) {
    if (!g_interpreter || !g_interpreter->plugin) {
        return 1;
    }

    return g_interpreter->plugin->load_module(g_interpreter, name, data, size, type);
}

EMSCRIPTEN_KEEPALIVE
char* ao_handle_message(const char* message_json) {
    if (!g_interpreter || !g_interpreter->plugin) {
        return strdup("{\"success\":false,\"error\":\"Interpreter not initialized\"}");
    }

    // Parse message
    ao_message_t message;
    if (ao_parse_message(message_json, &message) != 0) {
        return strdup("{\"success\":false,\"error\":\"Invalid message format\"}");
    }

    // Call plugin
    ao_result_t* result = g_interpreter->plugin->call_function(
        g_interpreter, "handle", &message);

    if (!result) {
        ao_free_message(&message);
        return strdup("{\"success\":false,\"error\":\"Function call failed\"}");
    }

    // Extract response
    char* response = strdup(result->json_response ? result->json_response : "null");

    // Cleanup
    g_interpreter->plugin->free_result(result);
    ao_free_message(&message);

    return response;
}
```

## 🐍 **Lua Plugin Implementation**

```c
// ao_lua_interpreter.c
#include "ao_plugin.h"
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

typedef struct {
    lua_State* L;
    int module_count;
} ao_lua_state_t;

static int lua_initialize(ao_interpreter_t* interpreter, const ao_config_t* config) {
    ao_lua_state_t* state = malloc(sizeof(ao_lua_state_t));
    state->L = luaL_newstate();
    state->module_count = 0;

    // Open standard libraries
    luaL_openlibs(state->L);

    // Install ao-specific bindings
    ao_lua_install_bindings(state->L, interpreter);

    // Set memory limits
    lua_setallocf(state->L, ao_lua_allocator, (void*)(uintptr_t)config->memory_limit);

    interpreter->plugin_state = state;
    return 0;
}

static int lua_load_module(ao_interpreter_t* interpreter, const char* name,
                          const void* data, size_t size, ao_module_type_t type) {
    ao_lua_state_t* state = (ao_lua_state_t*)interpreter->plugin_state;
    lua_State* L = state->L;

    if (type == AO_MODULE_SOURCE) {
        // Load Lua source code
        if (luaL_loadbuffer(L, (const char*)data, size, name) != LUA_OK) {
            return 1;
        }

        // Execute to define module
        if (lua_pcall(L, 0, 1, 0) != LUA_OK) {
            lua_pop(L, 1); // Remove error
            return 1;
        }

        // Store in package.loaded
        lua_getglobal(L, "package");
        lua_getfield(L, -1, "loaded");
        lua_pushvalue(L, -3); // Copy module result
        lua_setfield(L, -2, name);
        lua_pop(L, 3); // Clean stack

        state->module_count++;
        return 0;
    }

    return 1; // Unsupported module type
}

static ao_result_t* lua_call_function(ao_interpreter_t* interpreter,
                                     const char* function_name,
                                     const ao_message_t* message) {
    ao_lua_state_t* state = (ao_lua_state_t*)interpreter->plugin_state;
    lua_State* L = state->L;

    ao_result_t* result = malloc(sizeof(ao_result_t));
    result->success = 0;
    result->json_response = NULL;
    result->error_message = NULL;

    // Get global function
    lua_getglobal(L, function_name);
    if (!lua_isfunction(L, -1)) {
        lua_pop(L, 1);
        result->success = 1;
        result->error_message = strdup("Function not found");
        return result;
    }

    // Push message JSON as argument
    lua_pushstring(L, message->json_data);

    // Call function
    if (lua_pcall(L, 1, 1, 0) != LUA_OK) {
        const char* error = lua_tostring(L, -1);
        result->success = 1;
        result->error_message = strdup(error ? error : "Unknown error");
        lua_pop(L, 1);
        return result;
    }

    // Get result
    const char* response = lua_tostring(L, -1);
    result->json_response = strdup(response ? response : "null");
    lua_pop(L, 1);

    return result;
}

static void lua_free_result(ao_result_t* result) {
    if (result) {
        free(result->json_response);
        free(result->error_message);
        free(result);
    }
}

// Plugin vtable
static const ao_plugin_vtable_t lua_plugin = {
    .name = "lua",
    .version = "5.3.4",
    .api_version = AO_PLUGIN_API_VERSION,
    .initialize = lua_initialize,
    .shutdown = lua_shutdown,
    .load_module = lua_load_module,
    .call_function = lua_call_function,
    .free_result = lua_free_result
};

// Plugin entry point
const ao_plugin_vtable_t* ao_plugin_get_vtable(void) {
    return &lua_plugin;
}
```

## 📦 **Conan Package Definitions**

### **ao-interpreter (Core)**

```python
# conanfile.py
class AoInterpreterConan(ConanFile):
    name = "ao-interpreter"
    version = "1.0.0"

    settings = "os", "compiler", "build_type", "arch"
    options = {
        "target_bits": [32, 64],
        "enable_hot_reload": [True, False]
    }
    default_options = {
        "target_bits": 64,
        "enable_hot_reload": False
    }

    requires = "aolibc/1.0.0"
    build_requires = "emscripten/3.1.59"

    def package_info(self):
        self.cpp_info.libs = ["ao-interpreter"]
        self.cpp_info.includedirs = ["include"]
```

### **ao-lua (Plugin)**

```python
# conanfile.py
class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    requires = (
        "ao-interpreter/1.0.0",
        "lua/5.3.4"
    )

    def package_info(self):
        self.cpp_info.libs = ["ao-lua-plugin"]
        self.cpp_info.requires = ["ao-interpreter::ao-interpreter", "lua::lua"]
```

## 🎯 **Usage Example**

```cmake
# CMakeLists.txt for your ao process
find_package(ao-interpreter REQUIRED)
find_package(ao-lua REQUIRED)

# Create your process
add_executable(my_process)
target_sources(my_process PRIVATE
    src/main.c
    src/process.lua  # Will be embedded
)

# Link interpreter and language plugin
target_link_libraries(my_process
    ao-interpreter::ao-interpreter
    ao-lua::ao-lua-plugin
)

# Configure for WASM build
ao_configure_wasm_target(my_process
    LANGUAGE lua
    ENTRY_MODULE process
    TARGET_BITS 64
)
```

This architecture gives you:

1. **Clean separation**: Core runtime vs language implementations
2. **Plugin ecosystem**: Easy to add new languages
3. **Consistent interface**: Same API across all languages
4. **Independent versioning**: Update languages without changing core
5. **Minimal dependencies**: Only include what you need
6. **Hot swapping**: Could even switch languages at runtime!

The beauty is that your `ao-interpreter` becomes a universal compute platform, and language plugins become first-class packages that anyone can create and distribute via Conan.

```plaintext
ao-interpreter/
├── conanfile.py
├── include/
│   ├── ao_interpreter.h      # Main API
│   ├── ao_plugin.h          # Plugin interface
│   └── ao_types.h           # Common types
├── src/
│   ├── ao_runtime.c         # Core runtime
│   ├── ao_memory.c          # Memory management
│   ├── ao_messaging.c       # Message handling
│   └── ao_plugin_loader.c   # Plugin system
└── cmake/
    └── ao-interpreter-config.cmake
```

```plaintext
ao-lua/
├── conanfile.py
├── include/
│   └── ao_lua_plugin.h      # Lua-specific API
├── src/
│   ├── ao_lua_interpreter.c # Lua plugin implementation
│   ├── ao_lua_bindings.c    # ao-specific Lua bindings
│   └── lua_modules/         # Bundled Lua modules
│       ├── json.lua
│       └── ao.lua
└── cmake/
    └── ao-lua-config.cmake
```

```c
// ao_plugin.h
#ifndef AO_PLUGIN_H
#define AO_PLUGIN_H

#include "ao_types.h"

// Plugin interface version for compatibility checking
#define AO_PLUGIN_API_VERSION 1

// Forward declarations
typedef struct ao_interpreter ao_interpreter_t;
typedef struct ao_plugin ao_plugin_t;
typedef struct ao_message ao_message_t;
typedef struct ao_result ao_result_t;

// Plugin lifecycle callbacks
typedef struct {
    // Plugin metadata
    const char* name;           // "lua", "python", etc.
    const char* version;        // "5.3.4", "3.11", etc.
    int api_version;            // AO_PLUGIN_API_VERSION

    // Lifecycle management
    int (*initialize)(ao_interpreter_t* interpreter, const ao_config_t* config);
    int (*shutdown)(ao_interpreter_t* interpreter);

    // Module management
    int (*load_module)(ao_interpreter_t* interpreter, const char* name,
                      const void* data, size_t size, ao_module_type_t type);
    int (*unload_module)(ao_interpreter_t* interpreter, const char* name);

    // Function execution
    ao_result_t* (*call_function)(ao_interpreter_t* interpreter,
                                 const char* function_name,
                                 const ao_message_t* message);

    // Memory management
    void (*free_result)(ao_result_t* result);

    // Optional: Hot reloading support
    int (*reload_module)(ao_interpreter_t* interpreter, const char* name,
                        const void* data, size_t size);

} ao_plugin_vtable_t;

// Plugin registration function (exported by each plugin)
extern const ao_plugin_vtable_t* ao_plugin_get_vtable(void);

#endif // AO_PLUGIN_H
```

```c
// ao_types.h
#ifndef AO_TYPES_H
#define AO_TYPES_H

#include <stddef.h>
#include <stdint.h>

// Configuration for interpreter
typedef struct {
    uint32_t memory_limit;      // Memory limit in bytes
    uint32_t execution_timeout; // Timeout in milliseconds
    uint8_t target_bits;        // 32 or 64 bit
    const char* temp_dir;       // Temporary directory for modules
} ao_config_t;

// Message structure (JSON-based communication)
typedef struct {
    char* json_data;            // Raw JSON string
    size_t size;                // Size of JSON data

    // Parsed fields (optional, for performance)
    const char* from;
    const char* target;
    const char* action;
    const void* data;
    size_t data_size;
} ao_message_t;

// Result from function execution
typedef struct {
    int success;                // 0 = success, non-zero = error
    char* json_response;        // JSON response string
    size_t response_size;       // Size of response

    // Error information (if success == 0)
    int error_code;
    char* error_message;
    char* stack_trace;
} ao_result_t;

// Module types
typedef enum {
    AO_MODULE_SOURCE,           // Source code (Lua, Python, JS)
    AO_MODULE_BYTECODE,         // Compiled bytecode
    AO_MODULE_BINARY,           // Native binary (.so, .dll)
    AO_MODULE_WASM              // WebAssembly module
} ao_module_type_t;

// Core interpreter state (opaque to plugins)
typedef struct ao_interpreter {
    const ao_plugin_vtable_t* plugin;
    void* plugin_state;         // Plugin-specific state
    ao_config_t config;

    // Core runtime state (managed by ao-interpreter)
    void* memory_pool;
    void* message_queue;
    void* module_registry;
} ao_interpreter_t;

#endif // AO_TYPES_H
```

```c
// ao_runtime.c
#include "ao_interpreter.h"
#include "ao_plugin.h"
#include <emscripten.h>

static ao_interpreter_t* g_interpreter = NULL;

// WASM exported functions
EMSCRIPTEN_KEEPALIVE
int ao_initialize(const char* plugin_name, const char* config_json) {
    // Load plugin dynamically
    const ao_plugin_vtable_t* plugin = ao_load_plugin(plugin_name);
    if (!plugin) {
        printf("Failed to load plugin: %s\n", plugin_name);
        return 1;
    }

    // Validate plugin API version
    if (plugin->api_version != AO_PLUGIN_API_VERSION) {
        printf("Plugin API version mismatch: expected %d, got %d\n",
               AO_PLUGIN_API_VERSION, plugin->api_version);
        return 1;
    }

    // Create interpreter instance
    g_interpreter = malloc(sizeof(ao_interpreter_t));
    g_interpreter->plugin = plugin;

    // Parse configuration
    ao_config_t config;
    if (ao_parse_config(config_json, &config) != 0) {
        printf("Failed to parse configuration\n");
        free(g_interpreter);
        return 1;
    }
    g_interpreter->config = config;

    // Initialize core systems
    ao_memory_init(g_interpreter);
    ao_messaging_init(g_interpreter);
    ao_module_registry_init(g_interpreter);

    // Initialize plugin
    if (plugin->initialize(g_interpreter, &config) != 0) {
        printf("Plugin initialization failed\n");
        ao_cleanup(g_interpreter);
        return 1;
    }

    return 0;
}

EMSCRIPTEN_KEEPALIVE
int ao_load_module(const char* name, const void* data, size_t size, int type) {
    if (!g_interpreter || !g_interpreter->plugin) {
        return 1;
    }

    return g_interpreter->plugin->load_module(g_interpreter, name, data, size, type);
}

EMSCRIPTEN_KEEPALIVE
char* ao_handle_message(const char* message_json) {
    if (!g_interpreter || !g_interpreter->plugin) {
        return strdup("{\"success\":false,\"error\":\"Interpreter not initialized\"}");
    }

    // Parse message
    ao_message_t message;
    if (ao_parse_message(message_json, &message) != 0) {
        return strdup("{\"success\":false,\"error\":\"Invalid message format\"}");
    }

    // Call plugin
    ao_result_t* result = g_interpreter->plugin->call_function(
        g_interpreter, "handle", &message);

    if (!result) {
        ao_free_message(&message);
        return strdup("{\"success\":false,\"error\":\"Function call failed\"}");
    }

    // Extract response
    char* response = strdup(result->json_response ? result->json_response : "null");

    // Cleanup
    g_interpreter->plugin->free_result(result);
    ao_free_message(&message);

    return response;
}
```

```c
// ao_lua_interpreter.c
#include "ao_plugin.h"
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

typedef struct {
    lua_State* L;
    int module_count;
} ao_lua_state_t;

static int lua_initialize(ao_interpreter_t* interpreter, const ao_config_t* config) {
    ao_lua_state_t* state = malloc(sizeof(ao_lua_state_t));
    state->L = luaL_newstate();
    state->module_count = 0;

    // Open standard libraries
    luaL_openlibs(state->L);

    // Install ao-specific bindings
    ao_lua_install_bindings(state->L, interpreter);

    // Set memory limits
    lua_setallocf(state->L, ao_lua_allocator, (void*)(uintptr_t)config->memory_limit);

    interpreter->plugin_state = state;
    return 0;
}

static int lua_load_module(ao_interpreter_t* interpreter, const char* name,
                          const void* data, size_t size, ao_module_type_t type) {
    ao_lua_state_t* state = (ao_lua_state_t*)interpreter->plugin_state;
    lua_State* L = state->L;

    if (type == AO_MODULE_SOURCE) {
        // Load Lua source code
        if (luaL_loadbuffer(L, (const char*)data, size, name) != LUA_OK) {
            return 1;
        }

        // Execute to define module
        if (lua_pcall(L, 0, 1, 0) != LUA_OK) {
            lua_pop(L, 1); // Remove error
            return 1;
        }

        // Store in package.loaded
        lua_getglobal(L, "package");
        lua_getfield(L, -1, "loaded");
        lua_pushvalue(L, -3); // Copy module result
        lua_setfield(L, -2, name);
        lua_pop(L, 3); // Clean stack

        state->module_count++;
        return 0;
    }

    return 1; // Unsupported module type
}

static ao_result_t* lua_call_function(ao_interpreter_t* interpreter,
                                     const char* function_name,
                                     const ao_message_t* message) {
    ao_lua_state_t* state = (ao_lua_state_t*)interpreter->plugin_state;
    lua_State* L = state->L;

    ao_result_t* result = malloc(sizeof(ao_result_t));
    result->success = 0;
    result->json_response = NULL;
    result->error_message = NULL;

    // Get global function
    lua_getglobal(L, function_name);
    if (!lua_isfunction(L, -1)) {
        lua_pop(L, 1);
        result->success = 1;
        result->error_message = strdup("Function not found");
        return result;
    }

    // Push message JSON as argument
    lua_pushstring(L, message->json_data);

    // Call function
    if (lua_pcall(L, 1, 1, 0) != LUA_OK) {
        const char* error = lua_tostring(L, -1);
        result->success = 1;
        result->error_message = strdup(error ? error : "Unknown error");
        lua_pop(L, 1);
        return result;
    }

    // Get result
    const char* response = lua_tostring(L, -1);
    result->json_response = strdup(response ? response : "null");
    lua_pop(L, 1);

    return result;
}

static void lua_free_result(ao_result_t* result) {
    if (result) {
        free(result->json_response);
        free(result->error_message);
        free(result);
    }
}

// Plugin vtable
static const ao_plugin_vtable_t lua_plugin = {
    .name = "lua",
    .version = "5.3.4",
    .api_version = AO_PLUGIN_API_VERSION,
    .initialize = lua_initialize,
    .shutdown = lua_shutdown,
    .load_module = lua_load_module,
    .call_function = lua_call_function,
    .free_result = lua_free_result
};

// Plugin entry point
const ao_plugin_vtable_t* ao_plugin_get_vtable(void) {
    return &lua_plugin;
}
```

```python
# conanfile.py
class AoInterpreterConan(ConanFile):
    name = "ao-interpreter"
    version = "1.0.0"

    settings = "os", "compiler", "build_type", "arch"
    options = {
        "target_bits": [32, 64],
        "enable_hot_reload": [True, False]
    }
    default_options = {
        "target_bits": 64,
        "enable_hot_reload": False
    }

    requires = "aolibc/1.0.0"
    build_requires = "emscripten/3.1.59"

    def package_info(self):
        self.cpp_info.libs = ["ao-interpreter"]
        self.cpp_info.includedirs = ["include"]
```

```python
# conanfile.py
class AoLuaConan(ConanFile):
    name = "ao-lua"
    version = "5.3.4"

    requires = (
        "ao-interpreter/1.0.0",
        "lua/5.3.4"
    )

    def package_info(self):
        self.cpp_info.libs = ["ao-lua-plugin"]
        self.cpp_info.requires = ["ao-interpreter::ao-interpreter", "lua::lua"]
```

```cmake
# CMakeLists.txt for your ao process
find_package(ao-interpreter REQUIRED)
find_package(ao-lua REQUIRED)

# Create your process
add_executable(my_process)
target_sources(my_process PRIVATE
    src/main.c
    src/process.lua  # Will be embedded
)

# Link interpreter and language plugin
target_link_libraries(my_process
    ao-interpreter::ao-interpreter
    ao-lua::ao-lua-plugin
)

# Configure for WASM build
ao_configure_wasm_target(my_process
    LANGUAGE lua
    ENTRY_MODULE process
    TARGET_BITS 64
)
```
