local LIB_ROOT = ngx.var.lua_lib_root
local CMS_ROOT = '/home/q/cmstpl/'
local function readCmsFile(file_path)
    local file_str = ''
    local file = ''
    pcall(function() file = io.open( CMS_ROOT..file_path, "r") end)
    if file then
        file_str = file:read("*all")
    end
    return file_str
end
return {
    readCmsFile = readCmsFile
}
