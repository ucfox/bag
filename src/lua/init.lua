-- init script
--
-- load env var from $PRJ_HOME/run/$APP_SYS/env.lua

local function loadEnv()

    local f = assert(io.open(ngx.var.lua_env_path, "r"))
    local env = {}

    for line in io.lines(ngx.var.lua_env_path) do
        local l = require(ngx.var.lua_lib_root .. '/common').explode(line, '=')
        if l[1] ~= '_' and l[1] ~= 'SHELL' and ngx.re.match(l[1], '^[0-9A-Z_-]+$') then
            env[l[1]] = l[2]
        end
    end

    return env
end

local env = loadEnv()

return env
