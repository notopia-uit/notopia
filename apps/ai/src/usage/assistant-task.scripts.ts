export const CREATE_TASK = `
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
redis.call('SET', KEYS[1], ARGV[1], 'PXAT', ARGV[2])
return 1
`;

const loadTask = `
local raw = redis.call('GET', KEYS[1])
if not raw then return nil end
local task = cjson.decode(raw)
local clock = redis.call('TIME')
local now = tonumber(clock[1]) * 1000 + math.floor(tonumber(clock[2]) / 1000)
if task.deadline <= now then return nil end
`;

export const SUSPEND_TASK = `${loadTask}
if task.state ~= 'running' or task.revision ~= tonumber(ARGV[1]) then return nil end
local update = cjson.decode(ARGV[2])
if update.stepsUsed < 1 or update.stepsUsed >= task.remainingSteps then return nil end
if task.targetNoteId and task.targetNoteId ~= update.targetNoteId then return nil end
task.messages = update.messages
task.pendingCalls = update.pendingCalls
task.targetNoteId = update.targetNoteId
task.remainingSteps = task.remainingSteps - update.stepsUsed
task.tokenHash = ARGV[3]
task.state = 'awaiting-client'
task.revision = task.revision + 1
local encoded = cjson.encode(task)
if string.len(encoded) > 512000 then return nil end
redis.call('SET', KEYS[1], encoded, 'KEEPTTL')
return tostring(task.deadline)
`;

export const CLAIM_TASK = `${loadTask}
if task.owner.userId ~= ARGV[1] or task.owner.workspaceId ~= ARGV[2] then return nil end
if task.state ~= 'awaiting-client' or task.tokenHash ~= ARGV[3] then return nil end
task.tokenHash = nil
task.state = 'running'
task.revision = task.revision + 1
local encoded = cjson.encode(task)
redis.call('SET', KEYS[1], encoded, 'KEEPTTL')
return encoded
`;

export const FINISH_TASK = `${loadTask}
if task.owner.userId ~= ARGV[1] or task.owner.workspaceId ~= ARGV[2] then return nil end
if task.state == 'running' or task.state == 'awaiting-client' then
  task.state = ARGV[3]
  task.tokenHash = nil
  task.revision = task.revision + 1
  redis.call('SET', KEYS[1], cjson.encode(task), 'KEEPTTL')
end
return 1
`;
