// Simple JSON utilities for AssemblyScript
// This provides basic JSON parsing and stringification

export class JSONValue {
  static String(value: string): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.STRING;
    result.stringValue = value;
    return result;
  }

  static Number(value: f64): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.NUMBER;
    result.numberValue = value;
    return result;
  }

  static Bool(value: bool): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.BOOL;
    result.boolValue = value;
    return result;
  }

  static Null(): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.NULL;
    return result;
  }

  static Object(): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.OBJECT;
    result.objectValue = new Map<string, JSONValue>();
    return result;
  }

  static Array(): JSONValue {
    const result = new JSONValue();
    result.kind = JSONValueKind.ARRAY;
    result.arrayValue = new Array<JSONValue>();
    return result;
  }

  kind: JSONValueKind = JSONValueKind.NULL;
  stringValue: string = "";
  numberValue: f64 = 0;
  boolValue: bool = false;
  objectValue: Map<string, JSONValue> | null = null;
  arrayValue: Array<JSONValue> | null = null;

  isString(): bool {
    return this.kind == JSONValueKind.STRING;
  }

  isNumber(): bool {
    return this.kind == JSONValueKind.NUMBER;
  }

  isBool(): bool {
    return this.kind == JSONValueKind.BOOL;
  }

  isNull(): bool {
    return this.kind == JSONValueKind.NULL;
  }

  isObject(): bool {
    return this.kind == JSONValueKind.OBJECT;
  }

  isArray(): bool {
    return this.kind == JSONValueKind.ARRAY;
  }

  toString(): string {
    if (this.isString()) {
      return `"${this.stringValue}"`;
    } else if (this.isNumber()) {
      return this.numberValue.toString();
    } else if (this.isBool()) {
      return this.boolValue ? "true" : "false";
    } else if (this.isNull()) {
      return "null";
    } else if (this.isObject()) {
      let result = "{";
      let first = true;
      const obj = this.objectValue!;
      const keys = obj.keys();

      for (let i = 0; i < keys.length; i++) {
        if (!first) result += ",";
        const key = keys[i];
        const value = obj.get(key);
        result += `"${key}":${value.toString()}`;
        first = false;
      }

      result += "}";
      return result;
    } else if (this.isArray()) {
      let result = "[";
      const arr = this.arrayValue!;

      for (let i = 0; i < arr.length; i++) {
        if (i > 0) result += ",";
        result += arr[i].toString();
      }

      result += "]";
      return result;
    }

    return "null";
  }
}

enum JSONValueKind {
  STRING,
  NUMBER,
  BOOL,
  NULL,
  OBJECT,
  ARRAY
}

export class SimpleJSON {
  static stringify(obj: Map<string, string>): string {
    let result = "{";
    let first = true;
    const keys = obj.keys();

    for (let i = 0; i < keys.length; i++) {
      if (!first) result += ",";
      const key = keys[i];
      const value = obj.get(key);
      result += `"${key}":"${value}"`;
      first = false;
    }

    result += "}";
    return result;
  }

  static parseSimple(json: string): Map<string, string> | null {
    const result = new Map<string, string>();

    // Very basic JSON object parser - expects {"key":"value","key2":"value2"}
    if (!json.startsWith("{") || !json.endsWith("}")) {
      return null;
    }

    const content = json.substring(1, json.length - 1).trim();
    if (content.length == 0) {
      return result;
    }

    const pairs = content.split(",");
    for (let i = 0; i < pairs.length; i++) {
      const pair = pairs[i].trim();
      const colonIndex = pair.indexOf(":");

      if (colonIndex <= 0) continue;

      let key = pair.substring(0, colonIndex).trim();
      let value = pair.substring(colonIndex + 1).trim();

      // Remove quotes
      if (key.startsWith('"') && key.endsWith('"')) {
        key = key.substring(1, key.length - 1);
      }
      if (value.startsWith('"') && value.endsWith('"')) {
        value = value.substring(1, value.length - 1);
      }

      result.set(key, value);
    }

    return result;
  }
}

