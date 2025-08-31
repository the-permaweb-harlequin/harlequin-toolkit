-- Handlers module for AO Process
-- This module provides utilities for message handling

local handlers = {}

-- Utility function to create standardized responses
function handlers.create_response(target, action, data, extra_fields)
    local response = {
        Target = target or "unknown",
        Action = action,
        Data = data
    }

    -- Add any extra fields
    if extra_fields then
        for key, value in pairs(extra_fields) do
            response[key] = value
        end
    end

    return response
end

-- Utility function to create error responses
function handlers.create_error(target, message)
    return handlers.create_response(target, "Error", message)
end

-- Utility function to validate required fields
function handlers.validate_required(fields, msg)
    local missing = {}

    for _, field in ipairs(fields) do
        if not msg[field] and not (msg.Tags and msg.Tags[field]) then
            table.insert(missing, field)
        end
    end

    return #missing == 0, missing
end

-- Utility function to get field value from message or tags
function handlers.get_field(msg, field)
    return msg[field] or (msg.Tags and msg.Tags[field])
end

-- Utility function to log handler activity
function handlers.log(level, message, context)
    local log_entry = {
        timestamp = os.time(),
        level = level,
        message = message,
        context = context or {}
    }

    -- In a real AO environment, this might use a proper logging system
    print(string.format("[%s] %s: %s", level, os.date("%Y-%m-%d %H:%M:%S"), message))

    if context then
        for key, value in pairs(context) do
            print(string.format("  %s: %s", key, tostring(value)))
        end
    end
end

-- Utility function to sanitize input
function handlers.sanitize_string(input, max_length)
    if not input or type(input) ~= "string" then
        return nil
    end

    -- Remove control characters and limit length
    local sanitized = input:gsub("%c", "")

    if max_length and #sanitized > max_length then
        sanitized = sanitized:sub(1, max_length)
    end

    return sanitized
end

-- Utility function to validate key format
function handlers.is_valid_key(key)
    if not key or type(key) ~= "string" then
        return false
    end

    -- Key should be alphanumeric with underscores and hyphens allowed
    return key:match("^[%w_%-]+$") ~= nil and #key <= 64
end

return handlers

