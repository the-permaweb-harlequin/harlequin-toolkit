-- Luacheck configuration for AO Process
-- This file configures the Lua linter for the project

-- Standard library globals
std = "lua51+lua52+lua53+lua54"

-- Global variables allowed in AO environment
globals = {
    -- AO globals
    "ao",
    "Handlers",
    "State",
    "Inbox",
    "Outbox",

    -- Common Lua globals that might be available
    "require",
    "module",
    "package",

    -- Test globals (for busted)
    "describe",
    "it",
    "before_each",
    "after_each",
    "setup",
    "teardown",
    "assert",
    "spy",
    "stub",
    "mock",
    "pending",
    "finally"
}

-- Read-only globals (cannot be modified)
read_globals = {
    "ao",
    "Handlers",
    "require",
    "package"
}

-- Files and directories to exclude
exclude_files = {
    "dist/",
    "build/",
    "*.rock",
    ".luarocks/"
}

-- Maximum line length
max_line_length = 120

-- Maximum cyclomatic complexity
max_cyclomatic_complexity = 10

-- Warnings to ignore
ignore = {
    "212",  -- Unused argument
    "213",  -- Unused loop variable
    "631",  -- Line is too long (we set max_line_length instead)
}

-- Files with specific configurations
files = {
    ["test/*.lua"] = {
        -- Test files can use additional globals
        globals = {
            "describe", "it", "before_each", "after_each",
            "setup", "teardown", "assert", "spy", "stub",
            "mock", "pending", "finally"
        }
    },

    ["main.lua"] = {
        -- Main file can modify global state
        allow_defined_top = true
    }
}

