import { JSON } from "assemblyscript-json/assembly";

export function handle(msgJson: string, envJson: string): ArrayBuffer {

  const result = new JSON.Obj()
  result.set("Messages", new JSON.Arr())
  result.set("Spawns", new JSON.Arr())
  result.set("Assignments", new JSON.Arr())
  result.set("Output", "Hello, world!")
  result.set("Error", "")

  if (msgJson.includes('"Action":"Hello"')) {
    result.set("Output", "Hello, world!")
  } else {
    result.set("Output", "Unknown action")
    result.set("Error", "Supported actions: Hello")
  }

  // Return response as ArrayBuffer
  return String.UTF8.encode(result.stringify());
}
