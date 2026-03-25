-- ARGV: [project, scene, tokenExpire, refreshExpire]
-- KEYS: [userId, newSessionId, newAccessToken, newRefreshToken]

local prefix = ARGV[1]..'_auth_'..ARGV[2]..'_'
local token_expire = tonumber(ARGV[3])
local refresh_expire = tonumber(ARGV[4])
local user_id = KEYS[1]
local new_session_id = KEYS[2]
local new_access_token = KEYS[3]
local new_refresh_token = KEYS[4]

if redis.call('EXISTS', prefix..'token_session:'..new_session_id) == 1 then
    return {-1}
end
if redis.call('EXISTS', prefix..'access_token:'..new_access_token) == 1 then
    return {-2}
end
if redis.call('EXISTS', prefix..'refresh_token:'..new_refresh_token) == 1 then
    return {-3}
end

local user_key = prefix..'user:'..user_id
local stored_session_id = redis.call('GET', user_key)
if stored_session_id then
    local old_session = redis.call('HGETALL', prefix..'token_session:'..stored_session_id)
    if #old_session > 0 then
        local stored_access_token
        local stored_refresh_token
        for i = 1, #old_session, 2 do
            if old_session[i] == 'access_token' then
                stored_access_token = old_session[i + 1]
            elseif old_session[i] == 'refresh_token' then
                stored_refresh_token = old_session[i + 1]
            end
        end
        if stored_access_token then
            redis.call('DEL', prefix..'access_token:'..stored_access_token)
        end
        if stored_refresh_token then
            redis.call('DEL', prefix..'refresh_token:'..stored_refresh_token)
        end
    end
    redis.call('DEL', prefix..'token_session:'..stored_session_id)
    redis.call('DEL', prefix..'user:'..user_id)
end

redis.call('SETEX', prefix..'access_token:'..new_access_token, token_expire, new_session_id)
redis.call('HMSET', prefix..'refresh_token:'..new_refresh_token,
    'user_id', user_id,
    'access_token', new_access_token)
redis.call('EXPIRE', prefix..'refresh_token:'..new_refresh_token, refresh_expire)
redis.call('HMSET', prefix..'token_session:'..new_session_id,
    'user_id', user_id,
    'access_token', new_access_token,
    'refresh_token', new_refresh_token,
    'created_at', redis.call('TIME')[1])
redis.call('EXPIRE', prefix..'token_session:'..new_session_id, refresh_expire)
redis.call('SETEX', prefix..'user:'..user_id, refresh_expire, new_session_id)

return {1}
