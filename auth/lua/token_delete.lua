-- ARGV: [prefix, scene]
-- KEYS: [userId]

local prefix = ARGV[1]..'#'..ARGV[2]..'#'
local user_id = KEYS[1]

local user_key = prefix..'user:'..user_id
local session_id = redis.call('GET', user_key)
if not session_id then
    return {1}
end

local session_info = redis.call('HGETALL', prefix..'token_session:'..session_id)
if #session_info > 0 then
    local access_token
    local refresh_token
    for i = 1, #session_info, 2 do
        if session_info[i] == 'access_token' then
            access_token = session_info[i + 1]
        elseif session_info[i] == 'refresh_token' then
            refresh_token = session_info[i + 1]
        end
    end
    if access_token then
        redis.call('DEL', prefix..'access_token:'..access_token)
    end
    if refresh_token then
        redis.call('DEL', prefix..'refresh_token:'..refresh_token)
    end
end

redis.call('DEL', prefix..'token_session:'..session_id)
redis.call('DEL', user_key)

return {1}
