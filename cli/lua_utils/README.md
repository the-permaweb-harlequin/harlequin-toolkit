# Lua Utils

This package provides various utils for interacting with lua.

<!-- TODO: Busted test runner with golua -->
<!-- TODO: lcov runner with golua -->

### Basic Bundling of code (output a single lua file)

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
