export const RESERVE = `
local now = tonumber(ARGV[1])
local quota = tonumber(ARGV[2])
local expiry = tonumber(ARGV[3])
if redis.call('EXISTS', KEYS[2]) == 1 then return 'duplicate' end
redis.call('ZREMRANGEBYSCORE', KEYS[3], '-inf', now)
if redis.call('ZCARD', KEYS[3]) >= tonumber(ARGV[4]) then return 'concurrency' end
if tonumber(redis.call('GET', KEYS[4]) or '0') >= tonumber(ARGV[5]) then return 'rate' end
local used = tonumber(redis.call('GET', KEYS[1]) or '0')
if quota ~= -1 and used >= quota then return 'quota' end
redis.call('INCR', KEYS[1])
redis.call('EXPIREAT', KEYS[1], expiry)
redis.call('SET', KEYS[2], 'reserved:' .. ARGV[7], 'EXAT', expiry)
redis.call('ZADD', KEYS[3], now + tonumber(ARGV[6]), ARGV[7])
redis.call('PEXPIRE', KEYS[3], tonumber(ARGV[6]) + 1000)
if redis.call('INCR', KEYS[4]) == 1 then redis.call('EXPIRE', KEYS[4], 60) end
return 'ok'
`;

export const START = `
if redis.call('GET', KEYS[1]) ~= 'reserved:' .. ARGV[1] then return 0 end
redis.call('SET', KEYS[1], 'started:' .. ARGV[1], 'KEEPTTL')
return 1
`;

export const RELEASE = `
if ARGV[2] == '1' and redis.call('GET', KEYS[2]) == 'reserved:' .. ARGV[1] then
  if tonumber(redis.call('GET', KEYS[1]) or '0') > 0 then redis.call('DECR', KEYS[1]) end
  redis.call('DEL', KEYS[2])
end
redis.call('ZREM', KEYS[3], ARGV[1])
return 1
`;

export const THROTTLE = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('EXPIRE', KEYS[1], 60) end
if count > tonumber(ARGV[1]) then return 0 end
return 1
`;
