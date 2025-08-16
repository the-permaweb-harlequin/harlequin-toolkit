# Lua Utils - Lua Bundler

This package provides a Go implementation of a Lua module bundler, ported from JavaScript. It resolves dependencies and creates a single executable Lua file from a project with multiple modules.

## Features

- **Dependency Resolution**: Automatically discovers and resolves `require()` statements
- **Circular Dependency Protection**: Prevents infinite loops in dependency graphs
- **Module Deduplication**: Handles modules imported with different names but same paths
- **Topological Sorting**: Ensures dependencies are loaded in the correct order

## Usage

### Basic Bundling

```go
package main

import (
    "fmt"
    "log"
    "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/lua_utils"
)

func main() {
    // Bundle a Lua project
    bundledCode, err := luautils.Bundle("./project/main.lua")
    if err != nil {
        log.Fatalf("Failed to bundle: %v", err)
    }
    
    fmt.Println(bundledCode)
}
```

### Project Structure

Given a project structure like:
```
project/
├── main.lua
├── utils/
│   └── helper.lua
└── config.lua
```

Where `main.lua` contains:
```lua
local helper = require("utils.helper")
local config = require("config")

print("Hello from main!")
print(helper.greet())
```

The bundler will:
1. Parse `main.lua` and find `require()` statements
2. Recursively resolve dependencies (`utils.helper` and `config`)
3. Create a single bundled file with all modules properly loaded

### Output Format

The bundled output contains:
1. Module wrapper functions for each dependency
2. `_G.package.loaded` mappings to simulate `require()`
3. The main file content at the end

Example output:
```lua
-- module: "utils.helper"
local function _loaded_mod_utils_helper()
-- helper.lua content here
end

_G.package.loaded["utils.helper"] = _loaded_mod_utils_helper()

-- module: "config" 
local function _loaded_mod_config()
-- config.lua content here
end

_G.package.loaded["config"] = _loaded_mod_config()

-- main.lua content here
local helper = require("utils.helper")
local config = require("config")
print("Hello from main!")
```

## API Reference

### Types

```go
type Module struct {
    Name    string   // Module name (e.g., "utils.helper")
    Path    string   // File path
    Content *string  // File content (nil if not found)
}
```

### Functions

#### `Bundle(entryLuaPath string) (string, error)`
Bundles a Lua project starting from the entry file.

**Parameters:**
- `entryLuaPath`: Path to the main Lua file

**Returns:**
- Bundled Lua code as a string
- Error if bundling fails

#### `createProjectStructure(mainFile string) ([]Module, error)`
Internal function that builds the dependency graph using DFS.

#### `createExecutableFromProject(project []Module) (string, error)`
Internal function that converts the project structure into executable Lua code.

## How It Works

1. **Dependency Discovery**: Starting from the entry file, the bundler uses regex to find `require()` statements
2. **Depth-First Search**: Recursively explores each dependency to build a complete dependency graph
3. **Topological Sort**: Orders modules so dependencies are loaded before dependents
4. **Code Generation**: Wraps each module in a function and creates `package.loaded` mappings
5. **Bundling**: Combines all modules into a single executable file

## Supported Require Patterns

The bundler recognizes these `require()` patterns:
- `require("module")`
- `require('module')`
- `require("deeply.nested.module")`
- `require ( "module" )` (with spaces)

## Module Path Resolution

Module names are converted to file paths using these rules:
- Dots (`.`) become directory separators
- `.lua` extension is automatically added
- Paths are resolved relative to the entry file's directory

Examples:
- `require("config")` → `./config.lua`
- `require("utils.helper")` → `./utils/helper.lua`

## Error Handling

The bundler handles these error cases:
- Missing files (silently ignored - assumes external modules)
- Circular dependencies (prevented by cycle detection)
- Empty projects
- File read errors

## Testing

Run the test suite:
```bash
go test ./lua_utils -v
```

The tests cover:
- Module name transformation
- Dependency discovery
- Project structure creation
- Full bundling workflow
- Error cases
