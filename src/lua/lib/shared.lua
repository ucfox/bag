
local function get(key, dbname)

    local key = key

    if key == nil or key == "" then
        key = ngx.var.arg_device
    end

    local devicedb = dbname and ngx.shared[dbname] or ngx.shared.online_pandaren_bandb
    local res = devicedb:get(key)
    -- ngx.say(res)
    return res
end

local function update(file, dbname)

    local devicedb = dbname and ngx.shared[dbname] or ngx.shared.online_pandaren_bandb

    if file == nil or file == "" then
        file = ngx.var.shared_file
        -- file 传文件json 处理
        --
        --
    end

    for device_type,device_file in pairs(file) do

        local file,err = io.open(device_file)

        if file then
            local device_rule,predata = io.open( device_file , "r"):read("*all"),get(device_type, dbname)

            device_rule = string.gsub(device_rule,"\n","|")
            predata = predata or {}

            local prelen,updatelen =  #predata ,#device_rule

            if updatelen ==  0 then
                ngx.say("file len is 0 :: "..device_file)
                return
            elseif prelen/updatelen > 1.5 then
                ngx.say("update less then 50% ,prelen: ".. prelen .."updatelen: "..updatelen.." file is "..device_file)
                return
            end
            -- other check
            --

            devicedb:set(device_type,device_rule)

            if ngx.var.arg_showlog == "showlog" then
                ngx.print( device_type..": ".. prelen .. " --> " .. updatelen .. "; " )
            end
        else
            ngx.say("file error :: "..device_file..err)
        end
    end

end

return {
    get    = get,
    update =  update ,
}
