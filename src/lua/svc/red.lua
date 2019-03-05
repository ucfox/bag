--[[
--
--   red.lua
--
--   一个服务用来完成redis连接
--
--   参数：host ，port，pass(可选)
--
--
--   返回值1(red) bool
--
--   返回值2(err) 出错的错误信息(string)  没有错误就返回 nil
--
--   usage：
--
--   local red, errmsg = require( LIB_ROOT .. '/svc/red', host,port,pass)
--
--]]

local LIB_ROOT = ngx.var.lua_lib_root
local redis,tools,red = require "resty.redis",require(LIB_ROOT..'/lib/tools'),{}

function red:getConf(shost, port, pass,mhost)

    local ENV = require( ngx.var.lua_init )

    local port = tostring(port or ENV['REDIS_PORT_ROLL'])
    local pass = (pass or ENV['REDIS_PWD_' .. port])
    local shost = tostring(shost or ENV['REDIS_SLAVE_' .. port])
    local mhost = tostring(mhost or ENV['REDIS_HOST_' .. port])

    if shost == 'nil' or port == 'nil' or mhost == 'nil'then
        ngx.log(ngx.ERR, 'roll|redis conf error: host|port is nil'..' shost:'..tostring(shost)..' mhost:'..tostring(mhost)..' port:'..tostring(port)..' pass:'..tostring(pass))
        ngx.say(tools.makeJson('',1000,'conf nil'))
        ngx.exit(ngx.HTTP_OK)
    end

    self.redisconf = {
        slave = {
            host = shost,
            port = port,
            pass = pass
        },
        master = {
            host = mhost,
            port = port,
            pass = pass
        }
    }
end

function red:slave()
    if self._slave then
        return self._slave
    end

    local red = redis:new()
    local conf   = self.redisconf['slave']

    local port = tostring(conf['port'])
    local host = tostring(conf['host'])
    local pass = conf['pass']

    red:set_timeout(1000) -- 0.5 sec
    local ok, err = red:connect(host, port)

    if not ok then
        ngx.log(ngx.ERR, 'roll|redis slave error: '..tostring(err)..' host:'..tostring(host)..' port:'..tostring(port)..' pass:'..tostring(pass))
        ngx.say(tools.makeJson('',1000,'redis con err'))
        ngx.exit(ngx.HTTP_OK)
    end

    local times, err = red:get_reused_times()
     -- ngx.say(tostring(ok))
    if times == nil or times == 0 then
        if pass ~= nil and pass ~='' and pass ~= ngx.null  then
            local res, err = red:auth(pass)
            if err then
                ngx.log(ngx.ERR, 'roll|redis slave auth error: '..tostring(err)..' host:'..tostring(host)..' port:'..tostring(port)..' pass:'..tostring(pass))
                ngx.say(tools.makeJson('',1000,'redis auth err'))
                ngx.exit(ngx.HTTP_OK)
            end
        end
    end

    self._slave = red
    return self._slave
end

function red:master(host, port, pass)

    if self._master then
        return self._master
    end

    local red  = redis:new()
    local conf = self.redisconf['master']

    local port = tostring(conf['port'])
    local host = tostring(conf['host'])
    local pass = conf['pass']

    red:set_timeout(1000) -- 0.5 sec
    local ok, err = red:connect(host, port)

    if not ok then
        ngx.log(ngx.ERR, 'roll|redis master error: '..tostring(err)..' host:'..tostring(host)..' port:'..tostring(port)..' pass:'..tostring(pass))
        ngx.say(tools.makeJson('',1000,'redis con err'))
        ngx.exit(ngx.HTTP_OK)
    end

    local times, err = red:get_reused_times()
    -- ngx.say(tostring(ok))
    if times == nil or times == 0 then
        if pass ~= nil and pass ~='' and pass ~= ngx.null  then
            local res, err = red:auth(pass)
            if err then
                ngx.log(ngx.ERR, 'roll|redis master auth error: '..tostring(err)..' host:'..tostring(host)..' port:'..tostring(port)..' pass:'..tostring(pass))
                ngx.say(tools.makeJson('',1000,'redis auth err'))
                ngx.exit(ngx.HTTP_OK)
            end
        end
    end

    self._master = red
    return self._master
end

function red:get(key)
    local red = self:slave()
    local res, err = red:get(key)
    return res, err
end

function red:mget(keys)
    local red = self:slave()
    local res, err = red:mget(unpack(keys))
    return res, err
end

function red:hgetall(key)
    local red = self:slave()
    local res, err = red:hgetall(key)
    if err ~= nil then
        ngx.log(ngx.ERR, 'redis hgetall error: ' .. tostring(err) .. ' key:' .. tostring(key))
    else
        local kvpairs = {}
        for i=1, #(res), 2 do
            kvpairs[res[i]] = res[i + 1]
        end
        res = kvpairs
    end
    return res, err
end

function red:hincrby(key,field,incr)
    local red = self:master()
    local res, err = red:hincrby(key,field,incr)
    return res, err
end

function red:hget(key,field)
    local red = self:slave()
    local res, err = red:hget(key,field)
    return res, err
end

function red:eval(cmd, keys)
    local red = self:master()
    local res, err = red:eval(cmd, #keys, unpack(keys))
    if err then
        ngx.log(ngx.ERR, "redis eval err: " .. tostring(err))
    end
    return res, err
end

function red:evalslave(cmd, keys)
    local red = self:slave()
    local res, err = red:eval(cmd, #keys, unpack(keys))
    if err then
        ngx.log(ngx.ERR, "redis eval slave err: " .. tostring(err))
    end
    return res, err
end

function red:smembers(key)
    local red = self:slave()
    local res, err = red:smembers(key)
    if err ~= nil then
        ngx.log(ngx.ERR, 'redis smembers error: ' .. tostring(err) .. ' key:' .. tostring(key))
    end
    return res, err
end

function red:sismember(key,member)
    local red = self:slave()
    local res, err = red:sismember(key,member)
    if err then
        ngx.log(ngx.ERR, 'redis sismember error: ' .. tostring(err) .. ' key:' .. tostring(key))
    end
    return res, err
end

function red:set(key,v)
    local red = self:master()
    local ok, err = red:set(key,v)
    if ok == 1 then
        return true, err
    end
    return false, err
end

function red:exists(key)
    local red = self:slave()
    local ok, err = red:exists(key)
    if ok == 1 then
        return true, 'exists'
    end
    return false, err
end

function red:expire(key,n)
    local red = self:master()
    local ok, err = red:expire(key,n)
    if not ok then
        return false, err
    end
    return true, err
end

function red:incrby(key,n)
    local red = self:master()
    local ok, err = red:incrby(key,n)
    if not ok then
        return 0, err
    end
    return ok, err
end

function red:decrby(key,n)
    local red = self:master()
    local ok, err = red:decrby(key,n)
    if not ok then
        return -1, err
    end
    return ok, err
end

function red:hincrby(key,field,n)
    local red = self:master()
    local ok, err = red:hincrby(key,field,n)
    if not ok then
        return 0, err
    end
    return ok, err
end

function red:sadd(key,v)
    local red = self:master()
    local res, err = red:sadd(key,v)
    if err ~= nil then
        ngx.log(ngx.ERR, 'redis sadd error: '..tostring(err)..' host:'..tostring(self.tHost)..' port:'..tostring(self.tPort))
    end
    return res, err
end

function red:lpop(key)
    local red = self:master()
    local res, err = red:lpop(key)
    if err ~= nil then
        ngx.log(ngx.ERR, 'redis lpop error: '..tostring(err)..' host:'..tostring(self.tHost)..' port:'..tostring(self.tPort))
    end
    return res, err
end

function red:scard(key)
    local red = self:slave()
    local res, err = red:scard(key)
    return res, err
end

function red:select(key)
    local red = self:slave()
    local res, err = red:select(key)
    if not res then
        return false, err
    end
    return true, res
end

function red:close()
    local red = self:slave()
    local ok, err = red:set_keepalive(10000, 100)
    if not ok then red:close() end

    local red = self:master()
    local ok, err = red:set_keepalive(10000, 100)
    if not ok then red:close() end
end

function red:new (o)  --注意，此处使用冒号，可以免写self关键字；
    local o = o or {}  -- create object if user does not provide one
    setmetatable(o, {__index = self})
    self:getConf()
    return o
end

return red
