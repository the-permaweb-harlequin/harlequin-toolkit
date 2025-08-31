const fs = require('fs');
const path = require('path');
const loader = require('@assemblyscript/loader');

// Test runner for AssemblyScript AO Process
async function runTests() {
  console.log('🎭 AO Process (AssemblyScript) - Test Suite');
  console.log('============================================');

  try {
    // Load the WebAssembly module
    const wasmPath = path.join(__dirname, '../build/{{PROJECT_NAME}}.debug.wasm');
    const wasmBuffer = fs.readFileSync(wasmPath);
    const wasmModule = await loader.instantiate(wasmBuffer, {
      // Import functions that the WASM module might need
      env: {
        trace: (ptr, len) => {
          const memory = wasmModule.exports.memory;
          const message = loader.getString(memory.buffer, ptr, len);
          console.log(`[TRACE] ${message}`);
        }
      }
    });

    const {
      initProcess,
      handle,
      getState,
      clearState,
      getStateSize,
      setState,
      getStateValue
    } = wasmModule.exports;

    let testsPassed = 0;
    let testsTotal = 0;

    function test(name, testFn) {
      testsTotal++;
      try {
        console.log(`\n${testsTotal}. Testing ${name}:`);
        testFn();
        console.log(`✅ ${name} - PASSED`);
        testsPassed++;
      } catch (error) {
        console.log(`❌ ${name} - FAILED: ${error.message}`);
      }
    }

    function assertEquals(actual, expected, message = '') {
      if (actual !== expected) {
        throw new Error(`Expected "${expected}", got "${actual}". ${message}`);
      }
    }

    function assertContains(text, substring, message = '') {
      if (!text.includes(substring)) {
        throw new Error(`Expected "${text}" to contain "${substring}". ${message}`);
      }
    }

    function assertJSON(jsonStr, message = '') {
      try {
        return JSON.parse(jsonStr);
      } catch (error) {
        throw new Error(`Invalid JSON: ${jsonStr}. ${message}`);
      }
    }

    // Initialize the process
    initProcess();

    // Test 1: Info Action
    test('Info Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Tags: { Action: "Info" }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Info-Response");
      assertEquals(response.Target, "test-sender");
      assertContains(response.Data, "Hello from AO Process (AssemblyScript)");
    });

    // Test 2: Set Action
    test('Set Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Data: "test-value",
        Tags: {
          Action: "Set",
          Key: "test-key"
        }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Set-Response");
      assertContains(response.Data, "Successfully set test-key to test-value");
    });

    // Test 3: Get Action
    test('Get Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Tags: {
          Action: "Get",
          Key: "test-key"
        }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Get-Response");
      assertEquals(response.Data, "test-value");
      assertEquals(response.Key, "test-key");
    });

    // Test 4: List Action
    test('List Action', () => {
      // Add another key-value pair
      const setMessage = JSON.stringify({
        From: "test-sender",
        Data: "another-value",
        Tags: {
          Action: "Set",
          Key: "another-key"
        }
      });
      handle(loader.allocateString(wasmModule.exports, setMessage));

      const listMessage = JSON.stringify({
        From: "test-sender",
        Tags: { Action: "List" }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, listMessage)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "List-Response");
      const stateData = assertJSON(response.Data);
      assertEquals(stateData["test-key"], "test-value");
      assertEquals(stateData["another-key"], "another-value");
    });

    // Test 5: Remove Action
    test('Remove Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Tags: {
          Action: "Remove",
          Key: "test-key"
        }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Remove-Response");
      assertContains(response.Data, "Successfully removed test-key");
    });

    // Test 6: Clear Action
    test('Clear Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Tags: { Action: "Clear" }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Clear-Response");
      assertEquals(response.Data, "State cleared successfully");
      assertEquals(getStateSize(), 0);
    });

    // Test 7: Error Handling - Missing Action
    test('Error Handling - Missing Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Data: "some-data"
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Error");
      assertContains(response.Data, "Action is required");
    });

    // Test 8: Error Handling - Unknown Action
    test('Error Handling - Unknown Action', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Tags: { Action: "UnknownAction" }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Error");
      assertContains(response.Data, "Unknown action: UnknownAction");
    });

    // Test 9: Error Handling - Missing Key for Set
    test('Error Handling - Missing Key for Set', () => {
      const message = JSON.stringify({
        From: "test-sender",
        Data: "test-value",
        Tags: { Action: "Set" }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Error");
      assertContains(response.Data, "Key is required");
    });

    // Test 10: Direct State Operations
    test('Direct State Operations', () => {
      clearState();
      assertEquals(getStateSize(), 0);

      // Set values directly
      const success1 = setState(loader.allocateString(wasmModule.exports, "direct-key1"), loader.allocateString(wasmModule.exports, "direct-value1"));
      const success2 = setState(loader.allocateString(wasmModule.exports, "direct-key2"), loader.allocateString(wasmModule.exports, "direct-value2"));

      assertEquals(success1, 1); // true
      assertEquals(success2, 1); // true
      assertEquals(getStateSize(), 2);

      // Get values directly
      const value1 = loader.getString(wasmModule.exports.memory.buffer, getStateValue(loader.allocateString(wasmModule.exports, "direct-key1")));
      const value2 = loader.getString(wasmModule.exports.memory.buffer, getStateValue(loader.allocateString(wasmModule.exports, "direct-key2")));

      assertEquals(value1, "direct-value1");
      assertEquals(value2, "direct-value2");
    });

    // Test 11: Key Validation
    test('Key Validation', () => {
      const invalidKeyMessage = JSON.stringify({
        From: "test-sender",
        Data: "test-value",
        Tags: {
          Action: "Set",
          Key: "invalid key with spaces"
        }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, invalidKeyMessage)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Error");
      assertContains(response.Data, "Invalid key format");
    });

    // Test 12: Value Size Limits
    test('Value Size Limits', () => {
      const largeValue = "x".repeat(1001); // Exceeds 1000 character limit
      const message = JSON.stringify({
        From: "test-sender",
        Data: largeValue,
        Tags: {
          Action: "Set",
          Key: "test-key"
        }
      });

      const responseStr = loader.getString(wasmModule.exports.memory.buffer, handle(loader.allocateString(wasmModule.exports, message)));
      const response = assertJSON(responseStr);

      assertEquals(response.Action, "Error");
      assertContains(response.Data, "Invalid value or value too long");
    });

    // Test 13: Get State Function
    test('Get State Function', () => {
      clearState();
      setState(loader.allocateString(wasmModule.exports, "state-key1"), loader.allocateString(wasmModule.exports, "state-value1"));
      setState(loader.allocateString(wasmModule.exports, "state-key2"), loader.allocateString(wasmModule.exports, "state-value2"));

      const stateStr = loader.getString(wasmModule.exports.memory.buffer, getState());
      const state = assertJSON(stateStr);

      assertEquals(state["state-key1"], "state-value1");
      assertEquals(state["state-key2"], "state-value2");
    });

    // Summary
    console.log('\n' + '='.repeat(50));
    console.log(`Tests completed: ${testsPassed}/${testsTotal} passed`);

    if (testsPassed === testsTotal) {
      console.log('🎉 All tests passed!');
      process.exit(0);
    } else {
      console.log(`❌ ${testsTotal - testsPassed} tests failed`);
      process.exit(1);
    }

  } catch (error) {
    console.error('Failed to run tests:', error);
    process.exit(1);
  }
}

// Run the tests
runTests().catch(console.error);

