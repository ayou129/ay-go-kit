package ratelimit

// luaAllow 判定+记录（原子）
//
// KEYS[1] = 限流 key
// ARGV[1] = 窗口大小（毫秒）
// ARGV[2] = 当前时间（毫秒）
// ARGV[3] = 窗口内最大请求数
// ARGV[4] = 本次请求的唯一成员 ID
//
// 返回: {allowed(0/1), remaining, retry_after_ms, total}
const luaAllow = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local now = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

-- 1. 清理过期条目
redis.call("ZREMRANGEBYSCORE", key, "-inf", now - window)

-- 2. 当前窗口内的请求数
local count = tonumber(redis.call("ZCARD", key))

if count < limit then
    -- 3a. 放行：写入本次请求，刷新 TTL
    redis.call("ZADD", key, now, member)
    redis.call("PEXPIRE", key, window)
    return {1, limit - count - 1, 0, count + 1}
else
    -- 3b. 拒绝：计算最早条目过期时间作为 retry_after
    local oldest = redis.call("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES", "LIMIT", 0, 1)
    local retry_after = 0
    if #oldest >= 2 then
        retry_after = tonumber(oldest[2]) + window - now
        if retry_after < 0 then retry_after = 0 end
    end
    return {0, 0, retry_after, count}
end
`

// luaPeek 查询当前状态（不计入请求，会清理过期条目以保证计数精确；不适用于 readonly replica）
//
// KEYS[1] = 限流 key
// ARGV[1] = 窗口大小（毫秒）
// ARGV[2] = 当前时间（毫秒）
// ARGV[3] = 窗口内最大请求数
//
// 返回: {allowed(0/1), remaining, retry_after_ms, total}
const luaPeek = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local now = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

redis.call("ZREMRANGEBYSCORE", key, "-inf", now - window)

local count = tonumber(redis.call("ZCARD", key))
local remaining = limit - count
if remaining < 0 then remaining = 0 end

local allowed = 1
local retry_after = 0
if count >= limit then
    allowed = 0
    local oldest = redis.call("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES", "LIMIT", 0, 1)
    if #oldest >= 2 then
        retry_after = tonumber(oldest[2]) + window - now
        if retry_after < 0 then retry_after = 0 end
    end
end

return {allowed, remaining, retry_after, count}
`

// luaEntries 读取当前窗口内所有条目（监控/调试）
//
// KEYS[1] = 限流 key
// ARGV[1] = 窗口大小（毫秒）
// ARGV[2] = 当前时间（毫秒）
//
// 返回: {member1, score1, member2, score2, ...}
const luaEntries = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local now = tonumber(ARGV[2])

redis.call("ZREMRANGEBYSCORE", key, "-inf", now - window)
return redis.call("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES")
`

// ---- 拒绝日志 Lua 脚本 ----

// luaDenyRecord 记录一次拒绝（原子：写条目 + 更新排行 + 设 TTL）
//
// KEYS[1] = 月度拒绝日志 key  ({project}_ratelimit_deny_{month}_{key})
// KEYS[2] = 月度排行 key      ({project}_ratelimit_deny_top_{month})
// ARGV[1] = 时间戳（毫秒，作为 score）
// ARGV[2] = member（时间戳:METHOD:path）
// ARGV[3] = 限流 key（作为排行榜 member）
// ARGV[4] = TTL（秒）
const luaDenyRecord = `
local logKey = KEYS[1]
local topKey = KEYS[2]
local score = tonumber(ARGV[1])
local member = ARGV[2]
local rankMember = ARGV[3]
local ttl = tonumber(ARGV[4])

redis.call("ZADD", logKey, score, member)
redis.call("ZINCRBY", topKey, 1, rankMember)

-- 仅在新建时设 TTL（避免每次重置）
if redis.call("TTL", logKey) == -1 then
    redis.call("EXPIRE", logKey, ttl)
end
if redis.call("TTL", topKey) == -1 then
    redis.call("EXPIRE", topKey, ttl)
end

return 1
`

// luaDenyEntries 分页查询拒绝记录（按时间倒序）
//
// KEYS[1] = 月度拒绝日志 key
// ARGV[1] = offset (start index, 0-based)
// ARGV[2] = stop index (inclusive)
//
// 返回: {total, {member1, score1, member2, score2, ...}}
const luaDenyEntries = `
local key = KEYS[1]
local start = tonumber(ARGV[1])
local stop = tonumber(ARGV[2])

local total = redis.call("ZCARD", key)
local items = redis.call("ZREVRANGEBYSCORE", key, "+inf", "-inf", "WITHSCORES", "LIMIT", start, stop - start + 1)
return {total, items}
`

// luaDenyTop 拒绝排行（降序，按 score=拒绝次数排序）
//
// KEYS[1] = 月度排行 key
// ARGV[1] = 返回条数
//
// 返回: {member1, score1, member2, score2, ...}
const luaDenyTop = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
return redis.call("ZREVRANGE", key, 0, limit - 1, "WITHSCORES")
`
