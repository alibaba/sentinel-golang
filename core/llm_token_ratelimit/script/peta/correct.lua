-- Copyright 1999-2020 Alibaba Group Holding Ltd.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
-- KEYS[1]: Sliding Window Key ("{shard-<hashtag>}:sliding-window:<redisRatelimitKey>")
-- KEYS[2]: Token Bucket Key ("{shard-<hashtag>}:token-bucket:<redisRatelimitKey>")
-- KEYS[3]: Token Encoder Key ("{shard-<hashtag>}:token-encoder:<provider>:<model>:<redisRatelimitKey>")
-- ARGV[1]: Estimated token consumption
-- ARGV[2]: Current timestamp (milliseconds)
-- ARGV[3]: Token bucket capacity
-- ARGV[4]: Window size (milliseconds)
-- ARGV[5]: Actual token consumption
-- ARGV[6]: Random string for sliding window value (length less than or equal to 255)
local MAX_SEARCH_ITRATIONS = 64

local function calculate_tokens_in_range(key, start_time, end_time)
    local valid_list = redis.call('ZRANGEBYSCORE', key, start_time, end_time)
    local valid_tokens = 0
    for _, v in ipairs(valid_list) do
        local _, tokens = struct.unpack('Bc0L', v)
        valid_tokens = valid_tokens + tokens
    end
    return valid_tokens
end

local function binary_search_compensation_time(key, L, R, window_size, max_capacity, predicted_error)
    local iter = 0
    while L < R and iter < MAX_SEARCH_ITRATIONS do
        iter = iter + 1
        local mid = math.floor((L + R) / 2)
        local valid_tokens = calculate_tokens_in_range(key, mid - window_size, mid)
        if valid_tokens + predicted_error <= max_capacity then
            R = mid
        else
            L = mid + 1
        end
    end
    return L
end

local sliding_window_key = tostring(KEYS[1])
local token_bucket_key = tostring(KEYS[2])
local token_encoder_key = tostring(KEYS[3])

local estimated = tonumber(ARGV[1])
local current_timestamp = tonumber(ARGV[2])
local bucket_capacity = tonumber(ARGV[3])
local window_size = tonumber(ARGV[4])
local actual = tonumber(ARGV[5])
local random_string = tostring(ARGV[6])

-- Valid window start time
local window_start = current_timestamp - window_size
-- Get bucket
local bucket = redis.call('HMGET', token_bucket_key, 'capacity', 'max_capacity')
local current_capacity = tonumber(bucket[1])
local max_capacity = tonumber(bucket[2])
-- Initialize bucket manually if it doesn't exist
if not current_capacity then
    current_capacity = bucket_capacity
    max_capacity = bucket_capacity
    redis.call('HMSET', token_bucket_key, 'capacity', bucket_capacity, 'max_capacity', bucket_capacity)
    redis.call('ZADD', sliding_window_key, current_timestamp,
        struct.pack('Bc0L', string.len(random_string), random_string, 0))
end
-- Calculate expired tokens
local released_tokens = calculate_tokens_in_range(sliding_window_key, 0, window_start)
if released_tokens > 0 then -- Expired tokens exist, attempt to replenish new tokens
    -- Clean up expired data
    redis.call('ZREMRANGEBYSCORE', sliding_window_key, 0, window_start)
    -- Calculate valid tokens
    local valid_tokens = calculate_tokens_in_range(sliding_window_key, '-inf', '+inf')
    -- Update token count
    if current_capacity + released_tokens > max_capacity then -- If current capacity plus released tokens exceeds max capacity, reset to max capacity minus valid tokens
        current_capacity = max_capacity - valid_tokens
    else -- Otherwise, directly add the released tokens
        current_capacity = current_capacity + released_tokens
    end
    -- Immediately replenish new tokens
    redis.call('HSET', token_bucket_key, 'capacity', current_capacity)
end
-- Update the difference from the token encoder
local difference = actual - estimated
redis.call('SET', token_encoder_key, difference)
-- Correction result for reservation
local correct_result = 0
if estimated < 0 or actual < 0 then
    correct_result = 3 -- Invalid value
elseif estimated < actual then -- Underestimation
    -- Mainly handle underestimation cases to properly limit actual usage; overestimation may reject requests but won't affect downstream services
    -- Calculate prediction error
    local predicted_error = math.abs(actual - estimated)
    -- directly deduct all underestimated tokens
    current_capacity = current_capacity - predicted_error
    redis.call('HSET', token_bucket_key, 'capacity', current_capacity)
    -- Get the latest valid timestamp
    local last_valid_window = redis.call('ZRANGE', sliding_window_key, -1, -1, 'WITHSCORES')
    local compensation_start = tonumber(last_valid_window[2])
    if not compensation_start then -- Possibly all data just expired, use current timestamp minus window size as start
        compensation_start = current_timestamp
    end
    while predicted_error ~= 0 do -- Distribute to future windows until all error is distributed
        if max_capacity >= predicted_error then
            local compensation_time = binary_search_compensation_time(sliding_window_key, compensation_start,
                compensation_start + window_size, window_size, max_capacity, predicted_error)
            if calculate_tokens_in_range(sliding_window_key, compensation_time - window_size, compensation_time) +
                predicted_error > max_capacity then
                correct_result = 1 -- If the compensation time exceeds max capacity, return 1 to indicate failure
                break
            end
            redis.call('ZADD', sliding_window_key, compensation_time,
                struct.pack('Bc0L', string.len(random_string), random_string, predicted_error))
            predicted_error = 0
        else
            redis.call('ZADD', sliding_window_key, compensation_start,
                struct.pack('Bc0L', string.len(random_string), random_string, max_capacity))
            predicted_error = predicted_error - max_capacity
            compensation_start = compensation_start + window_size
        end
    end
elseif estimated > actual then -- Overestimation
    correct_result = 2
end

-- Set expiration time to window size plus 5 seconds buffer
redis.call('PEXPIRE', sliding_window_key, window_size + 5000)
redis.call('PEXPIRE', token_bucket_key, window_size + 5000)
redis.call('PEXPIRE', token_encoder_key, window_size + 5000)

return {correct_result}
