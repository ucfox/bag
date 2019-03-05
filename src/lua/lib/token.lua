local token_key = "84D988A6-79F3-3975-01FE-BE29AEE5360F"
local mtoken_key = "84D988A6-79F3-3975-01FE-PANDAEMOBILE"
local rolltoken_key = "pandaroll"

local md5 = ngx.md5

local function get(rid)
    if not rid then return false end
    rid = tostring(rid)
    return md5(rid .. '|' .. token_key)
end

local function check(rid, token)
    if not rid or not token then return false end
    rid, token = tostring(rid), tostring(token)
    local s_token = md5(rid .. '|' .. token_key)
    return token == s_token
end

local function mcheck(rid, token,ttl)
    if not rid or not token or not ttl then return false end
    rid, token, ttl = tostring(rid), tostring(token),tonumber(ttl)
    if os.time()-ttl > 43200 then
        return false
    end
    local s_token = md5(rid .. '||' .. ttl .. '||' .. mtoken_key)
    return token == s_token
end

local function checkrolltoken(roomid,time,token)
    if not roomid or not time or not token then return false end
    roomid, time,token = tostring(roomid), tostring(time), tostring(token)
    local s_token = md5(roomid .. '|' .. time .. '|'.. rolltoken_key)
    -- ngx.say(s_token)
    return token == s_token
end

return {
    get   = get,
    check = check,
    mcheck = mcheck,
    checkrolltoken = checkrolltoken ,
}
