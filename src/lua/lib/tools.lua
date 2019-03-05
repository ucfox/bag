local cjson = require "json"

if cjson.encode_empty_table_as_object then
    cjson.encode_empty_table_as_object(false) -- 空的table默认为array
end

--{{{ makeJson(data,errno,errmsg)
local function makeJson(data,errno,errmsg)
    local allres  = {}
    allres['errno'] = errno
    allres['errmsg'] = errmsg
    allres['data'] = data
    return cjson.encode(allres)
end
--}}}
--{{{ explode(delimiter, str, ...)
local function explode(delimiter, str, ...)
    if delimiter == '' then
        return str
    end
    local pos,arr = 0,{}
    local limit = ...
    local num = 0
    for st,sp in function() return string.find(str,delimiter,pos,true) end do
        table.insert(arr,string.sub(str,pos,sp-1))
        pos = sp + 1
        num = num + 1
        if limit and num == limit then
            break
        end
    end
    table.insert(arr,string.sub(str,pos))
    return arr
end
--}}}
--{{{ function empty(var)
local function empty(var)
    return var == nil or var == '' or var == {} or tonumber(var) == 0 or var == ngx.null or not var
end
--}}}
return {
    makeJson = makeJson,
    explode = explode,
    empty = empty,
    cjson = cjson
}
