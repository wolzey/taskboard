--[[
  Deletes a job from the queue

  Input:
    KEYS[1] 'prefix' - Queue prefix (e.g., 'bull:myqueue')

    ARGV[1] jobId - The job ID to delete

  Output:
    1 if successful (job deleted)
    0 if job not found

  Events:
    'removed' event
]]

local rcall = redis.call
local prefix = KEYS[1]
local jobId = ARGV[1]

local jobKey = prefix .. ":" .. jobId

-- Check if the job exists
local exists = rcall("EXISTS", jobKey)
if exists == 0 then
  return 0
end

-- All possible states where the job could be
local states = {"wait", "active", "delayed", "failed", "completed", "paused", "waiting-children"}

-- Remove from all possible states
for _, state in ipairs(states) do
  local stateKey = prefix .. ":" .. state

  if state == "wait" or state == "active" or state == "paused" or state == "waiting-children" then
    -- These are stored as lists
    rcall("LREM", stateKey, 0, jobId)
  else
    -- delayed, failed, completed are stored as sorted sets
    rcall("ZREM", stateKey, jobId)
  end
end

-- Delete the job hash itself
rcall("DEL", jobKey)

-- Delete any job logs
rcall("DEL", jobKey .. ":logs")

-- Delete any job progress data
rcall("DEL", jobKey .. ":progress")

-- Emit removed event
rcall("XADD", prefix .. ":events", "*", "event", "removed", "jobId", jobId)

return 1
