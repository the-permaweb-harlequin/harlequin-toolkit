#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <emscripten.h>

#include "lua.h"
#include "lauxlib.h"
#include "lualib.h"

// Global Lua state
static lua_State *L = NULL;

// Initialize the Lua interpreter and load the main script
EMSCRIPTEN_KEEPALIVE
int init_lua_process() {
    // Create new Lua state
    L = luaL_newstate();
    if (!L) {
        printf("Error: Failed to create Lua state\n");
        return 0;
    }

    // Open standard libraries
    luaL_openlibs(L);

    // Load and execute the main Lua script
    int result = luaL_dofile(L, "main.lua");
    if (result != LUA_OK) {
        printf("Error loading main.lua: %s\n", lua_tostring(L, -1));
        lua_close(L);
        L = NULL;
        return 0;
    }

    printf("Lua process initialized successfully\n");
    return 1;
}

// Clean up the Lua state
EMSCRIPTEN_KEEPALIVE
void cleanup_lua_process() {
    if (L) {
        lua_close(L);
        L = NULL;
    }
}

// Handle incoming messages by calling the Lua handle function
EMSCRIPTEN_KEEPALIVE
char* handle_message(const char* message_json) {
    if (!L) {
        return strdup("{\"error\":\"Lua state not initialized\"}");
    }

    if (!message_json) {
        return strdup("{\"error\":\"No message provided\"}");
    }

    // Get the global handle function
    lua_getglobal(L, "handle");
    if (!lua_isfunction(L, -1)) {
        lua_pop(L, 1);
        return strdup("{\"error\":\"handle function not found\"}");
    }

    // Parse the JSON message and create a Lua table
    // For simplicity, we'll pass the JSON string directly
    // In a full implementation, you'd parse JSON to Lua table
    lua_pushstring(L, message_json);

    // Call the handle function with 1 argument, expecting 1 result
    int call_result = lua_pcall(L, 1, 1, 0);
    if (call_result != LUA_OK) {
        const char* error_msg = lua_tostring(L, -1);
        char* error_response = malloc(strlen(error_msg) + 50);
        snprintf(error_response, strlen(error_msg) + 50,
                "{\"error\":\"Lua execution error: %s\"}", error_msg);
        lua_pop(L, 1);
        return error_response;
    }

    // Get the result from Lua
    const char* result = lua_tostring(L, -1);
    char* response = strdup(result ? result : "{\"error\":\"No response from handle function\"}");
    lua_pop(L, 1);

    return response;
}

// AO-compatible message handler that mimics the AO dev-cli structure
EMSCRIPTEN_KEEPALIVE
char* ao_handle(const char* msg_id, const char* msg_from, const char* msg_owner,
                const char* msg_target, const char* msg_anchor, const char* msg_data,
                const char* msg_tags, const char* msg_timestamp, const char* msg_block_height,
                const char* msg_hash_chain) {

    if (!L) {
        return strdup("{\"error\":\"Lua state not initialized\"}");
    }

    // Create a Lua table representing the AO message
    lua_newtable(L);

    // Set message fields
    if (msg_id) {
        lua_pushstring(L, "Id");
        lua_pushstring(L, msg_id);
        lua_settable(L, -3);
    }

    if (msg_from) {
        lua_pushstring(L, "From");
        lua_pushstring(L, msg_from);
        lua_settable(L, -3);
    }

    if (msg_owner) {
        lua_pushstring(L, "Owner");
        lua_pushstring(L, msg_owner);
        lua_settable(L, -3);
    }

    if (msg_target) {
        lua_pushstring(L, "Target");
        lua_pushstring(L, msg_target);
        lua_settable(L, -3);
    }

    if (msg_anchor) {
        lua_pushstring(L, "Anchor");
        lua_pushstring(L, msg_anchor);
        lua_settable(L, -3);
    }

    if (msg_data) {
        lua_pushstring(L, "Data");
        lua_pushstring(L, msg_data);
        lua_settable(L, -3);
    }

    if (msg_timestamp) {
        lua_pushstring(L, "Timestamp");
        lua_pushstring(L, msg_timestamp);
        lua_settable(L, -3);
    }

    if (msg_block_height) {
        lua_pushstring(L, "Block-Height");
        lua_pushstring(L, msg_block_height);
        lua_settable(L, -3);
    }

    if (msg_hash_chain) {
        lua_pushstring(L, "Hash-Chain");
        lua_pushstring(L, msg_hash_chain);
        lua_settable(L, -3);
    }

    // Parse and set tags if provided
    if (msg_tags) {
        lua_pushstring(L, "Tags");
        lua_newtable(L);

        // Simple tag parsing (assumes "key1=value1,key2=value2" format)
        char* tags_copy = strdup(msg_tags);
        char* tag_pair = strtok(tags_copy, ",");

        while (tag_pair) {
            char* equals_pos = strchr(tag_pair, '=');
            if (equals_pos) {
                *equals_pos = '\0';
                char* key = tag_pair;
                char* value = equals_pos + 1;

                lua_pushstring(L, key);
                lua_pushstring(L, value);
                lua_settable(L, -3);
            }
            tag_pair = strtok(NULL, ",");
        }

        free(tags_copy);
        lua_settable(L, -3);
    }

    // Get the global handle function
    lua_getglobal(L, "handle");
    if (!lua_isfunction(L, -1)) {
        lua_pop(L, 2); // Pop function and message table
        return strdup("{\"error\":\"handle function not found\"}");
    }

    // Move the message table to the top of the stack
    lua_insert(L, -2);

    // Call the handle function with the message table
    int call_result = lua_pcall(L, 1, 1, 0);
    if (call_result != LUA_OK) {
        const char* error_msg = lua_tostring(L, -1);
        char* error_response = malloc(strlen(error_msg) + 50);
        snprintf(error_response, strlen(error_msg) + 50,
                "{\"error\":\"Lua execution error: %s\"}", error_msg);
        lua_pop(L, 1);
        return error_response;
    }

    // Get the result from Lua
    const char* result = lua_tostring(L, -1);
    char* response = strdup(result ? result : "{\"error\":\"No response from handle function\"}");
    lua_pop(L, 1);

    return response;
}

// Utility function to execute arbitrary Lua code
EMSCRIPTEN_KEEPALIVE
char* eval_lua(const char* lua_code) {
    if (!L) {
        return strdup("{\"error\":\"Lua state not initialized\"}");
    }

    if (!lua_code) {
        return strdup("{\"error\":\"No Lua code provided\"}");
    }

    int result = luaL_dostring(L, lua_code);
    if (result != LUA_OK) {
        const char* error_msg = lua_tostring(L, -1);
        char* error_response = malloc(strlen(error_msg) + 50);
        snprintf(error_response, strlen(error_msg) + 50,
                "{\"error\":\"Lua execution error: %s\"}", error_msg);
        lua_pop(L, 1);
        return error_response;
    }

    // If there's a result on the stack, return it
    if (lua_gettop(L) > 0) {
        const char* result_str = lua_tostring(L, -1);
        char* response = strdup(result_str ? result_str : "{\"result\":\"nil\"}");
        lua_pop(L, 1);
        return response;
    }

    return strdup("{\"result\":\"success\"}");
}

// Get the current state of the Lua process
EMSCRIPTEN_KEEPALIVE
char* get_lua_state() {
    if (!L) {
        return strdup("{\"error\":\"Lua state not initialized\"}");
    }

    // Get the global State variable
    lua_getglobal(L, "State");
    if (lua_isnil(L, -1)) {
        lua_pop(L, 1);
        return strdup("{\"State\":{}}");
    }

    // Convert the State table to a JSON-like string
    // This is a simplified implementation
    if (lua_istable(L, -1)) {
        // For now, just return a placeholder
        // In a full implementation, you'd serialize the table to JSON
        lua_pop(L, 1);
        return strdup("{\"State\":\"<table>\"}");
    }

    const char* state_str = lua_tostring(L, -1);
    char* response = malloc(strlen(state_str) + 20);
    snprintf(response, strlen(state_str) + 20, "{\"State\":\"%s\"}", state_str);
    lua_pop(L, 1);

    return response;
}

// Main function for testing
int main() {
    printf("Initializing Lua AO Process Trampoline...\n");

    if (!init_lua_process()) {
        printf("Failed to initialize Lua process\n");
        return 1;
    }

    // Test the handle function
    char* result = handle_message("{\"Action\":\"Info\",\"From\":\"test-sender\"}");
    printf("Test result: %s\n", result);
    free(result);

    cleanup_lua_process();
    printf("Lua process cleaned up\n");

    return 0;
}

