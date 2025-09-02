#!/usr/bin/env node

import { readFileSync } from 'fs';
import { describe, it } from 'node:test';
import assert from 'assert';
import { default as AoLoader } from '@permaweb/ao-loader';

const STUB_ADDRESS = 'arweave-address-'.padEnd(43, '1');

const AO_LOADER_OPTIONS = {
  format: 'wasm32-unknown-emscripten3',
  inputEncoding: 'JSON-1',
  outputEncoding: 'JSON-1',
  memoryLimit: '1073741824', // 1 GiB in bytes
  computeLimit: (9e12).toString(),
  extensions: [],
};

const AO_LOADER_HANDLER_ENV = {
  Process: {
    Id: STUB_ADDRESS,
    Owner: STUB_ADDRESS,
    Tags: [],
  },
  Module: {
    Id: ''.padEnd(43, '1'),
    Tags: [],
  },
};

const DEFAULT_HANDLE_OPTIONS = {
  Id: ''.padEnd(43, '1'),
  ['Block-Height']: '1',
  Owner: STUB_ADDRESS,
  Module: 'AO-GO',
  Target: STUB_ADDRESS,
  From: STUB_ADDRESS,
  Timestamp: Date.now(),
  Reference: '1',
};

const WASM_MODULE = readFileSync('./build/process.wasm');

async function createHandleFunction(wasmModule = WASM_MODULE) {
  const handle = await AoLoader(wasmModule, AO_LOADER_OPTIONS);
  return { handle, memory: null };
}

function createHandleWrapper(
  ogHandle: any,
  startMem: any,
  defaultHandleOptions = DEFAULT_HANDLE_OPTIONS,
  aoLoaderHandlerEnv = AO_LOADER_HANDLER_ENV,
) {
  return async function (options = {}, mem = startMem) {
    return ogHandle(
      mem,
      {
        ...defaultHandleOptions,
        ...options,
      },
      aoLoaderHandlerEnv,
    );
  };
}

describe('AO Go Process', async () => {
  const { handle: ogHandle, memory: startMem } = await createHandleFunction();
  const handle = createHandleWrapper(ogHandle, startMem);

  it('should respond to Hello action', async () => {
    const result = await handle({
      Tags: [{ name: 'Action', value: 'Hello' }],
    });

    console.log('📋 Hello Response:', JSON.stringify(result, null, 2));
    assert.strictEqual(result.Output, 'Hello, world!');
  });

  it('should echo data with Echo action', async () => {
    const testData = 'Hello from test!';
    const result = await handle({
      Tags: [{ name: 'Action', value: 'Echo' }],
      Data: testData,
    });

    console.log('📋 Echo Response:', JSON.stringify(result, null, 2));
    assert(result.Output.includes(testData), 'Echo should include the input data');
  });
});
