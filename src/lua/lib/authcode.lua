local CHAT_CONNECT = '84D9CHAT-79F3-3390-01CN-BE2PANDACONN'
local CHAT_HOMERRM = '84D9CHAT-79F3-3390-01CN-BEPANDAHOMER'
local md5 = ngx.md5

local function getAuthCodeChat(rid, appid, authType, ts)
    if not ts or ts == '' then
        ts = math.floor(ngx.now()*1000)
    end
    return md5(rid..'||'..appid..'||'..ts..'||'..authType..'||'..CHAT_CONNECT)
end

local function getChatroomSign(rid, roomid, ts)
    if not ts or ts == '' then
        ts = math.floor(ngx.now()*1000)
    end
    return md5(rid..'||'..roomid..'||'..ts..'||'..CHAT_HOMERRM)
end

return {
    getAuthCodeChat = getAuthCodeChat,
    getChatroomSign = getChatroomSign
}
