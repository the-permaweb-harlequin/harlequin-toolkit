//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
)

// AOMessage represents an AO message
type AOMessage struct {
	Target string                 `json:"Target"`
	Action string                 `json:"Action"`
	Data   string                 `json:"Data"`
	Anchor string                 `json:"Anchor"`
	Tags   []map[string]string    `json:"Tags"`
}

// AOEnvironment represents the AO environment
type AOEnvironment struct {
	Process AOProcess `json:"Process"`
	Module  string    `json:"Module"`
}

// AOProcess represents the process information
type AOProcess struct {
	ID    string              `json:"Id"`
	Owner string              `json:"Owner"`
	Tags  []map[string]string `json:"Tags"`
}

// AOResponse represents the AO response format
type AOResponse struct {
	Output      string        `json:"Output"`
	Error       string        `json:"Error"`
	Messages    []interface{} `json:"Messages"`
	Spawns      []interface{} `json:"Spawns"`
	Assignments []interface{} `json:"Assignments"`
	GasUsed     int           `json:"GasUsed"`
}

// handleAO processes an AO message and returns the response
func handleAO(msgJson, envJson string) string {
	// Parse the message JSON
	var msg AOMessage
	if err := json.Unmarshal([]byte(msgJson), &msg); err != nil {
		return createErrorResponse("Invalid message JSON: " + err.Error())
	}

	// Parse the environment JSON
	var env AOEnvironment
	if err := json.Unmarshal([]byte(envJson), &env); err != nil {
		return createErrorResponse("Invalid environment JSON: " + err.Error())
	}

	// Extract action from message tags or Action field
	action := msg.Action
	if action == "" {
		for _, tag := range msg.Tags {
			if tag["name"] == "Action" || tag["Name"] == "Action" {
				action = tag["value"]
				if action == "" {
					action = tag["Value"]
				}
				break
			}
		}
	}

	// Process the action
	response := AOResponse{
		Output:      "",
		Error:       "",
		Messages:    []interface{}{},
		Spawns:      []interface{}{},
		Assignments: []interface{}{},
		GasUsed:     0,
	}

	switch action {
	case "Hello":
		response.Output = "Hello, AO World from Go!"
	case "Info":
		response.Output = "This is an AO process built with Go and WebAssembly"
	case "Echo":
		response.Output = "Echo: " + msg.Data
	case "ProcessInfo":
		info := map[string]interface{}{
			"ProcessId": env.Process.ID,
			"Owner":     env.Process.Owner,
			"Module":    env.Module,
		}
		infoJson, _ := json.Marshal(info)
		response.Output = string(infoJson)
	default:
		response.Output = "Unknown action: " + action
		response.Error = "Supported actions: Hello, Info, Echo, ProcessInfo"
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return createErrorResponse("Failed to marshal response: " + err.Error())
	}

	return string(jsonBytes)
}

// createErrorResponse creates an error response
func createErrorResponse(errMsg string) string {
	response := AOResponse{
		Output:      "",
		Error:       errMsg,
		Messages:    []interface{}{},
		Spawns:      []interface{}{},
		Assignments: []interface{}{},
		GasUsed:     0,
	}

	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// handle is the main export function that AO loader calls
func handle(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		errorResp := createErrorResponse("Invalid number of arguments")
		return stringToArrayBuffer(errorResp)
	}

	msgJson := args[0].String()
	envJson := args[1].String()

	result := handleAO(msgJson, envJson)

	// Convert to ArrayBuffer for AO compatibility
	return stringToArrayBuffer(result)
}

// Convert string to ArrayBuffer for AO compatibility
func stringToArrayBuffer(s string) js.Value {
	bytes := []byte(s)
	uint8Array := js.Global().Get("Uint8Array").New(len(bytes))
	js.CopyBytesToJS(uint8Array, bytes)
	return uint8Array.Get("buffer")
}

func main() {
	// Export the handle function globally so AO loader can find it
	js.Global().Set("handle", js.FuncOf(handle))

	// Keep the program running
	select {}
}
