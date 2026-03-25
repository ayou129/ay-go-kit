-- ARGV: [project, scene, accessExpire]
-- KEYS: [access_token, refresh_token]

local prefix = ARGV[1]..'_auth_'..ARGV[2]..'_'
local access_expire = tonumber(ARGV[3])
local access_token = KEYS[1]

local token_session_id = redis.call('GET', prefix..'access_token:'..access_token)
if not token_session_id then
    return {-2}
end

local session_info = redis.call('HGETALL', prefix..'token_session:'..token_session_id)
if #session_info == 0 then
    return {-1}
end

redis.call('EXPIRE', prefix..'access_token:'..access_token, access_expire)
redis.call('HSET', prefix..'token_session:'..token_session_id, 'last_access', redis.call('TIME')[1])

return session_info
