# Lua Utils Commands

The lua-utils commands provide utilities for working with Lua files, including bundling multiple files into a single executable.

## Interactive Lua Utils (Recommended)

For the best development experience, use the interactive TUI:

```bash
harlequin
```

Then navigate to **Lua Utils** from the welcome screen. This launches the interactive interface where you can:

1. **Select Command** - Choose "Lua Utils" from the main menu
2. **Choose Utility** - Select "Bundle" (more utilities coming soon)
3. **Select Entrypoint** - Pick your main Lua file (auto-discovery or manual selection)
4. **Configure Output** - Set the output path for the bundled file
5. **Monitor Progress** - Watch real-time bundling progress
6. **View Results** - See success confirmation with output file location

## Bundle Command

The bundle command combines multiple Lua files into a single executable by resolving `require()` statements and creating a self-contained script.

### Non-Interactive Bundle

For automation, CI/CD, or when you know exactly what you want to bundle:

#### Syntax

```bash
harlequin lua-utils bundle --entrypoint <file> [flags]
```

#### Required Flags

- `--entrypoint <file>` - Path to the main Lua file to bundle

#### Optional Flags

- `--outputPath <file>` - Path to output the bundled file (default: `<entrypoint>.bundled.lua`)
- `-d, --debug` - Enable debug logging for detailed output
- `-h, --help` - Show help message

### Examples

#### Basic Bundle

```bash
harlequin lua-utils bundle --entrypoint main.lua
```

This creates `main.bundled.lua` in the same directory as `main.lua`.

#### Bundle with Custom Output

```bash
harlequin lua-utils bundle --entrypoint src/app.lua --outputPath dist/bundle.lua
```

#### Bundle with Debug Output

```bash
harlequin lua-utils bundle --entrypoint main.lua --debug
```

#### Complete Example

```bash
harlequin lua-utils bundle --entrypoint src/main.lua --outputPath build/app.bundled.lua --debug
```

## How Bundling Works

The bundling process:

1. **Analyzes Entry Point** - Scans your main Lua file for `require()` statements
2. **Resolves Dependencies** - Recursively finds all required modules
3. **Handles Circular Dependencies** - Gracefully manages circular imports
4. **Creates Module Functions** - Wraps each module in a local function
5. **Generates Package Mappings** - Creates `package.loaded` mappings for `require()` compatibility
6. **Combines Content** - Merges all modules and main file into a single script

### Input Structure

```
project/
├── main.lua          # Entry point
├── utils/
│   ├── helper.lua     # Required by main.lua
│   └── math.lua       # Required by helper.lua
└── config.lua         # Required by main.lua
```

### Example main.lua

```lua
local utils = require("utils.helper")
local config = require("config")

print("App starting...")
print(utils.getMessage())
```

### Example utils/helper.lua

```lua
local math = require("utils.math")

local function getMessage()
    return "Hello from helper! Math result: " .. math.add(2, 3)
end

return {
    getMessage = getMessage
}
```

### Bundle Output

The bundled file will contain:

```lua
-- module: "utils.math"
local function _loaded_mod_utils_math()
-- ... module content ...
end

_G.package.loaded["utils.math"] = _loaded_mod_utils_math()

-- module: "utils.helper"
local function _loaded_mod_utils_helper()
-- ... module content with require("utils.math") working ...
end

_G.package.loaded["utils.helper"] = _loaded_mod_utils_helper()

-- module: "config"
local function _loaded_mod_config()
-- ... module content ...
end

_G.package.loaded["config"] = _loaded_mod_config()

-- Main file content
local utils = require("utils.helper")
local config = require("config")

print("App starting...")
print(utils.getMessage())
```

## Supported Patterns

### Require Statements

The bundler recognizes these `require()` patterns:

```lua
-- Direct string
local mod = require("module")

-- Dot notation for nested modules
local helper = require("utils.helper")

-- Variables and expressions are NOT supported
local name = "module"
local mod = require(name)  -- ❌ Not supported
```

### Module Return Patterns

Your modules can use any valid Lua return pattern:

```lua
-- Table export
return {
    func1 = function() end,
    value = 42
}

-- Function export
return function()
    -- module code
end

-- Mixed export
local module = {}
module.func = function() end
return module
```

### Circular Dependencies

The bundler handles circular dependencies gracefully:

```lua
-- a.lua
local b = require("b")
return { from_a = "hello" }

-- b.lua
local a = require("a")  -- Circular reference
return { from_b = "world" }
```

Both modules will be available, though you should avoid circular dependencies in your logic where possible.

## Advanced Usage

### Directory Structure

The bundler works with any directory structure:

```
src/
├── main.lua
├── lib/
│   ├── core/
│   │   ├── init.lua
│   │   └── utils.lua
│   └── helpers.lua
└── config/
    └── settings.lua
```

### Require Mapping

Module paths are mapped to file paths:

- `require("lib.helpers")` → `src/lib/helpers.lua`
- `require("config.settings")` → `src/config/settings.lua`
- `require("lib.core.utils")` → `src/lib/core/utils.lua`

### Build Integration

Use bundled Lua files in your build process:

```bash
# Bundle first
harlequin lua-utils bundle --entrypoint src/main.lua --outputPath dist/app.lua

# Then build with bundled file
harlequin build --entrypoint dist/app.lua --outputDir build
```

## Debug Mode

When using `--debug`, you'll see detailed logging including:

- Dependency tree analysis
- File resolution details
- Module wrapping process
- Circular dependency detection
- Output file creation

```bash
harlequin lua-utils bundle --entrypoint main.lua --debug
```

## Common Patterns

### Development Workflow

Use interactive mode during development:

```bash
harlequin  # Navigate to Lua Utils → Bundle
```

### CI/CD Pipeline

Use non-interactive mode in automated environments:

```bash
harlequin lua-utils bundle --entrypoint src/main.lua --outputPath dist/bundle.lua
```

### Library Distribution

Create distributable single-file versions of your Lua libraries:

```bash
harlequin lua-utils bundle --entrypoint lib/init.lua --outputPath releases/mylib-v1.0.lua
```

## Troubleshooting

### Common Issues

**File Not Found**

```
Error: entrypoint file does not exist: main.lua
```

Check that the entrypoint file path is correct and the file exists.

**Require Not Found**

```
Error: failed to read file utils/helper.lua: no such file or directory
```

Ensure all required modules exist as `.lua` files in the expected locations.

**Permission Denied**

```
Error: failed to write bundled file: permission denied
```

Check that you have write permissions to the output directory.

### Best Practices

1. **Use Relative Paths** - Keep your modules in predictable locations relative to your main file
2. **Avoid Dynamic Requires** - Use string literals in `require()` statements
3. **Test Bundled Output** - Always test your bundled files to ensure they work correctly
4. **Version Control** - Consider whether to commit bundled files or generate them during build
5. **Documentation** - Document which files are entry points vs. bundled outputs

## Future Enhancements

Planned lua-utils commands:

- **Format** - Format Lua code according to style guidelines
- **Lint** - Check Lua code for common issues and best practices
- **Minify** - Remove whitespace and comments for smaller file sizes
- **Analyze** - Generate dependency graphs and complexity reports
