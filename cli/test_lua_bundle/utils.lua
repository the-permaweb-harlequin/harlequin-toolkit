-- utils.lua - A simple utility module
local function helper()
    return "Hello from utils!"
end

local function calculate(x, y)
    return x + y + 10
end

return {
    helper = helper,
    calculate = calculate
}
