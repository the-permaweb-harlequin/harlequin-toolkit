#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <emscripten.h>

// AO Process Template in C
// This demonstrates basic message handling and state management

// Simple key-value store
#define MAX_ENTRIES 100
#define MAX_KEY_LENGTH 64
#define MAX_VALUE_LENGTH 256

typedef struct {
    char key[MAX_KEY_LENGTH];
    char value[MAX_VALUE_LENGTH];
    int active;
} StateEntry;

static StateEntry state[MAX_ENTRIES];
static int state_count = 0;

// Initialize the state
void init_state() {
    for (int i = 0; i < MAX_ENTRIES; i++) {
        state[i].active = 0;
    }
    state_count = 0;
}

// Find a state entry by key
int find_state_entry(const char* key) {
    for (int i = 0; i < MAX_ENTRIES; i++) {
        if (state[i].active && strcmp(state[i].key, key) == 0) {
            return i;
        }
    }
    return -1;
}

// Set a key-value pair
int set_state(const char* key, const char* value) {
    if (!key || !value) return 0;

    // Check if key already exists
    int index = find_state_entry(key);

    if (index >= 0) {
        // Update existing entry
        strncpy(state[index].value, value, MAX_VALUE_LENGTH - 1);
        state[index].value[MAX_VALUE_LENGTH - 1] = '\0';
        return 1;
    }

    // Find empty slot
    for (int i = 0; i < MAX_ENTRIES; i++) {
        if (!state[i].active) {
            strncpy(state[i].key, key, MAX_KEY_LENGTH - 1);
            state[i].key[MAX_KEY_LENGTH - 1] = '\0';
            strncpy(state[i].value, value, MAX_VALUE_LENGTH - 1);
            state[i].value[MAX_VALUE_LENGTH - 1] = '\0';
            state[i].active = 1;
            state_count++;
            return 1;
        }
    }

    return 0; // No space available
}

// Get a value by key
const char* get_state(const char* key) {
    if (!key) return NULL;

    int index = find_state_entry(key);
    if (index >= 0) {
        return state[index].value;
    }

    return NULL;
}

// Export functions for AO integration
EMSCRIPTEN_KEEPALIVE
char* handle_message(const char* action, const char* key, const char* value, const char* from) {
    static char response[1024];

    if (!action) {
        snprintf(response, sizeof(response),
                "{\"Target\":\"%s\",\"Action\":\"Error\",\"Data\":\"Action is required\"}",
                from ? from : "unknown");
        return response;
    }

    if (strcmp(action, "Info") == 0) {
        snprintf(response, sizeof(response),
                "{\"Target\":\"%s\",\"Action\":\"Info-Response\",\"Data\":\"Hello from AO Process (C)! State entries: %d\"}",
                from ? from : "unknown", state_count);
        return response;
    }

    if (strcmp(action, "Set") == 0) {
        if (!key || !value) {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Error\",\"Data\":\"Key and value are required\"}",
                    from ? from : "unknown");
            return response;
        }

        if (set_state(key, value)) {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Set-Response\",\"Data\":\"Successfully set %s to %s\"}",
                    from ? from : "unknown", key, value);
        } else {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Error\",\"Data\":\"Failed to set %s (storage full?)\"}",
                    from ? from : "unknown", key);
        }
        return response;
    }

    if (strcmp(action, "Get") == 0) {
        if (!key) {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Error\",\"Data\":\"Key is required\"}",
                    from ? from : "unknown");
            return response;
        }

        const char* stored_value = get_state(key);
        if (stored_value) {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Get-Response\",\"Key\":\"%s\",\"Data\":\"%s\"}",
                    from ? from : "unknown", key, stored_value);
        } else {
            snprintf(response, sizeof(response),
                    "{\"Target\":\"%s\",\"Action\":\"Get-Response\",\"Key\":\"%s\",\"Data\":\"Not found\"}",
                    from ? from : "unknown", key);
        }
        return response;
    }

    if (strcmp(action, "List") == 0) {
        char state_json[2048] = "{";
        int first = 1;

        for (int i = 0; i < MAX_ENTRIES; i++) {
            if (state[i].active) {
                if (!first) {
                    strcat(state_json, ",");
                }
                char entry[256];
                snprintf(entry, sizeof(entry), "\"%s\":\"%s\"", state[i].key, state[i].value);
                strcat(state_json, entry);
                first = 0;
            }
        }
        strcat(state_json, "}");

        snprintf(response, sizeof(response),
                "{\"Target\":\"%s\",\"Action\":\"List-Response\",\"Data\":%s}",
                from ? from : "unknown", state_json);
        return response;
    }

    // Unknown action
    snprintf(response, sizeof(response),
            "{\"Target\":\"%s\",\"Action\":\"Error\",\"Data\":\"Unknown action: %s. Available actions: Info, Set, Get, List\"}",
            from ? from : "unknown", action);
    return response;
}

// Initialize the process
EMSCRIPTEN_KEEPALIVE
void init_process() {
    init_state();
    printf("AO Process (C) initialized\n");
}

// Main function (for testing)
int main() {
    init_process();

    // Test the functionality
    printf("Testing C AO Process...\n");

    // Test Info
    char* result = handle_message("Info", NULL, NULL, "test-sender");
    printf("Info: %s\n", result);

    // Test Set
    result = handle_message("Set", "testKey", "testValue", "test-sender");
    printf("Set: %s\n", result);

    // Test Get
    result = handle_message("Get", "testKey", NULL, "test-sender");
    printf("Get: %s\n", result);

    // Test List
    result = handle_message("List", NULL, NULL, "test-sender");
    printf("List: %s\n", result);

    return 0;
}

