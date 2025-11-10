--[[
   Get counts per provided states

   Input:
	KEYS[1] 'prefix'

	ARGV[1...] types
]]

local rcall = redis.call
local prefix = KEYS[1]
local results = {}

for i = 1, #ARGV do
	local stateKey = prefix .. ":" .. ARGV[i]
	local keyType = rcall("TYPE", stateKey)["ok"]

	if ARGV[i] == "wait" or ARGV[i] == "paused" then
		if keyType == "list" then
			local marker = rcall("LINDEX", stateKey, -1)

			if marker and string.sub(marker, 1, 2) == "0:" then
				local count = rcall("LLEN", stateKey)

				if count > 1 then
					rcall("RPOP", stateKey)
					results[#results+1] = count-1
				else
					results[#results+1] = 0
				end
			else
				results[#results+1] = rcall("LLEN", stateKey)
			end
		else
			results[#results+1] = 0
		end
	elseif ARGV[i] == "active" then
		if keyType == "list" then
			results[#results+1] = rcall("LLEN", stateKey)
		else
			results[#results+1] = 0
		end
	else
		if keyType == "zset" then
			results[#results+1] = rcall("ZCARD", stateKey)
		else
			results[#results+1] = 0
		end
	end
end

return results
