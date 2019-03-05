local LIB_ROOT = ngx.var.lua_lib_root
local ENV,http,common = require( ngx.var.lua_init ),require "resty.http",require(LIB_ROOT .. '/common')
local logger = require('syslog.logkit').logger('lua_ibag_api')
logger.log_level(info)

local function bagNum(app, uid)
    local resolver = require('resolver')
    local ibag_ip = resolver.resolve(ENV.IBAG_HOST)
    if not ibag_ip then
        logger.error('ibag ip is nil')
        return ""
    end
    local params = {
        app     = app,
        uid     = uid,
        _caller = 'bag'
    }
    local ok, code, headers, status, body = http:new():request {
        url     = "http://".. ibag_ip .. "/bag/num?" .. common.http_build_query(params),
        method  = "GET",
        headers = {Host = ENV.IBAG_HOST},
        timeout = 2000,
        port    = 8360,
    }
    return body
end

return {
    bagNum = bagNum
}
