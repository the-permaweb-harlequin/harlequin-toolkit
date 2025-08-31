use {{PROJECT_NAME}}::{handle_message, AOMessage, AOResponse, ProcessState};
use serde_json;
use std::collections::HashMap;

#[test]
fn test_info_handler() {
    let msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "Info".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&msg).unwrap();
    assert_eq!(response.action, "Info-Response");
    assert_eq!(response.target, "test-sender");
    assert!(response.data.contains("Hello from AO Process (Rust)"));
}

#[test]
fn test_set_and_get_handlers() {
    // Clear state first
    ProcessState::clear().unwrap();

    // Test Set
    let set_msg = AOMessage {
        from: Some("test-sender".to_string()),
        data: Some("test-value".to_string()),
        tags: Some([
            ("Action".to_string(), "Set".to_string()),
            ("Key".to_string(), "test-key".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    let set_response = handle_message(&set_msg).unwrap();
    assert_eq!(set_response.action, "Set-Response");
    assert!(set_response.data.contains("Successfully set test-key to test-value"));

    // Test Get
    let get_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([
            ("Action".to_string(), "Get".to_string()),
            ("Key".to_string(), "test-key".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    let get_response = handle_message(&get_msg).unwrap();
    assert_eq!(get_response.action, "Get-Response");
    assert_eq!(get_response.data, "test-value");
    assert_eq!(get_response.extra_fields.get("Key"), Some(&"test-key".to_string()));
}

#[test]
fn test_list_handler() {
    // Clear state and add some test data
    ProcessState::clear().unwrap();
    ProcessState::set("key1", "value1").unwrap();
    ProcessState::set("key2", "value2").unwrap();

    let list_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "List".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&list_msg).unwrap();
    assert_eq!(response.action, "List-Response");

    let state: HashMap<String, String> = serde_json::from_str(&response.data).unwrap();
    assert_eq!(state.get("key1"), Some(&"value1".to_string()));
    assert_eq!(state.get("key2"), Some(&"value2".to_string()));
}

#[test]
fn test_remove_handler() {
    // Clear state and add test data
    ProcessState::clear().unwrap();
    ProcessState::set("test-key", "test-value").unwrap();

    let remove_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([
            ("Action".to_string(), "Remove".to_string()),
            ("Key".to_string(), "test-key".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&remove_msg).unwrap();
    assert_eq!(response.action, "Remove-Response");
    assert!(response.data.contains("Successfully removed test-key"));

    // Verify the key is actually removed
    assert_eq!(ProcessState::get("test-key").unwrap(), None);
}

#[test]
fn test_clear_handler() {
    // Add some test data
    ProcessState::set("key1", "value1").unwrap();
    ProcessState::set("key2", "value2").unwrap();

    let clear_msg = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "Clear".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&clear_msg).unwrap();
    assert_eq!(response.action, "Clear-Response");
    assert_eq!(response.data, "State cleared successfully");

    // Verify state is actually cleared
    assert_eq!(ProcessState::size().unwrap(), 0);
}

#[test]
fn test_error_handling() {
    // Test missing action
    let msg_no_action = AOMessage {
        from: Some("test-sender".to_string()),
        ..Default::default()
    };

    let response = handle_message(&msg_no_action).unwrap();
    assert_eq!(response.action, "Error");
    assert!(response.data.contains("Action is required"));

    // Test unknown action
    let msg_unknown_action = AOMessage {
        from: Some("test-sender".to_string()),
        tags: Some([("Action".to_string(), "UnknownAction".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&msg_unknown_action).unwrap();
    assert_eq!(response.action, "Error");
    assert!(response.data.contains("Unknown action: UnknownAction"));

    // Test missing key for Set action
    let msg_no_key = AOMessage {
        from: Some("test-sender".to_string()),
        data: Some("test-value".to_string()),
        tags: Some([("Action".to_string(), "Set".to_string())].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&msg_no_key).unwrap();
    assert_eq!(response.action, "Error");
    assert!(response.data.contains("Key is required"));
}

#[test]
fn test_key_validation() {
    let msg_invalid_key = AOMessage {
        from: Some("test-sender".to_string()),
        data: Some("test-value".to_string()),
        tags: Some([
            ("Action".to_string(), "Set".to_string()),
            ("Key".to_string(), "invalid key with spaces".to_string()),
        ].iter().cloned().collect()),
        ..Default::default()
    };

    let response = handle_message(&msg_invalid_key).unwrap();
    assert_eq!(response.action, "Error");
    assert!(response.data.contains("Invalid key format"));
}

#[test]
fn test_state_operations() {
    ProcessState::clear().unwrap();

    // Test basic operations
    assert_eq!(ProcessState::size().unwrap(), 0);

    ProcessState::set("test", "value").unwrap();
    assert_eq!(ProcessState::size().unwrap(), 1);
    assert_eq!(ProcessState::get("test").unwrap(), Some("value".to_string()));

    let state = ProcessState::list().unwrap();
    assert_eq!(state.len(), 1);
    assert_eq!(state.get("test"), Some(&"value".to_string()));

    assert!(ProcessState::remove("test").unwrap());
    assert!(!ProcessState::remove("nonexistent").unwrap());
    assert_eq!(ProcessState::size().unwrap(), 0);
}

#[test]
fn test_value_size_limits() {
    let large_value = "x".repeat(1001); // Exceeds 1000 character limit

    let result = ProcessState::set("test", &large_value);
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Value must be less than 1000 characters"));
}

#[test]
fn test_key_size_limits() {
    let large_key = "x".repeat(65); // Exceeds 64 character limit

    let result = ProcessState::set(&large_key, "value");
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Key must be between 1 and 64 characters"));

    // Test empty key
    let result = ProcessState::set("", "value");
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Key must be between 1 and 64 characters"));
}

#[test]
fn test_concurrent_operations() {
    use std::thread;
    use std::sync::Arc;
    use std::sync::atomic::{AtomicUsize, Ordering};

    ProcessState::clear().unwrap();
    let counter = Arc::new(AtomicUsize::new(0));
    let mut handles = vec![];

    // Spawn multiple threads to test concurrent access
    for i in 0..10 {
        let counter_clone = Arc::clone(&counter);
        let handle = thread::spawn(move || {
            let key = format!("key{}", i);
            let value = format!("value{}", i);

            ProcessState::set(&key, &value).unwrap();
            counter_clone.fetch_add(1, Ordering::SeqCst);
        });
        handles.push(handle);
    }

    // Wait for all threads to complete
    for handle in handles {
        handle.join().unwrap();
    }

    assert_eq!(counter.load(Ordering::SeqCst), 10);
    assert_eq!(ProcessState::size().unwrap(), 10);
}

