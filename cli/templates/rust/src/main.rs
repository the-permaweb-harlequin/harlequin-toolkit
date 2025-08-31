use {{PROJECT_NAME}}::{handle_message, AOMessage, ProcessState};
use serde_json;
use std::collections::HashMap;

fn main() {
    println!("🎭 AO Process (Rust) - Testing Mode");
    println!("=====================================");

    // Test the Info action
    println!("\n1. Testing Info action:");
    let info_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "Info".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&info_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test the Set action
    println!("\n2. Testing Set action:");
    let set_msg = AOMessage {
        from: Some("test-sender".to_string()),
        data: Some("test-value".to_string()),
        tags: Some([
            ("Action".to_string(), "Set".to_string()),
            ("Key".to_string(), "test-key".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&set_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test the Get action
    println!("\n3. Testing Get action:");
    let get_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([
            ("Action".to_string(), "Get".to_string()),
            ("Key".to_string(), "test-key".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&get_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test the List action
    println!("\n4. Testing List action:");
    let list_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "List".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&list_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test error handling
    println!("\n5. Testing error handling (unknown action):");
    let error_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "UnknownAction".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&error_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test state operations
    println!("\n6. Testing direct state operations:");

    // Set multiple values
    ProcessState::set("name", "Alice").unwrap();
    ProcessState::set("age", "30").unwrap();
    ProcessState::set("city", "New York").unwrap();

    println!("State size: {}", ProcessState::size().unwrap());

    let state = ProcessState::list().unwrap();
    println!("Current state: {}", serde_json::to_string_pretty(&state).unwrap());

    // Test Remove action
    println!("\n7. Testing Remove action:");
    let remove_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([
            ("Action".to_string(), "Remove".to_string()),
            ("Key".to_string(), "age".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&remove_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    // Test Clear action
    println!("\n8. Testing Clear action:");
    let clear_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "Clear".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    match handle_message(&clear_msg) {
        Ok(response) => {
            let json = serde_json::to_string_pretty(&response).unwrap();
            println!("Response: {}", json);
        }
        Err(e) => println!("Error: {}", e),
    }

    println!("\n✅ All tests completed!");
}

