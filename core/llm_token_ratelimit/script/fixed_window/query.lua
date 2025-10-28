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
-- KEYS[1]: Fixed Window Key ("<redisRatelimitKey>")
-- ARGV[1]: Maximum Token capacity
-- ARGV[2]: Window size (milliseconds)
local fixed_window_key = KEYS[1]

local max_token_capacity = tonumber(ARGV[1])
local window_size = tonumber(ARGV[2])

local ttl = redis.call('PTTL', fixed_window_key)
if ttl < 0 then
    redis.call('SET', fixed_window_key, max_token_capacity, 'PX', window_size)
    return {max_token_capacity, window_size}
end
return {tonumber(redis.call('GET', fixed_window_key)), ttl}
