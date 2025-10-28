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
-- ARGV[5]: Random string for sliding window unique value (length less than or equal to 255)
local function calculate_tokens_in_range(key, start_time, end_time)
    local valid_list = redis.call('ZRANGEBYSCORE', key, start_time, end_time)
    local valid_tokens = 0
    for _, v in ipairs(valid_list) do
        local _, tokens = struct.unpack('Bc0L', v)
        valid_tokens = valid_tokens + tokens
    end
    return valid_tokens
end

local sliding_window_key = tostring(KEYS[1])
local token_bucket_key = tostring(KEYS[2])
local token_encoder_key = tostring(KEYS[3])

local estimated = tonumber(ARGV[1])
local current_timestamp = tonumber(ARGV[2])
local bucket_capacity = tonumber(ARGV[3])
local window_size = tonumber(ARGV[4])
local random_string = tostring(ARGV[5])

-- Valid window start time
local window_start = current_timestamp - window_size
-- Waiting time
local waiting_time = 0
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
-- Plus the difference from the token encoder if it exists
local ttl = redis.call('PTTL', token_encoder_key)
local difference = tonumber(redis.call('GET', token_encoder_key))
if ttl < 0 then
    difference = 0
else
    if difference + estimated >= 0 then
        estimated = estimated + difference
    else
        redis.call('SET', token_encoder_key, 0)
    end
end
-- Check if the request can be satisfied
if max_capacity < estimated or estimated <= 0 then -- If max capacity is less than estimated consumption or estimated is less than or equal to 0, return -1 indicating rejection
    waiting_time = -1
elseif current_capacity < estimated then -- If current capacity is insufficient to satisfy estimated consumption, calculate waiting time
    -- Get the earliest valid timestamp
    local first_valid_window = redis.call('ZRANGE', sliding_window_key, 0, 0, 'WITHSCORES')
    local first_valid_start = tonumber(first_valid_window[2])
    if not first_valid_start then
        first_valid_start = current_timestamp
    end
    -- Waiting time = fixed delay + window size - valid window interval
    waiting_time = 3 + window_size - (current_timestamp - first_valid_start)
else -- Otherwise, capacity satisfies estimated consumption, no waiting required, update data
    redis.call('ZADD', sliding_window_key, current_timestamp,
        struct.pack('Bc0L', string.len(random_string), random_string, estimated))
    current_capacity = current_capacity - estimated
    redis.call('HSET', token_bucket_key, 'capacity', current_capacity)
end

-- Set expiration time to window size plus 5 seconds buffer
redis.call('PEXPIRE', sliding_window_key, window_size + 5000)
redis.call('PEXPIRE', token_bucket_key, window_size + 5000)
redis.call('PEXPIRE', token_encoder_key, window_size + 5000)

return {current_capacity, waiting_time, estimated, difference}
