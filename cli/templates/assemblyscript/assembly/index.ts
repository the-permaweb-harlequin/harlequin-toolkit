// AO Process Template in AssemblyScript
// This demonstrates basic message handling and state management

import { JSON } from "assemblyscript-json/assembly";

// AO Message structure
@json
class AOMessage {
  Id: string | null = null;
  From: string | null = null;
  Owner: string | null = null;
  Target: string | null = null;
  Anchor: string | null = null;
  Data: string | null = null;
  Tags: Map<string, string> | null = null;
  Timestamp: string | null = null;
  "Block-Height": string | null = null;
  "Hash-Chain": string | null = null;
}

// AO Response structure
@json
class AOResponse {
  Target: string;
  Action: string;
  Data: string;

  constructor(target: string, action: string, data: string) {
    this.Target = target;
    this.Action = action;
    this.Data = data;
  }
}

// Enhanced response with extra fields
@json
class AOResponseWithFields extends AOResponse {
  Key: string | null = null;

  constructor(target: string, action: string, data: string, key: string | null = null) {
    super(target, action, data);
    this.Key = key;
  }
}

// Global state storage
const state = new Map<string, string>();

// Logging function
function log(message: string): void {
  // In WebAssembly, we can use trace for debugging
  trace(message);
}

// Utility functions
function isValidKey(key: string): bool {
  if (key.length == 0 || key.length > 64) {
    return false;
  }

  // Check if key contains only alphanumeric characters, underscores, and hyphens
  for (let i = 0; i < key.length; i++) {
    const char = key.charCodeAt(i);
    const isAlphaNumeric = (char >= 48 && char <= 57) || // 0-9
                          (char >= 65 && char <= 90) || // A-Z
                          (char >= 97 && char <= 122);  // a-z
    const isUnderscore = char == 95; // _
    const isHyphen = char == 45;     // -

    if (!isAlphaNumeric && !isUnderscore && !isHyphen) {
      return false;
    }
  }

  return true;
}

function sanitizeValue(value: string): string | null {
  if (value.length > 1000) {
    return null;
  }

  // Remove control characters
  let sanitized = "";
  for (let i = 0; i < value.length; i++) {
    const char = value.charCodeAt(i);
    if (char >= 32 && char <= 126) { // Printable ASCII characters
      sanitized += value.charAt(i);
    }
  }

  return sanitized;
}

function getTagValue(tags: Map<string, string> | null, key: string): string | null {
  if (tags == null) {
    return null;
  }
  return tags.has(key) ? tags.get(key) : null;
}

function createErrorResponse(target: string, message: string): AOResponse {
  return new AOResponse(target, "Error", message);
}

// Message handlers
function handleInfo(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";
  const stateSize = state.size.toString();
  const data = `Hello from AO Process (AssemblyScript)! State entries: ${stateSize}`;

  log(`Info request from ${from}`);
  return new AOResponse(from, "Info-Response", data);
}

function handleSet(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";

  const key = getTagValue(msg.Tags, "Key");
  if (key == null) {
    return createErrorResponse(from, "Key is required");
  }

  if (!isValidKey(key)) {
    return createErrorResponse(from, "Invalid key format. Use alphanumeric characters, underscores, and hyphens only");
  }

  const value = msg.Data;
  if (value == null) {
    return createErrorResponse(from, "Value is required");
  }

  const sanitizedValue = sanitizeValue(value);
  if (sanitizedValue == null) {
    return createErrorResponse(from, "Invalid value or value too long (max 1000 characters)");
  }

  state.set(key, sanitizedValue);

  const data = `Successfully set ${key} to ${sanitizedValue}`;
  log(`Set: ${key} = ${sanitizedValue}`);

  return new AOResponse(from, "Set-Response", data);
}

function handleGet(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";

  const key = getTagValue(msg.Tags, "Key");
  if (key == null) {
    return createErrorResponse(from, "Key is required");
  }

  const value = state.has(key) ? state.get(key) : "Not found";

  log(`Get: ${key} = ${value}`);
  return new AOResponseWithFields(from, "Get-Response", value, key);
}

function handleList(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";

  // Create JSON representation of state
  let stateJson = "{";
  const keys = state.keys();
  let first = true;

  for (let i = 0; i < keys.length; i++) {
    const key = keys[i];
    const value = state.get(key);

    if (!first) {
      stateJson += ",";
    }
    stateJson += `"${key}":"${value}"`;
    first = false;
  }
  stateJson += "}";

  log(`List: returning ${state.size} entries`);
  return new AOResponse(from, "List-Response", stateJson);
}

function handleRemove(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";

  const key = getTagValue(msg.Tags, "Key");
  if (key == null) {
    return createErrorResponse(from, "Key is required");
  }

  const existed = state.has(key);
  if (existed) {
    state.delete(key);
  }

  const data = existed ?
    `Successfully removed ${key}` :
    `Key ${key} not found`;

  log(`Remove: ${key} (existed: ${existed})`);
  return new AOResponse(from, "Remove-Response", data);
}

function handleClear(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";

  const previousSize = state.size;
  state.clear();

  log(`Clear: removed ${previousSize} entries`);
  return new AOResponse(from, "Clear-Response", "State cleared successfully");
}

// Parse tags from string format "key1=value1,key2=value2"
function parseTags(tagsStr: string): Map<string, string> {
  const tags = new Map<string, string>();

  if (tagsStr.length == 0) {
    return tags;
  }

  const pairs = tagsStr.split(",");
  for (let i = 0; i < pairs.length; i++) {
    const pair = pairs[i].trim();
    const equalIndex = pair.indexOf("=");

    if (equalIndex > 0 && equalIndex < pair.length - 1) {
      const key = pair.substring(0, equalIndex).trim();
      const value = pair.substring(equalIndex + 1).trim();
      tags.set(key, value);
    }
  }

  return tags;
}

// Main message handler
function handleMessage(msg: AOMessage): AOResponse {
  log(`Received message from ${msg.From || "unknown"}`);

  const action = getTagValue(msg.Tags, "Action");
  if (action == null) {
    return createErrorResponse(msg.From || "unknown", "Action is required");
  }

  let response: AOResponse;

  if (action == "Info") {
    response = handleInfo(msg);
  } else if (action == "Set") {
    response = handleSet(msg);
  } else if (action == "Get") {
    response = handleGet(msg);
  } else if (action == "List") {
    response = handleList(msg);
  } else if (action == "Remove") {
    response = handleRemove(msg);
  } else if (action == "Clear") {
    response = handleClear(msg);
  } else {
    const errorMsg = `Unknown action: ${action}. Available actions: Info, Set, Get, List, Remove, Clear`;
    response = createErrorResponse(msg.From || "unknown", errorMsg);
  }

  log(`Sending response: ${response.Action}`);
  return response;
}

// WebAssembly exports
export function initProcess(): void {
  log("AO Process (AssemblyScript) initialized");
}

export function handle(messageJson: string): string {
  try {
    // Parse the JSON message
    const jsonObj = JSON.parse(messageJson);

    // Create AOMessage from JSON
    const msg = new AOMessage();

    // Extract fields from JSON
    if (jsonObj.has("Id")) {
      msg.Id = jsonObj.getString("Id")!.valueOf();
    }
    if (jsonObj.has("From")) {
      msg.From = jsonObj.getString("From")!.valueOf();
    }
    if (jsonObj.has("Owner")) {
      msg.Owner = jsonObj.getString("Owner")!.valueOf();
    }
    if (jsonObj.has("Target")) {
      msg.Target = jsonObj.getString("Target")!.valueOf();
    }
    if (jsonObj.has("Anchor")) {
      msg.Anchor = jsonObj.getString("Anchor")!.valueOf();
    }
    if (jsonObj.has("Data")) {
      msg.Data = jsonObj.getString("Data")!.valueOf();
    }
    if (jsonObj.has("Timestamp")) {
      msg.Timestamp = jsonObj.getString("Timestamp")!.valueOf();
    }
    if (jsonObj.has("Block-Height")) {
      msg["Block-Height"] = jsonObj.getString("Block-Height")!.valueOf();
    }
    if (jsonObj.has("Hash-Chain")) {
      msg["Hash-Chain"] = jsonObj.getString("Hash-Chain")!.valueOf();
    }

    // Parse tags
    if (jsonObj.has("Tags")) {
      const tagsObj = jsonObj.getObj("Tags")!;
      const tags = new Map<string, string>();

      const tagKeys = tagsObj.keys;
      for (let i = 0; i < tagKeys.length; i++) {
        const key = tagKeys[i];
        const value = tagsObj.getString(key);
        if (value != null) {
          tags.set(key, value.valueOf());
        }
      }
      msg.Tags = tags;
    }

    // Handle the message
    const response = handleMessage(msg);

    // Convert response to JSON
    let responseJson = `{"Target":"${response.Target}","Action":"${response.Action}","Data":"${response.Data}"`;

    // Add extra fields if present
    if (response instanceof AOResponseWithFields) {
      const extendedResponse = response as AOResponseWithFields;
      if (extendedResponse.Key != null) {
        responseJson += `,"Key":"${extendedResponse.Key}"`;
      }
    }

    responseJson += "}";
    return responseJson;

  } catch (error) {
    const errorResponse = `{"Target":"unknown","Action":"Error","Data":"JSON parse error or processing error"}`;
    return errorResponse;
  }
}

export function getState(): string {
  try {
    // Convert state to JSON
    let stateJson = "{";
    const keys = state.keys();
    let first = true;

    for (let i = 0; i < keys.length; i++) {
      const key = keys[i];
      const value = state.get(key);

      if (!first) {
        stateJson += ",";
      }
      stateJson += `"${key}":"${value}"`;
      first = false;
    }
    stateJson += "}";

    return stateJson;
  } catch (error) {
    return "{}";
  }
}

export function clearState(): bool {
  try {
    state.clear();
    log("State cleared");
    return true;
  } catch (error) {
    return false;
  }
}

export function getStateSize(): i32 {
  return state.size;
}

// Utility exports for testing
export function setState(key: string, value: string): bool {
  if (!isValidKey(key)) {
    return false;
  }

  const sanitizedValue = sanitizeValue(value);
  if (sanitizedValue == null) {
    return false;
  }

  state.set(key, sanitizedValue);
  return true;
}

export function getStateValue(key: string): string {
  return state.has(key) ? state.get(key) : "";
}

