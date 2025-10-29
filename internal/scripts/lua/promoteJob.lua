--[[
  Promotes a job from any state (delayed, failed, completed, waiting-children) to the wait queue

  Input:
    KEYS[1] 'prefix' - Queue prefix (e.g., 'bull:myqueue')

    ARGV[1] jobId - The job ID to promote
    ARGV[2] fromState - The current state of the job ('delayed', 'failed', 'completed', 'waiting-children')

  Output:
    1 if successful
    0 if job not found in the specified state
    -1 if invalid state

  Events:
    'waiting' event
]]

local rcall = redis.call
local prefix = KEYS[1]
local jobId = ARGV[1]
local fromState = ARGV[2]

-- Validate state
local validStates = {
  delayed = true,
  failed = true,
  completed = true,
  ["waiting-children"] = true
}

if not validStates[fromState] then
  return -1
end

-- Construct the keys
local waitKey = prefix .. ":wait"
local stateKey = prefix .. ":" .. fromState
local jobKey = prefix .. ":" .. jobId

-- Check if the job exists
local exists = rcall("EXISTS", jobKey)
if exists == 0 then
  return 0
end

-- Remove from current state
local removed = 0
if fromState == "wait" or fromState == "paused" or fromState == "waiting-children" then
  -- These are stored as lists
  removed = rcall("LREM", stateKey, 0, jobId)
else
  -- delayed, failed, completed are stored as sorted sets
  removed = rcall("ZREM", stateKey, jobId)
end

if removed == 0 then
  return 0
end

-- Add to wait queue (add to the right/end of the list)
rcall("RPUSH", waitKey, jobId)

-- Remove delay timestamp if it exists
rcall("HDEL", jobKey, "delay")

-- Update job timestamp
local timestamp = rcall("TIME")
rcall("HSET", jobKey, "timestamp", timestamp[1] * 1000 + timestamp[2] / 1000)

-- Emit waiting event
rcall("XADD", prefix .. ":events", "*", "event", "waiting", "jobId", jobId, "prev", fromState)

return 1
