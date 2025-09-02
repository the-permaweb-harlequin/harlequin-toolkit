package templates

// Rust template definitions

const rustReadmeTemplate = `# {{.ProjectName}}

An AO process built with Rust.

## Author

{{.AuthorName}}{{if .GitHubUser}} ([@{{.GitHubUser}}](https://github.com/{{.GitHubUser}})){{end}}

## Description

This is a Rust-based AO process compiled to WebAssembly for high performance and memory safety.

## Project Structure

- ` + "`src/lib.rs`" + ` - Main library entry point
- ` + "`src/handlers.rs`" + ` - Message handlers
- ` + "`Cargo.toml`" + ` - Rust package configuration
- ` + "`wasm-pack.json`" + ` - WebAssembly build configuration
- ` + "`test/`" + ` - Test files
- ` + "`docs/`" + ` - Documentation

## Development

### Prerequisites

- Rust (latest stable)
- wasm-pack
- Node.js (for testing)

### Building

` + "```bash" + `
# Build for WebAssembly
wasm-pack build --target web

# Build for testing
cargo build
` + "```" + `

### Testing

` + "```bash" + `
cargo test
` + "```" + `

## Deployment

Deploy your process to AO using the Harlequin CLI:

` + "```bash" + `
harlequin build
harlequin upload-module
` + "```" + `
`

const rustCargoTomlTemplate = `[package]
name = "{{.ProjectName}}"
version = "1.0.0"
edition = "2021"
authors = ["{{.AuthorName}}{{if .GitHubUser}} <{{.GitHubUser}}@users.noreply.github.com>{{end}}"]
description = "An AO process built with Rust"
license = "MIT"
keywords = ["ao", "process", "rust", "arweave", "wasm"]

[lib]
crate-type = ["cdylib", "rlib"]

[dependencies]
wasm-bindgen = "0.2"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

[dependencies.web-sys]
version = "0.3"
features = [
  "console",
]

[dependencies.wee_alloc]
version = "0.4.5"
optional = true

[features]
default = ["console_error_panic_hook"]
console_error_panic_hook = ["console_error_panic_hook"]

[profile.release]
opt-level = 3
lto = true
debug = false
panic = "abort"
codegen-units = 1
`

const rustLibTemplate = `// {{.ProjectName}} - Rust AO Process
// Author: {{.AuthorName}}

use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

mod handlers;
use handlers::*;

// Import the console.log function from the browser
#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

// Define a macro to print to the console
macro_rules! console_log {
    ($($t:tt)*) => (log(&format_args!($($t)*).to_string()))
}

// Message structure for AO
#[derive(Serialize, Deserialize)]
pub struct Message {
    #[serde(rename = "Action")]
    pub action: Option<String>,
    #[serde(rename = "Data")]
    pub data: Option<String>,
    #[serde(flatten)]
    pub other: serde_json::Map<String, serde_json::Value>,
}

// Response structure
#[derive(Serialize, Deserialize)]
pub struct Response {
    #[serde(rename = "Output")]
    pub output: String,
    #[serde(rename = "Data")]
    pub data: Option<serde_json::Value>,
}

// Main handle function exported to WebAssembly
#[wasm_bindgen]
pub fn handle(message_json: &str) -> String {
    console_error_panic_hook::set_once();

    let message: Message = match serde_json::from_str(message_json) {
        Ok(msg) => msg,
        Err(e) => {
            let error_response = Response {
                output: "error".to_string(),
                data: Some(serde_json::json!({
                    "message": format!("Failed to parse message: {}", e)
                })),
            };
            return serde_json::to_string(&error_response).unwrap_or_default();
        }
    };

    let response = match message.action.as_deref() {
        Some("ping") => handle_ping(&message),
        Some("info") => handle_info(&message),
        _ => Response {
            output: "error".to_string(),
            data: Some(serde_json::json!({
                "message": "Unknown action"
            })),
        },
    };

    serde_json::to_string(&response).unwrap_or_default()
}

// For testing without WebAssembly
pub fn handle_message(message: Message) -> Response {
    match message.action.as_deref() {
        Some("ping") => handle_ping(&message),
        Some("info") => handle_info(&message),
        _ => Response {
            output: "error".to_string(),
            data: Some(serde_json::json!({
                "message": "Unknown action"
            })),
        },
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ping() {
        let message = Message {
            action: Some("ping".to_string()),
            data: None,
            other: serde_json::Map::new(),
        };
        let response = handle_message(message);
        assert_eq!(response.output, "pong");
    }

    #[test]
    fn test_info() {
        let message = Message {
            action: Some("info".to_string()),
            data: None,
            other: serde_json::Map::new(),
        };
        let response = handle_message(message);
        assert_eq!(response.output, "info");
    }
}
`

const rustHandlersTemplate = `// {{.ProjectName}} Message Handlers
// Author: {{.AuthorName}}

use crate::{Message, Response};
use serde_json::json;

pub fn handle_ping(_message: &Message) -> Response {
    Response {
        output: "pong".to_string(),
        data: Some(json!("Hello from {{.ProjectName}}!")),
    }
}

pub fn handle_info(_message: &Message) -> Response {
    Response {
        output: "info".to_string(),
        data: Some(json!({
            "name": "{{.ProjectName}}",
            "author": "{{.AuthorName}}",
            "version": "1.0.0",
            "language": "rust"
        })),
    }
}

// Add more handlers here
`

const rustWasmPackTemplate = `{
  "out-dir": "pkg",
  "out-name": "{{.ProjectName}}",
  "target": "web"
}
`
