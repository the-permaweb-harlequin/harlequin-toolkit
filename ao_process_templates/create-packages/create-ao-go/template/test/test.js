import { readFileSync } from 'fs';
import { createAOSLoader } from '@permaweb/ao-loader';

async function test() {
  console.log('🧪 Testing AO Go Process...\n');

  try {
    // Load the compiled WASM
    const wasmBinary = readFileSync('./build/process.wasm');
    const loader = createAOSLoader();

    // Create test environment
    const environment = {
      Process: {
        Id: 'test-process-id',
        Owner: 'test-owner',
        Tags: []
      },
      Module: 'test-module'
    };

    // Test Hello action
    console.log('📨 Testing Hello action...');
    const helloMessage = {
      Target: 'test-process',
      Action: 'Hello',
      Data: '',
      Anchor: '0',
      Tags: [
        { name: 'Action', value: 'Hello' }
      ]
    };

    const helloResult = await loader(wasmBinary, helloMessage, environment);
    console.log('📋 Hello Response:', JSON.stringify(helloResult, null, 2));

    if (helloResult.Output === 'Hello, AO World from Go!') {
      console.log('✅ Test passed: Hello action works correctly');
    } else {
      console.log('❌ Test failed: Unexpected hello output');
      process.exit(1);
    }

    // Test Echo action
    console.log('\n📨 Testing Echo action...');
    const echoMessage = {
      Target: 'test-process',
      Action: 'Echo',
      Data: 'Hello from Go test!',
      Anchor: '0',
      Tags: [
        { name: 'Action', value: 'Echo' }
      ]
    };

    const echoResult = await loader(wasmBinary, echoMessage, environment);
    console.log('📋 Echo Response:', JSON.stringify(echoResult, null, 2));

    if (echoResult.Output.includes('Hello from Go test!')) {
      console.log('✅ Test passed: Echo action works correctly');
    } else {
      console.log('❌ Test failed: Echo did not work as expected');
      process.exit(1);
    }

    // Test ProcessInfo action
    console.log('\n📨 Testing ProcessInfo action...');
    const infoMessage = {
      Target: 'test-process',
      Action: 'ProcessInfo',
      Data: '',
      Anchor: '0',
      Tags: [
        { name: 'Action', value: 'ProcessInfo' }
      ]
    };

    const infoResult = await loader(wasmBinary, infoMessage, environment);
    console.log('📋 ProcessInfo Response:', JSON.stringify(infoResult, null, 2));

    try {
      const infoData = JSON.parse(infoResult.Output);
      if (infoData.ProcessId === 'test-process-id' && infoData.Owner === 'test-owner') {
        console.log('✅ Test passed: ProcessInfo action works correctly');
      } else {
        console.log('❌ Test failed: ProcessInfo returned unexpected data');
        process.exit(1);
      }
    } catch (e) {
      console.log('❌ Test failed: ProcessInfo did not return valid JSON');
      process.exit(1);
    }

    // Test unknown action
    console.log('\n📨 Testing unknown action...');
    const unknownMessage = {
      Target: 'test-process',
      Action: 'UnknownAction',
      Data: '',
      Anchor: '0',
      Tags: [
        { name: 'Action', value: 'UnknownAction' }
      ]
    };

    const unknownResult = await loader(wasmBinary, unknownMessage, environment);
    console.log('📋 Unknown Action Response:', JSON.stringify(unknownResult, null, 2));

    if (unknownResult.Output.includes('Unknown action') && unknownResult.Error.includes('Supported actions')) {
      console.log('✅ Test passed: Unknown action handling works correctly');
    } else {
      console.log('❌ Test failed: Unknown action handling did not work as expected');
      process.exit(1);
    }

    console.log('\n🎉 All tests passed!');

  } catch (error) {
    console.error('❌ Test failed with error:', error);
    process.exit(1);
  }
}

test();
