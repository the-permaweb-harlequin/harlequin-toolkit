package templates

// Lua template definitions

const luaReadmeTemplate = `# {{.ProjectName}}

An AO process built with Lua.

## Author

{{.AuthorName}}{{if .GitHubUser}} ([@{{.GitHubUser}}](https://github.com/{{.GitHubUser}})){{end}}

## Description

This is a Lua-based AO process with C trampoline for performance optimization.

## Project Structure

- ` + "`process.lua`" + ` - Main process logic
- ` + "`handlers.lua`" + ` - Message handlers
- ` + "`wasm/c/`" + ` - C trampoline implementation
- ` + "`test/`" + ` - Test files
- ` + "`docs/`" + ` - Documentation

## Development

### Building

The C trampoline can be built using:

` + "```bash" + `
cd wasm/c
make
` + "```" + `

### Testing

Run tests with:

` + "```bash" + `
# Add your test command here
` + "```" + `

## Deployment

Deploy your process to AO using the Harlequin CLI:

` + "```bash" + `
harlequin build
harlequin upload-module
` + "```" + `
`

const luaProcessTemplate = `-- {{.ProjectName}} Process
-- Author: {{.AuthorName}}

local handlers = require("handlers")

-- Main message handler
function handle(msg)
    if handlers[msg.Action] then
        return handlers[msg.Action](msg)
    else
        return {
            Output = "Unknown action: " .. (msg.Action or "nil")
        }
    end
end

-- Export for testing
return {
    handle = handle
}
`

const luaHandlersTemplate = `-- {{.ProjectName}} Handlers
-- Author: {{.AuthorName}}

local handlers = {}

-- Default ping handler
handlers.ping = function(msg)
    return {
        Output = "pong",
        Data = "Hello from {{.ProjectName}}!"
    }
end

-- Info handler
handlers.info = function(msg)
    return {
        Output = "info",
        Data = {
            name = "{{.ProjectName}}",
            author = "{{.AuthorName}}",
            version = "1.0.0"
        }
    }
end

-- Add more handlers here

return handlers
`

const luaPackageJsonTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "description": "An AO process built with Lua",
  "author": "{{.AuthorName}}{{if .GitHubUser}} <{{.GitHubUser}}@users.noreply.github.com>{{end}}",
  "scripts": {
    "build": "cd wasm/c && make",
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "keywords": ["ao", "process", "lua", "arweave"],
  "license": "MIT",
  "devDependencies": {},
  "dependencies": {}
}
`
