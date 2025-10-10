local prefix = ARGV[1]
local queue = ARGV[2]
local state = ARGV[3]
local start = ARGV[4]
local stop = ARGV[5]

local results = {}
local rcall = redis.call
local key = prefix ":" .. queue .. ":" .. state
local total_count = rcall('ZCARD', key)
local revrange_result = rcall('ZREVRANGE', key, start, stop)

results['count'] = total_count
results['jobs'] = revrange_result

return results
