package templates

// AssemblyScript template definitions

const asReadmeTemplate = `# {{.ProjectName}}

An AO process built with AssemblyScript.

## Author

{{.AuthorName}}{{if .GitHubUser}} ([@{{.GitHubUser}}](https://github.com/{{.GitHubUser}})){{end}}

## Description

This is an AssemblyScript-based AO process with TypeScript-like syntax compiled to WebAssembly.

## Project Structure

- ` + "`assembly/index.ts`" + ` - Main entry point
- ` + "`assembly/handlers.ts`" + ` - Message handlers
- ` + "`package.json`" + ` - NPM package configuration
- ` + "`asconfig.json`" + ` - AssemblyScript build configuration
- ` + "`test/`" + ` - Test files
- ` + "`docs/`" + ` - Documentation

## Development

### Prerequisites

- Node.js 16+
- npm or yarn
- AssemblyScript compiler

### Installation

` + "```bash" + `
npm install
` + "```" + `

### Building

` + "```bash" + `
# Build for WebAssembly
npm run asbuild

# Build optimized version
npm run asbuild:optimized
` + "```" + `

### Testing

` + "```bash" + `
npm test
` + "```" + `

## Deployment

Deploy your process to AO using the Harlequin CLI:

` + "```bash" + `
harlequin build
harlequin upload-module
` + "```" + `
`

const asPackageJsonTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "description": "An AO process built with AssemblyScript",
  "author": "{{.AuthorName}}{{if .GitHubUser}} <{{.GitHubUser}}@users.noreply.github.com>{{end}}",
  "scripts": {
    "asbuild:debug": "asc assembly/index.ts --target debug",
    "asbuild:release": "asc assembly/index.ts --target release",
    "asbuild": "npm run asbuild:debug && npm run asbuild:release",
    "asbuild:optimized": "asc assembly/index.ts --target release --optimize --converge",
    "test": "npm run asbuild && node test/index.js"
  },
  "keywords": ["ao", "process", "assemblyscript", "arweave", "wasm"],
  "license": "MIT",
  "devDependencies": {
    "assemblyscript": "^0.27.0"
  },
  "dependencies": {
    "@assemblyscript/loader": "^0.27.0"
  },
  "type": "module"
}
`

const asConfigTemplate = `{
  "targets": {
    "debug": {
      "outFile": "build/debug.wasm",
      "textFile": "build/debug.wat",
      "sourceMap": true,
      "debug": true
    },
    "release": {
      "outFile": "build/release.wasm",
      "textFile": "build/release.wat",
      "sourceMap": true,
      "optimizeLevel": 3,
      "shrinkLevel": 0,
      "converge": false,
      "noAssert": false
    }
  },
  "options": {
    "bindings": "esm"
  }
}
`

const asIndexTemplate = `// {{.ProjectName}} - AssemblyScript AO Process
// Author: {{.AuthorName}}

import { Handlers } from "./handlers";

// JSON parsing helpers (simplified for demo)
class JSONParser {
  static parseMessage(json: string): Map<string, string> {
    let map = new Map<string, string>();
    // Simple JSON parsing - in production, use a proper JSON library
    // This is a simplified example
    if (json.includes('"Action"')) {
      let start = json.indexOf('"Action"') + 9;
      let end = json.indexOf('"', start + 1);
      if (end > start) {
        let action = json.substring(start + 1, end);
        map.set("Action", action);
      }
    }
    return map;
  }

  static stringify(output: string, data: string): string {
    return '{"Output":"' + output + '","Data":' + data + '}';
  }
}

// Main handle function exported to WebAssembly
export function handle(messagePtr: i32, messageLen: i32): i32 {
  // Get message string from memory
  let messageBytes = new Uint8Array(messageLen);
  for (let i = 0; i < messageLen; i++) {
    messageBytes[i] = load<u8>(messagePtr + i);
  }

  let messageStr = String.UTF8.decode(messageBytes.buffer);
  let messageMap = JSONParser.parseMessage(messageStr);

  let action = messageMap.has("Action") ? messageMap.get("Action") : "";
  let response: string;

  if (action == "ping") {
    response = Handlers.handlePing(messageMap);
  } else if (action == "info") {
    response = Handlers.handleInfo(messageMap);
  } else {
    response = JSONParser.stringify("error", '"Unknown action"');
  }

  // Allocate memory for response and copy string
  let responseBytes = String.UTF8.encode(response);
  let responsePtr = heap.alloc(responseBytes.byteLength);
  memory.copy(responsePtr, changetype<usize>(responseBytes), responseBytes.byteLength);

  return responsePtr;
}

// Helper function for testing
export function handleString(message: string): string {
  let messageMap = JSONParser.parseMessage(message);
  let action = messageMap.has("Action") ? messageMap.get("Action") : "";

  if (action == "ping") {
    return Handlers.handlePing(messageMap);
  } else if (action == "info") {
    return Handlers.handleInfo(messageMap);
  } else {
    return JSONParser.stringify("error", '"Unknown action"');
  }
}
`

const asHandlersTemplate = `// {{.ProjectName}} Message Handlers
// Author: {{.AuthorName}}

export class Handlers {
  static handlePing(message: Map<string, string>): string {
    return '{"Output":"pong","Data":"Hello from {{.ProjectName}}!"}';
  }

  static handleInfo(message: Map<string, string>): string {
    return '{"Output":"info","Data":{"name":"{{.ProjectName}}","author":"{{.AuthorName}}","version":"1.0.0","language":"assemblyscript"}}';
  }

  // Add more handlers here
}
`

const asTsConfigTemplate = `{
  "extends": "assemblyscript/std/assembly.json",
  "include": [
    "./**/*.ts"
  ]
}
`

// Common gitignore template for all projects
const gitignoreTemplate = `# Dependencies
node_modules/
target/
build/
pkg/
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build artifacts
*.wasm
*.wat
*.so
*.dll
*.dylib

# Test outputs
coverage/
test-results/

# Environment
.env
.env.local
`
