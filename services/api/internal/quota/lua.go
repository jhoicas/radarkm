package quota

// consumeQuotaScript atomically:
// 1) honors dedup key (no extra decrement)
// 2) rejects when used >= limit
// 3) otherwise INCR + EXPIRE(ttl) + SET dedup EX 90s
//
// KEYS[1] = quota:{device}:{day}
// KEYS[2] = dedup:{device}:{offer}
// ARGV[1] = limit (N)
// ARGV[2] = quota TTL seconds until next Bogotá midnight (relative to injectable Clock)
// ARGV[3] = dedup TTL seconds
const consumeQuotaScript = `
local limit = tonumber(ARGV[1])
local quotaTTL = tonumber(ARGV[2])
local dedupTTL = tonumber(ARGV[3])

if redis.call('EXISTS', KEYS[2]) == 1 then
  local used = tonumber(redis.call('GET', KEYS[1]) or '0')
  local remaining = limit - used
  if remaining < 0 then remaining = 0 end
  return {'DEDUP', tostring(used), tostring(remaining)}
end

local used = tonumber(redis.call('GET', KEYS[1]) or '0')
if used >= limit then
  return {'QUOTA_EXCEEDED', tostring(used), '0'}
end

used = redis.call('INCR', KEYS[1])
if quotaTTL < 1 then quotaTTL = 1 end
redis.call('EXPIRE', KEYS[1], quotaTTL)
redis.call('SET', KEYS[2], '1', 'EX', dedupTTL)
local remaining = limit - used
if remaining < 0 then remaining = 0 end
return {'ALLOWED', tostring(used), tostring(remaining)}
`
