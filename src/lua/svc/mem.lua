--[[
--
--   mem.lua
--
--   一个服务用来完成mem连接
--
--   参数：host ，port
--
--
--   返回值1(red) bool
--
--   返回值2(err) 出错的错误信息(string)  没有错误就返回 nil
--
--   usage：
--
--   local red, errmsg = require( LIB_ROOT .. '/svc/mem', host,port)
--
--]]

local LIB_ROOT = ngx.var.lua_lib_root
local memcached,mem = require "resty.memcached" ,{}

function mem:connect(host, port)
    local ENV = require( ngx.var.lua_init )
    host  = host or ENV['IDENTITY_MEMC']
    port  = port or ENV['IDENTITY_PORT']
    local mem = memcached:new()
    mem:set_timeout(500) -- 0.5 sec
    local ok, err = mem:connect(host, port)
    if not ok then
        return false, err
    end
    self.mem = mem
    return true, 'mem ok'
end

function mem:get(key)
    local mem = self.mem
    local res, flag, err = mem:get(key)
    return res, flag, err
end

function mem:set(key,v,exp)
    local mem = self.mem
    local ok, err = mem:set(key,v,exp)
    if ok == 1 then
        return true, err
    end
    return false, err
end

function mem:incr(key,n)
    local mem = self.mem
    local ok, err = mem:incr(key,n)
    if not ok then
        return 0, err
    end
    return ok, err
end

function mem:close()
    local mem = self.mem
    local ok, err = mem:set_keepalive(10000, 100)
    if not ok then mem:close() end
end

function mem:new (o)  --注意，此处使用冒号，可以免写self关键字；
    o = o or {}  -- create object if user does not provide one
    setmetatable(o, {__index = self})
    return o
end

return mem
