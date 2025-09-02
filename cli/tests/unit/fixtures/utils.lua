-- Sample utils module for testing
local utils = {}

function utils.formatResponse(action, data)
    return {
        Action = action,
        Data = data,
        Timestamp = os.time()
    }
end

function utils.formatError(message)
    return {
        Error = message,
        Timestamp = os.time()
    }
end

function utils.isEmpty(str)
    return str == nil or str == ""
end

return utils
