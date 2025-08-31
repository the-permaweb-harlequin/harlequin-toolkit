use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;

// Import the `console.log` function from the `console` module
#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

// Define a macro for easier logging
macro_rules! console_log {
    ($($t:tt)*) => (log(&format_args!($($t)*).to_string()))
}

// AO Message structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AOMessage {
    #[serde(rename = "Id")]
    pub id: Option<String>,
    #[serde(rename = "From")]
    pub from: Option<String>,
    #[serde(rename = "Owner")]
    pub owner: Option<String>,
    #[serde(rename = "Target")]
    pub target: Option<String>,
    #[serde(rename = "Anchor")]
    pub anchor: Option<String>,
    #[serde(rename = "Data")]
    pub data: Option<String>,
    #[serde(rename = "Tags")]
    pub tags: Option<HashMap<String, String>>,
    #[serde(rename = "Timestamp")]
    pub timestamp: Option<String>,
    #[serde(rename = "Block-Height")]
    pub block_height: Option<String>,
    #[serde(rename = "Hash-Chain")]
    pub hash_chain: Option<String>,
}

// AO Response structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AOResponse {
    #[serde(rename = "Target")]
    pub target: String,
    #[serde(rename = "Action")]
    pub action: String,
    #[serde(rename = "Data")]
    pub data: String,
    #[serde(flatten)]
    pub extra_fields: HashMap<String, String>,
}

impl AOResponse {
    pub fn new(target: &str, action: &str, data: &str) -> Self {
        Self {
            target: target.to_string(),
            action: action.to_string(),
            data: data.to_string(),
            extra_fields: HashMap::new(),
        }
    }

    pub fn with_field(mut self, key: &str, value: &str) -> Self {
        self.extra_fields.insert(key.to_string(), value.to_string());
        self
    }

    pub fn error(target: &str, message: &str) -> Self {
        Self::new(target, "Error", message)
    }
}

// Global state management
static STATE: Mutex<HashMap<String, String>> = Mutex::new(HashMap::new());

// Process state management
pub struct ProcessState;

impl ProcessState {
    pub fn set(key: &str, value: &str) -> Result<(), String> {
        if key.is_empty() || key.len() > 64 {
            return Err("Key must be between 1 and 64 characters".to_string());
        }

        if value.len() > 1000 {
            return Err("Value must be less than 1000 characters".to_string());
        }

        let mut state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        state.insert(key.to_string(), value.to_string());
        Ok(())
    }

    pub fn get(key: &str) -> Result<Option<String>, String> {
        let state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        Ok(state.get(key).cloned())
    }

    pub fn list() -> Result<HashMap<String, String>, String> {
        let state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        Ok(state.clone())
    }

    pub fn remove(key: &str) -> Result<bool, String> {
        let mut state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        Ok(state.remove(key).is_some())
    }

    pub fn clear() -> Result<(), String> {
        let mut state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        state.clear();
        Ok(())
    }

    pub fn size() -> Result<usize, String> {
        let state = STATE.lock().map_err(|e| format!("State lock error: {}", e))?;
        Ok(state.len())
    }
}

// Message handlers
pub fn handle_info(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");
    let state_size = ProcessState::size()?;

    let data = format!(
        "Hello from AO Process (Rust)! State entries: {}",
        state_size
    );

    Ok(AOResponse::new(from, "Info-Response", &data))
}

pub fn handle_set(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");

    let key = msg.tags
        .as_ref()
        .and_then(|tags| tags.get("Key"))
        .ok_or("Key is required")?;

    let value = msg.data.as_deref().ok_or("Value is required")?;

    // Validate key format
    if !key.chars().all(|c| c.is_alphanumeric() || c == '_' || c == '-') {
        return Ok(AOResponse::error(from, "Invalid key format. Use alphanumeric characters, underscores, and hyphens only"));
    }

    ProcessState::set(key, value)?;

    let data = format!("Successfully set {} to {}", key, value);
    Ok(AOResponse::new(from, "Set-Response", &data))
}

pub fn handle_get(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");

    let key = msg.tags
        .as_ref()
        .and_then(|tags| tags.get("Key"))
        .ok_or("Key is required")?;

    let value = ProcessState::get(key)?
        .unwrap_or_else(|| "Not found".to_string());

    Ok(AOResponse::new(from, "Get-Response", &value)
        .with_field("Key", key))
}

pub fn handle_list(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");

    let state = ProcessState::list()?;
    let state_json = serde_json::to_string(&state)
        .map_err(|e| format!("JSON serialization error: {}", e))?;

    Ok(AOResponse::new(from, "List-Response", &state_json))
}

pub fn handle_remove(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");

    let key = msg.tags
        .as_ref()
        .and_then(|tags| tags.get("Key"))
        .ok_or("Key is required")?;

    let removed = ProcessState::remove(key)?;

    let data = if removed {
        format!("Successfully removed {}", key)
    } else {
        format!("Key {} not found", key)
    };

    Ok(AOResponse::new(from, "Remove-Response", &data))
}

pub fn handle_clear(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");

    ProcessState::clear()?;

    Ok(AOResponse::new(from, "Clear-Response", "State cleared successfully"))
}

// Main message handler
pub fn handle_message(msg: &AOMessage) -> Result<AOResponse, String> {
    console_log!("Received message: {:?}", msg);

    let action = msg.tags
        .as_ref()
        .and_then(|tags| tags.get("Action"))
        .ok_or("Action is required")?;

    let response = match action.as_str() {
        "Info" => handle_info(msg)?,
        "Set" => handle_set(msg)?,
        "Get" => handle_get(msg)?,
        "List" => handle_list(msg)?,
        "Remove" => handle_remove(msg)?,
        "Clear" => handle_clear(msg)?,
        _ => {
            let from = msg.from.as_deref().unwrap_or("unknown");
            AOResponse::error(
                from,
                &format!(
                    "Unknown action: {}. Available actions: Info, Set, Get, List, Remove, Clear",
                    action
                )
            )
        }
    };

    console_log!("Sending response: {:?}", response);
    Ok(response)
}

// WebAssembly exports
#[wasm_bindgen]
pub fn init_process() {
    console_log::init_with_level(log::Level::Info).unwrap();
    console_log!("AO Process (Rust) initialized");
}

#[wasm_bindgen]
pub fn handle(message_json: &str) -> String {
    let msg: AOMessage = match serde_json::from_str(message_json) {
        Ok(msg) => msg,
        Err(e) => {
            let error_response = AOResponse::error("unknown", &format!("JSON parse error: {}", e));
            return serde_json::to_string(&error_response).unwrap_or_else(|_| {
                r#"{"Target":"unknown","Action":"Error","Data":"Critical JSON error"}"#.to_string()
            });
        }
    };

    let response = match handle_message(&msg) {
        Ok(response) => response,
        Err(e) => {
            let from = msg.from.as_deref().unwrap_or("unknown");
            AOResponse::error(from, &e)
        }
    };

    serde_json::to_string(&response).unwrap_or_else(|_| {
        r#"{"Target":"unknown","Action":"Error","Data":"Response serialization error"}"#.to_string()
    })
}

#[wasm_bindgen]
pub fn get_state() -> String {
    match ProcessState::list() {
        Ok(state) => serde_json::to_string(&state).unwrap_or_else(|_| "{}".to_string()),
        Err(e) => {
            console_log!("Error getting state: {}", e);
            "{}".to_string()
        }
    }
}

#[wasm_bindgen]
pub fn clear_state() -> bool {
    match ProcessState::clear() {
        Ok(()) => {
            console_log!("State cleared");
            true
        }
        Err(e) => {
            console_log!("Error clearing state: {}", e);
            false
        }
    }
}

// Utility functions for testing
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_process_state() {
        ProcessState::clear().unwrap();

        // Test set and get
        ProcessState::set("test_key", "test_value").unwrap();
        assert_eq!(ProcessState::get("test_key").unwrap(), Some("test_value".to_string()));

        // Test list
        let state = ProcessState::list().unwrap();
        assert_eq!(state.get("test_key"), Some(&"test_value".to_string()));

        // Test remove
        assert!(ProcessState::remove("test_key").unwrap());
        assert_eq!(ProcessState::get("test_key").unwrap(), None);
    }

    #[test]
    fn test_handle_info() {
        let msg = AOMessage {
            from: Some("test-sender".to_string()),
            tags: Some([("Action".to_string(), "Info".to_string())].iter().cloned().collect()),
            ..Default::default()
        };

        let response = handle_info(&msg).unwrap();
        assert_eq!(response.action, "Info-Response");
        assert!(response.data.contains("Hello from AO Process (Rust)"));
    }
}

// Default implementation for AOMessage
impl Default for AOMessage {
    fn default() -> Self {
        Self {
            id: None,
            from: None,
            owner: None,
            target: None,
            anchor: None,
            data: None,
            tags: None,
            timestamp: None,
            block_height: None,
            hash_chain: None,
        }
    }
}

