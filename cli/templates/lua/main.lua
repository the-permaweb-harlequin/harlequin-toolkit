-- AO Process Template
-- This is a basic AO process template that demonstrates:
-- - Message handling
-- - State management
-- - Response formatting
-- - C trampoline integration

-- Initialize process state
State = State or {}

-- Load utility modules
local handlers_utils = require('handlers.init')
local utils = require('utils.init')

-- Global handle function for C trampoline integration
function handle(msg)
    -- Log incoming message
    handlers_utils.log("INFO", "Received message", {
        action = msg.Action or msg.Tags and msg.Tags.Action,
        from = msg.From,
        id = msg.Id
    })

    -- Determine action from message
    local action = msg.Action or (msg.Tags and msg.Tags.Action)

    if not action then
        return utils.json_encode(handlers_utils.create_error(msg.From, "Action is required"))
    end

    -- Route to appropriate handler
    if action == "Info" then
        return handle_info(msg)
    elseif action == "Set" then
        return handle_set(msg)
    elseif action == "Get" then
        return handle_get(msg)
    elseif action == "List" then
        return handle_list(msg)
    else
        return utils.json_encode(handlers_utils.create_error(
            msg.From,
            "Unknown action: " .. action .. ". Available actions: Info, Set, Get, List"
        ))
    end
end

-- Handler functions
function handle_info(msg)
    local response = handlers_utils.create_response(
        msg.From,
        "Info-Response",
        "Hello from AO Process! Process ID: " .. (ao and ao.id or "unknown") ..
        ". State entries: " .. utils.table_size(State)
    )
    return utils.json_encode(response)
end

function handle_set(msg)
    local key = handlers_utils.get_field(msg, "Key")
    local value = msg.Data

    if not key then
        return utils.json_encode(handlers_utils.create_error(msg.From, "Key is required"))
    end

    if not handlers_utils.is_valid_key(key) then
        return utils.json_encode(handlers_utils.create_error(msg.From, "Invalid key format"))
    end

    -- Sanitize the value
    local sanitized_value = handlers_utils.sanitize_string(value, 1000)
    if not sanitized_value then
        return utils.json_encode(handlers_utils.create_error(msg.From, "Invalid value"))
    end

    State[key] = sanitized_value

    local response = handlers_utils.create_response(
        msg.From,
        "Set-Response",
        "Successfully set " .. key .. " to " .. sanitized_value
    )
    return utils.json_encode(response)
end

function handle_get(msg)
    local key = handlers_utils.get_field(msg, "Key")

    if not key then
        return utils.json_encode(handlers_utils.create_error(msg.From, "Key is required"))
    end

    local value = State[key] or "Not found"

    local response = handlers_utils.create_response(
        msg.From,
        "Get-Response",
        value,
        { Key = key }
    )
    return utils.json_encode(response)
end

function handle_list(msg)
    local response = handlers_utils.create_response(
        msg.From,
        "List-Response",
        utils.json_encode(State)
    )
    return utils.json_encode(response)
end

-- Handler for incoming messages
Handlers.add(
    "info",
    Handlers.utils.hasMatchingTag("Action", "Info"),
    function(msg)
        ao.send({
            Target = msg.From,
            Action = "Info-Response",
            Data = "Hello from AO Process! Process ID: " .. ao.id
        })
    end
)

-- Handler for setting data
Handlers.add(
    "set",
    Handlers.utils.hasMatchingTag("Action", "Set"),
    function(msg)
        local key = msg.Tags.Key
        local value = msg.Data

        if not key then
            ao.send({
                Target = msg.From,
                Action = "Error",
                Data = "Key is required"
            })
            return
        end

        State[key] = value

        ao.send({
            Target = msg.From,
            Action = "Set-Response",
            Data = "Successfully set " .. key .. " to " .. value
        })
    end
)

-- Handler for getting data
Handlers.add(
    "get",
    Handlers.utils.hasMatchingTag("Action", "Get"),
    function(msg)
        local key = msg.Tags.Key

        if not key then
            ao.send({
                Target = msg.From,
                Action = "Error",
                Data = "Key is required"
            })
            return
        end

        local value = State[key] or "Not found"

        ao.send({
            Target = msg.From,
            Action = "Get-Response",
            Key = key,
            Data = value
        })
    end
)

-- Handler for listing all state
Handlers.add(
    "list",
    Handlers.utils.hasMatchingTag("Action", "List"),
    function(msg)
        local stateJson = require('json').encode(State)

        ao.send({
            Target = msg.From,
            Action = "List-Response",
            Data = stateJson
        })
    end
)

-- Default handler for unrecognized actions
Handlers.add(
    "default",
    function(msg)
        return not msg.Action or not Handlers.utils.hasMatchingTag("Action", "Info")(msg) and
               not Handlers.utils.hasMatchingTag("Action", "Set")(msg) and
               not Handlers.utils.hasMatchingTag("Action", "Get")(msg) and
               not Handlers.utils.hasMatchingTag("Action", "List")(msg)
    end,
    function(msg)
        ao.send({
            Target = msg.From,
            Action = "Error",
            Data = "Unknown action: " .. (msg.Action or "none") .. ". Available actions: Info, Set, Get, List"
        })
    end
)
