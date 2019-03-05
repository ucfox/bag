local LIB_ROOT = ngx.var.lua_lib_root
local args = ngx.req.get_uri_args()
local tools,common = require(LIB_ROOT..'/lib/tools'),require(LIB_ROOT..'/common')

local data  = require('gate/ruc').auth()
if data['errno'] ~= 0 or type(data['data']) ~= "table" or data['data']['rid'] == nil  then
    ngx.say(tools.makeJson('', 200, "账号异常，请重新登录"))
    return
end

local ibagapi = require(LIB_ROOT..'/svc/ibagapi')
local res = {
    total = "45",
    use = "0",
}

local ret = ibagapi.bagNum('pandaren', data['data']['rid'])
ret = tools.cjson.decode(ret) or {}
if ret.errno == 0 and type(ret.data) == "table" then
    ngx.say(tools.makeJson(ret.data, ret.errno, ret.errmsg))
else
    ngx.say(tools.makeJson(res, 0, "success"))
end
ngx.exit(200)
