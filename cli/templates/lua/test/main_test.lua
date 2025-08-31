-- Test suite for main.lua AO process
-- Uses busted testing framework

describe("AO Process Tests", function()
    local process

    before_each(function()
        -- Reset state before each test
        State = {}

        -- Mock AO environment
        ao = {
            id = "test-process-id",
            send = function(msg)
                -- Store sent messages for verification
                ao.sent = ao.sent or {}
                table.insert(ao.sent, msg)
            end,
            sent = {}
        }

        -- Mock Handlers
        Handlers = {
            handlers = {},
            add = function(name, pattern, handler)
                Handlers.handlers[name] = {
                    pattern = pattern,
                    handler = handler
                }
            end,
            utils = {
                hasMatchingTag = function(key, value)
                    return function(msg)
                        return msg.Tags and msg.Tags[key] == value
                    end
                end
            }
        }

        -- Load the process
        dofile("main.lua")
    end)

    describe("Info Handler", function()
        it("should respond with process info", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "Info" }
            }

            local handler = Handlers.handlers.info.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("test-sender", ao.sent[1].Target)
            assert.are.equal("Info-Response", ao.sent[1].Action)
            assert.is_truthy(string.find(ao.sent[1].Data, "Hello from AO Process"))
        end)
    end)

    describe("Set Handler", function()
        it("should set a key-value pair", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "Set", Key = "testKey" },
                Data = "testValue"
            }

            local handler = Handlers.handlers.set.handler
            handler(msg)

            assert.are.equal("testValue", State.testKey)
            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Set-Response", ao.sent[1].Action)
        end)

        it("should return error when key is missing", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "Set" },
                Data = "testValue"
            }

            local handler = Handlers.handlers.set.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Error", ao.sent[1].Action)
            assert.are.equal("Key is required", ao.sent[1].Data)
        end)
    end)

    describe("Get Handler", function()
        it("should retrieve a stored value", function()
            -- Set up state
            State.testKey = "testValue"

            local msg = {
                From = "test-sender",
                Tags = { Action = "Get", Key = "testKey" }
            }

            local handler = Handlers.handlers.get.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Get-Response", ao.sent[1].Action)
            assert.are.equal("testKey", ao.sent[1].Key)
            assert.are.equal("testValue", ao.sent[1].Data)
        end)

        it("should return 'Not found' for missing key", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "Get", Key = "missingKey" }
            }

            local handler = Handlers.handlers.get.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Get-Response", ao.sent[1].Action)
            assert.are.equal("Not found", ao.sent[1].Data)
        end)

        it("should return error when key is missing", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "Get" }
            }

            local handler = Handlers.handlers.get.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Error", ao.sent[1].Action)
            assert.are.equal("Key is required", ao.sent[1].Data)
        end)
    end)

    describe("List Handler", function()
        it("should return all state as JSON", function()
            -- Set up state
            State.key1 = "value1"
            State.key2 = "value2"

            -- Mock JSON module
            package.loaded.json = {
                encode = function(data)
                    return '{"key1":"value1","key2":"value2"}'
                end
            }

            local msg = {
                From = "test-sender",
                Tags = { Action = "List" }
            }

            local handler = Handlers.handlers.list.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("List-Response", ao.sent[1].Action)
            assert.is_truthy(string.find(ao.sent[1].Data, "key1"))
            assert.is_truthy(string.find(ao.sent[1].Data, "value1"))
        end)
    end)

    describe("Default Handler", function()
        it("should handle unknown actions", function()
            local msg = {
                From = "test-sender",
                Tags = { Action = "UnknownAction" }
            }

            local handler = Handlers.handlers.default.handler
            handler(msg)

            assert.is_not_nil(ao.sent[1])
            assert.are.equal("Error", ao.sent[1].Action)
            assert.is_truthy(string.find(ao.sent[1].Data, "Unknown action"))
        end)
    end)
end)

