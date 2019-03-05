module('common', package.seeall)

local function var_dump1(tb)
    for k,v in pairs(tb) do
        if ngx then
            ngx.say( k..k)
        else
            print(k,v)
        end
    end
end

local function var_dump(val)
    level = level or 0
    local space = 1
    if type(val) == 'table' then
        for i, j in pairs(val) do
            ngx.say(string.rep('\t', space*level) .. i .. ' = {')
            var_dump(j, level + 1)
            ngx.say(string.rep('\t', space*level) .. '}')
        end
    else
        ngx.say(string.rep('\t', space*level) .. tostring(val))
    end
end


local function explode(szFullString, szSeparator)

    local nFindStartIndex = 1
    local nSplitIndex = 1
    local nSplitArray = {}
    while true do
        local nFindLastIndex = string.find(szFullString, szSeparator, nFindStartIndex)
        if not nFindLastIndex then
            nSplitArray[nSplitIndex] = string.sub(szFullString, nFindStartIndex, string.len(szFullString))
            break
        end
        nSplitArray[nSplitIndex] = string.sub(szFullString, nFindStartIndex, nFindLastIndex - 1)
        nFindStartIndex = nFindLastIndex + string.len(szSeparator)
        nSplitIndex = nSplitIndex + 1
    end
    return nSplitArray

end

local function http_build_query(data, prefix, sep, _key)
    local ret = {}
    local prefix = prefix or ''
    local sep = sep or '&'
    local _key = _key or ''

    for k,v in pairs(data) do
        if (type(k) == "number" and prefix ~= '') then
            k = ngx.escape_uri(prefix .. k)
        end
        if (_key ~= '' or _key == 0) then
            k = ngx.escape_uri(("%s[%s]"):format(_key, k))
        end
        if (type(v) == 'table') then
            table.insert(ret, http_build_query(v, '', sep, k))
        else
            table.insert(ret, ("%s=%s"):format(k, ngx.escape_uri(v)))
        end
    end
    return table.concat(ret, sep)
end

local function parse_str(str)
    local tb = {}
    local pos = 1
    local s = string.find(str,"&",pos)
    while true do
        local kv = string.sub(str,1,s-1)
        local k,v = string.match(kv,"([^=]*)=([^=]*)")
        tb[k] = v
        pos   = s + 1
        str   = string.sub(str,pos)
        s     = string.find(str,"&",1)
        if s == nil then
            k,v = string.match(str,"([^=]*)=([^=]*)")
            tb[k] = v
            break
        end
    end
    return tb
end
local function Trim(str)
    if type(str) ~= 'string' then
        return ''
    end
    local s = string.gsub(str, "^%s*(.-)%s*$", "%1")
    return s
end

local function rawurldecode(s)
     s = string.gsub(s, '%%(%x%x)', function(h) return string.char(tonumber(h, 16)) end)
     return s
end
local function compare(str1,str2)

    local len1  = string.len(str1)
    local len2  = string.len(str2)
    var_dump(len1)
    var_dump(len2)
    ngx.say(len1 == len2 )
    for i = 1 , len1 do
        local char1 = string.sub(str1,i,i)
        local char2 = string.sub(str2,i,i)
        ngx.say( char1 == char2 )
        ngx.say(char1, char2)
    end
end

local function getClientIp()

    local headers = ngx.req.get_headers()
    local IP = ngx.var.http_x_forwarded_for

    -- ngx.log(ngx.ERR, 'http_remoteip='..tostring(ngx.var.http_remoteip)..' remote_addr:'..tostring(ngx.var.remote_addr)..' http_x_forwarded_for:'..tostring(ngx.var.http_x_forwarded_for))

    if IP == nil then
        IP  = ngx.var.x_forwarded_for
    end
    if IP == nil then
        IP  = ngx.var.http_client_ip
    end
    if IP == nil then
        IP  = headers["x-forwarded-for"]
    end
    if IP == nil then
        IP  = headers["remoteip"]
    end
    if IP == nil then
        IP  = ngx.var.remote_addr
    end

    if IP == nil then
        IP  = "unknown"
    end
    return IP
end

local function resolveDomain(domain, nameservers)
    if nameservers == nil then
        return
    end
    local ENV = require( ngx.var.lua_init )
    local resolver = require('resty.resolver')
    local logger = require('syslog.logkit').logger('resolver')
    local r, err = resolver:new{
        nameservers = nameservers or {"10.120.254.94", {"10.120.254.94", 53} },
        retrans = 5,  -- 5 retransmissions on receive timeout
        timeout = 2000,  -- 2 sec
    }

    if not r then
        logger.error("failed to instantiate the resolver: "..err)
        return
    end

    local answers, err = r:query(domain)
    if not answers then
        logger.error("failed to query the DNS server: "..err)
        return
    end

    if answers.errcode then
        logger.error("server returned error code: ".. answers.errcode..": "..answers.errstr)
    end
    local ip = nil
    for i, ans in ipairs(answers) do
        ip = ans.address
        logger.log_level('info')
        logger.info("nameservers:"..nameservers[1][1].." "..ans.name.. " ".. (ans.address or ans.cname).." type:"..ans.type.." class:"..ans.class.." ttl:"..ans.ttl)
    end
    return ip

end

return {
    rawurldecode = rawurldecode,
    explode      = explode,
    parse_str    = parse_str,
    getClientIp  = getClientIp,
    var_dump     = var_dump,
    var_dump1    = var_dump1,
    Trim         = Trim,
    compare      = compare,
    http_build_query = http_build_query,
    resolveDomain=resolveDomain,
}
