-- Sample Lua file for testing bundling
local utils = require("utils")

function handle(msg)
    local action = msg.Tags.Action

    if action == "Info" then
        return utils.formatResponse("Process info", {
            name = "Test Process",
            version = "1.0.0"
        })
    elseif action == "Echo" then
        return utils.formatResponse("Echo", {
            message = msg.Data
        })
    else
        return utils.formatError("Unknown action: " .. action)
    end
end
