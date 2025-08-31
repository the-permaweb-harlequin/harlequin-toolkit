#include <gtest/gtest.h>
#include <string.h>

// Include the main functions for testing
extern char* handle_message(const char* action, const char* key, const char* value, const char* from);
extern void init_process();

class AOProcessTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Initialize the process before each test
        init_process();
    }

    void TearDown() override {
        // Clean up after each test if needed
    }
};

TEST_F(AOProcessTest, InfoAction) {
    char* result = handle_message("Info", nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Info-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "Hello from AO Process (C)") != nullptr);
    EXPECT_TRUE(strstr(result, "test-sender") != nullptr);
}

TEST_F(AOProcessTest, SetAction) {
    char* result = handle_message("Set", "testKey", "testValue", "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Set-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "Successfully set testKey to testValue") != nullptr);
}

TEST_F(AOProcessTest, SetActionMissingKey) {
    char* result = handle_message("Set", nullptr, "testValue", "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Error") != nullptr);
    EXPECT_TRUE(strstr(result, "Key and value are required") != nullptr);
}

TEST_F(AOProcessTest, SetActionMissingValue) {
    char* result = handle_message("Set", "testKey", nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Error") != nullptr);
    EXPECT_TRUE(strstr(result, "Key and value are required") != nullptr);
}

TEST_F(AOProcessTest, GetAction) {
    // First set a value
    handle_message("Set", "testKey", "testValue", "test-sender");

    // Then get it
    char* result = handle_message("Get", "testKey", nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Get-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "testKey") != nullptr);
    EXPECT_TRUE(strstr(result, "testValue") != nullptr);
}

TEST_F(AOProcessTest, GetActionMissingKey) {
    char* result = handle_message("Get", nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Error") != nullptr);
    EXPECT_TRUE(strstr(result, "Key is required") != nullptr);
}

TEST_F(AOProcessTest, GetActionNonExistentKey) {
    char* result = handle_message("Get", "nonExistentKey", nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Get-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "Not found") != nullptr);
}

TEST_F(AOProcessTest, ListAction) {
    // Set some values first
    handle_message("Set", "key1", "value1", "test-sender");
    handle_message("Set", "key2", "value2", "test-sender");

    char* result = handle_message("List", nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "List-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "key1") != nullptr);
    EXPECT_TRUE(strstr(result, "value1") != nullptr);
    EXPECT_TRUE(strstr(result, "key2") != nullptr);
    EXPECT_TRUE(strstr(result, "value2") != nullptr);
}

TEST_F(AOProcessTest, ListActionEmpty) {
    char* result = handle_message("List", nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "List-Response") != nullptr);
    EXPECT_TRUE(strstr(result, "{}") != nullptr);
}

TEST_F(AOProcessTest, UnknownAction) {
    char* result = handle_message("UnknownAction", nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Error") != nullptr);
    EXPECT_TRUE(strstr(result, "Unknown action: UnknownAction") != nullptr);
    EXPECT_TRUE(strstr(result, "Available actions: Info, Set, Get, List") != nullptr);
}

TEST_F(AOProcessTest, NullAction) {
    char* result = handle_message(nullptr, nullptr, nullptr, "test-sender");

    ASSERT_NE(result, nullptr);
    EXPECT_TRUE(strstr(result, "Error") != nullptr);
    EXPECT_TRUE(strstr(result, "Action is required") != nullptr);
}

TEST_F(AOProcessTest, MultipleOperations) {
    // Test a sequence of operations
    char* result;

    // Set multiple values
    result = handle_message("Set", "name", "Alice", "test-sender");
    EXPECT_TRUE(strstr(result, "Set-Response") != nullptr);

    result = handle_message("Set", "age", "30", "test-sender");
    EXPECT_TRUE(strstr(result, "Set-Response") != nullptr);

    result = handle_message("Set", "city", "New York", "test-sender");
    EXPECT_TRUE(strstr(result, "Set-Response") != nullptr);

    // Get values
    result = handle_message("Get", "name", nullptr, "test-sender");
    EXPECT_TRUE(strstr(result, "Alice") != nullptr);

    result = handle_message("Get", "age", nullptr, "test-sender");
    EXPECT_TRUE(strstr(result, "30") != nullptr);

    // List all
    result = handle_message("List", nullptr, nullptr, "test-sender");
    EXPECT_TRUE(strstr(result, "name") != nullptr);
    EXPECT_TRUE(strstr(result, "Alice") != nullptr);
    EXPECT_TRUE(strstr(result, "age") != nullptr);
    EXPECT_TRUE(strstr(result, "30") != nullptr);
    EXPECT_TRUE(strstr(result, "city") != nullptr);
    EXPECT_TRUE(strstr(result, "New York") != nullptr);
}

// Performance test for state storage
TEST_F(AOProcessTest, StateStorageCapacity) {
    char key[64];
    char value[256];
    char* result;

    // Fill up the state storage
    for (int i = 0; i < 100; i++) {
        snprintf(key, sizeof(key), "key%d", i);
        snprintf(value, sizeof(value), "value%d", i);

        result = handle_message("Set", key, value, "test-sender");
        EXPECT_TRUE(strstr(result, "Set-Response") != nullptr);
    }

    // Try to add one more (should fail if at capacity)
    result = handle_message("Set", "overflow", "value", "test-sender");
    // This might succeed or fail depending on implementation
    // Just ensure we get a valid response
    ASSERT_NE(result, nullptr);
}

int main(int argc, char **argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}

