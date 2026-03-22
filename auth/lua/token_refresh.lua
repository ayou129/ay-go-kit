-- ARGV: [prefix, scene, tokenExpire, refreshExpire]
-- KEYS: [oldAccessToken, oldRefreshToken, newSessionId, newAccessToken, newRefreshToken]

local prefix = ARGV[1]..'#'..ARGV[2]..'#'
local token_expire = tonumber(ARGV[3])
local refresh_expire = tonumber(ARGV[4])
local old_access_token = KEYS[1]
local old_refresh_token = KEYS[2]
local new_session_id = KEYS[3]
local new_access_token = KEYS[4]
local new_refresh_token = KEYS[5]

if redis.call('EXISTS', prefix..'token_session:'..new_session_id) == 1 then
    return {-1}
end
if redis.call('EXISTS', prefix..'access_token:'..new_access_token) == 1 then
    return {-2}
end
if redis.call('EXISTS', prefix..'refresh_token:'..new_refresh_token) == 1 then
    return {-3}
end

local refresh_info = redis.call('HGETALL', prefix..'refresh_token:'..old_refresh_token)
if #refresh_info == 0 then
    return {-4}
end

local user_id
local stored_access_token
for i = 1, #refresh_info, 2 do
    if refresh_info[i] == 'user_id' then
        user_id = refresh_info[i + 1]
    elseif refresh_info[i] == 'access_token' then
        stored_access_token = refresh_info[i + 1]
    end
end
if not user_id then
    return {-5}
end
if stored_access_token ~= old_access_token then
    return {-5}
end

local user_key = prefix..'user:'..user_id
local stored_session_id = redis.call('GET', user_key)
if stored_session_id then
    local old_session = redis.call('HGETALL', prefix..'token_session:'..stored_session_id)
    if #old_session > 0 then
        local sa, sr
        for i = 1, #old_session, 2 do
            if old_session[i] == 'access_token' then sa = old_session[i + 1]
            elseif old_session[i] == 'refresh_token' then sr = old_session[i + 1]
            end
        end
        if sa then redis.call('DEL', prefix..'access_token:'..sa) end
        if sr then redis.call('DEL', prefix..'refresh_token:'..sr) end
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
