local prefix = ARGV[1]
local cursor = ARGV[2]
local count = ARGV[3] or '100'
local rcall = redis.call
local queues = {}

local match_pattern = prefix .. '*' .. ':meta'

repeat
	local scan_result = rcall('SCAN', cursor, 'MATCH', match_pattern, 'COUNT', count)
	cursor = scan_result[1]
	local keys = scan_result[2]

	for _, queue in ipairs(keys) do
		local inserted_queue = queue:gsub(":meta", "")

		table.insert(queues, inserted_queue)
    end
until cursor == '0' or #queues >= 1000 -- large amounts of queues... no way

return queues

