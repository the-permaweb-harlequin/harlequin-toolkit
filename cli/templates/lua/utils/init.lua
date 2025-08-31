-- Utilities module for AO Process
-- This module provides common utility functions

local utils = {}

-- JSON handling utilities
function utils.json_encode(data)
    local json = require('json')
    local success, result = pcall(json.encode, data)

    if success then
        return result
    else
        -- Fallback to simple serialization
        return utils.simple_serialize(data)
    end
end

function utils.json_decode(str)
    local json = require('json')
    local success, result = pcall(json.decode, str)

    if success then
        return result
    else
        return nil, "Invalid JSON"
    end
end

-- Simple serialization for basic data types
function utils.simple_serialize(data)
    if type(data) == "table" then
        local parts = {}
        for key, value in pairs(data) do
            table.insert(parts, string.format('"%s":"%s"', tostring(key), tostring(value)))
        end
        return "{" .. table.concat(parts, ",") .. "}"
    else
        return tostring(data)
    end
end

-- String utilities
function utils.trim(str)
    if not str then return nil end
    return str:match("^%s*(.-)%s*$")
end

function utils.split(str, delimiter)
    if not str then return {} end

    local result = {}
    local pattern = string.format("([^%s]+)", delimiter or "%s")

    for match in str:gmatch(pattern) do
        table.insert(result, match)
    end

    return result
end

function utils.starts_with(str, prefix)
    if not str or not prefix then return false end
    return str:sub(1, #prefix) == prefix
end

function utils.ends_with(str, suffix)
    if not str or not suffix then return false end
    return str:sub(-#suffix) == suffix
end

-- Table utilities
function utils.table_copy(original)
    if type(original) ~= 'table' then return original end

    local copy = {}
    for key, value in pairs(original) do
        copy[key] = utils.table_copy(value)
    end

    return copy
end

function utils.table_merge(t1, t2)
    local result = utils.table_copy(t1)

    for key, value in pairs(t2) do
        result[key] = value
    end

    return result
end

function utils.table_keys(tbl)
    local keys = {}
    for key, _ in pairs(tbl) do
        table.insert(keys, key)
    end
    return keys
end

function utils.table_values(tbl)
    local values = {}
    for _, value in pairs(tbl) do
        table.insert(values, value)
    end
    return values
end

function utils.table_size(tbl)
    local count = 0
    for _ in pairs(tbl) do
        count = count + 1
    end
    return count
end

-- Validation utilities
function utils.is_empty(value)
    if value == nil then return true end
    if type(value) == "string" then return #utils.trim(value) == 0 end
    if type(value) == "table" then return utils.table_size(value) == 0 end
    return false
end

function utils.is_valid_email(email)
    if not email or type(email) ~= "string" then return false end
    return email:match("^[%w%._%+%-]+@[%w%._%+%-]+%.%w+$") ~= nil
end

function utils.is_valid_url(url)
    if not url or type(url) ~= "string" then return false end
    return url:match("^https?://[%w%._%+%-/]+$") ~= nil
end

-- Crypto utilities (basic)
function utils.simple_hash(str)
    if not str then return nil end

    local hash = 0
    for i = 1, #str do
        hash = ((hash * 31) + str:byte(i)) % 2147483647
    end

    return tostring(hash)
end

-- Time utilities
function utils.timestamp()
    return os.time()
end

function utils.format_timestamp(timestamp, format)
    format = format or "%Y-%m-%d %H:%M:%S"
    return os.date(format, timestamp)
end

function utils.time_ago(timestamp)
    local now = os.time()
    local diff = now - timestamp

    if diff < 60 then
        return "just now"
    elseif diff < 3600 then
        return math.floor(diff / 60) .. " minutes ago"
    elseif diff < 86400 then
        return math.floor(diff / 3600) .. " hours ago"
    else
        return math.floor(diff / 86400) .. " days ago"
    end
end

-- Math utilities
function utils.round(num, decimals)
    local mult = 10^(decimals or 0)
    return math.floor(num * mult + 0.5) / mult
end

function utils.clamp(value, min, max)
    return math.max(min, math.min(max, value))
end

function utils.random_string(length)
    length = length or 8
    local chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    local result = {}

    for i = 1, length do
        local rand = math.random(1, #chars)
        table.insert(result, chars:sub(rand, rand))
    end

    return table.concat(result)
end

-- Error handling utilities
function utils.safe_call(func, ...)
    local success, result = pcall(func, ...)

    if success then
        return result, nil
    else
        return nil, result
    end
end

function utils.retry(func, max_attempts, delay)
    max_attempts = max_attempts or 3
    delay = delay or 1

    for attempt = 1, max_attempts do
        local result, error = utils.safe_call(func)

        if result then
            return result, nil
        end

        if attempt < max_attempts then
            -- In a real environment, you might want to use a proper sleep function
            local start = os.time()
            while os.time() - start < delay do
                -- Busy wait (not ideal, but works for simple cases)
            end
        else
            return nil, error
        end
    end
end

return utils

