local function htmlspecialchars(str)
    local str, n, err = ngx.re.gsub(str, "&", "&amp;")
    str, n, err = ngx.re.gsub(str, "<", "&lt;")
    str, n, err = ngx.re.gsub(str, ">", "&gt;")
    str, n, err = ngx.re.gsub(str, '"', "&quot;")
    str, n, err = ngx.re.gsub(str, "'", "&apos;")
    return str
end
local function getClientIp()
    local IP = ngx.req.get_headers()["X-Real-IP"]

    if IP == nil then
        IP  = ngx.var.remote_addr
    end

    return IP
end

local function getDictByffi(dickey ,str)

    -- pylon_smasher 共享内存要先初始化才能使用，否则会有未知C++ Exception抛出
    local ffi = require("ffi")
    ffi.cdef[[
    void shared_dict_create(const char*  proc_space, int msize,int dynload,int loadpgs);
    void shared_dict_using(const char*  proc_space );
    void shared_dict_remove(const char*  proc_space);
    void shared_dict_data(const char * data_file,const char* key_prefix , const char* data_prefix);
    int  shared_dict_find(const char* cls, char * buf , int buf_len);
    int  shared_dict_count();
    ]]

    local pylon = ffi.load("libpylon_smasher_2_0_0")

    pylon.shared_dict_using(dickey)

    local n, res = 102400, nil
    local buf = ffi.new("char[?]", n)

    pylon.shared_dict_find(str, buf, n)

    res = ffi.string(buf)
    -- ngx.log(ngx.ERR, ' finding ' .. str .. ' in ' .. dickey .. ': ' .. res)
    return res
end

local function getSharedDict( key )
    local LIB_ROOT, db = ngx.var.lua_lib_root, ngx.var.ban_db
    -- ngx.log(ngx.ERR, ngx.var.ban_db)
    local shared = require( LIB_ROOT .. "/lib/shared")
    return shared.get( key, db )
end

local function addslashes(str)
    --local str, n, err = ngx.re.gsub(str, "\\", '\\\\')
    local str, n, err = ngx.re.gsub(str, '"', '\\"')
    str, n, err = ngx.re.gsub(str, "'", "\\'")
    return str
end
local function utfstrlen(str)
    local len = #str;
    local left = len;
    local cnt = 0;
    local arr={0,0xc0,0xe0,0xf0,0xf8,0xfc};
    while left ~= 0 do
        local tmp=string.byte(str,-left);
        local i=#arr;
        while arr[i] do
            if tmp>=arr[i] then left=left-i;break;end
            i=i-1;
        end
        cnt=cnt+1;
    end
    return cnt;
end
local function detectIP(str)
    local res = ngx.re.match(str,'((([0-9]?[0-9])|(1[0-9]{2})|(2[0-4][0-9])|(25[0-5]))\\.){3}(([0-9]?[0-9])|(1[0-9]{2})|(2[0-4][0-9])|(25[0-5]))',"isjo")
    return res
end

local function detectURL(str)
    local pat = "http[s]?:\\/\\/(([0-9]{1,3}\\.){3}[0-9]{1,3}|([0-9a-z_!~*\\'()-]+\\.)*([0-9a-z][0-9a-z-]{0,61})?[0-9a-z]\\.[a-z]{2,6})(:[0-9]{1,4})?((\\/\\?)|(\\/[0-9a-zA-Z_!~\\*\\'\\(\\)\\.;\\?:@&=\\+\\$,%#-\\/]*)?)"
    local res = ngx.re.match(str,pat,"isjo")
    return res
end

local function detectClientIP()
    local ip, util, ips ,iputils,dips = getClientIp() ,require('utils'), {} ,require("resty.iputils"),getSharedDict("forbidips")

    if  dips ~="" and dips ~= nil then

        ips = util.split(getSharedDict("forbidips"),"|")
        ips = iputils.parse_cidrs(ips)

        if  iputils.ip_in_cidrs(ip, ips) then
            ngx.log(ngx.ERR, "ip:" .. ip .. " is baned.")
            return true
        end
    end

    return false
end

local function forbidden(str)

    local util, words = require('utils'), getSharedDict("forbidwords") or "草榴|万庆良|钟绍军|郭伯雄|贾庆L|贾庆林|谷俊山|宋祖英|宋祖Y|宋ZU英|姘头|刘源|刘乐飞|军委深改组|徐才厚|刘氏父子|梁光烈|买官卖爵|中宣部|爱液横流|安门事|八九民|八九学|八九政治|把邓小平|把学生整|办本科|办理本科|办理各种|办理票据|办理文凭|办理真实|办理证书|办理资格|办文凭|办怔|办证|冰淫传|苍蝇水|藏春阁|藏獨|操了嫂|操嫂子|插屁屁|城管灭|惩公安|惩贪难|春水横溢|催眠水|催情粉|催情药|催情藥|打标语|打错门|打飞机专|大鸡巴|大雞巴|大奶子|大肉棒|代办发票|代办各|代办文|代办学|代办制|代表烦|代理发票|代理票据|代您考|代您考|代写毕|代写论|戴海静|到花心|丁子霖|顶花心|东北独立|独立台湾|法车仑|法伦功|法轮|法轮佛|法维权|法一轮|法正乾|范燕琼|方迷香|防电子眼|防身药水|仿真证|夫妻交换|妇销魂|肛交|肛门是邻|港澳博球|搞媛交|共狗|和狗交|和狗性|胡江内斗|胡紧套|胡錦濤|胡适眼|胡耀邦|华国锋|换妻|激情电|激情短|激情妹|激情炮|急需嫖|挤乳汁|擠乳汁|奸成瘾|兼职上门|监听器|监听王|简易炸|江太上|江系人|江贼民|疆獨|蒋彦永|叫自慰|揭贪难|姐包夜|姐服务|姐兼职|姐上门|金钟气|精子射在|巨乳|绝食声|黎阳平|李洪志|李咏曰|理是影帝|聊斋艳|六四事|六月联盟|乱奸|乱伦类|乱伦小|亂倫|乱伦|轮功|轮手枪|裸聊网|裸舞视|麻果配|麻果丸|麻将透|麻醉狗|麻醉枪|麻醉槍|麻醉藥|蟆叫专家|卖发票|卖银行卡|卖自考|忙爱国|毛一鲜|妹按摩|妹上门|门按摩|门保健|門服務|蒙汗药|迷幻型|迷幻药|迷幻藥|迷昏口|迷昏药|迷魂香|迷魂药|迷魂藥|迷奸药|迷情水|迷情药|迷藥|谜奸药|蜜穴|内射|嫩穴|嫩阴|妞上门|女被人家搞|女任职名|女上门|女優|拍肩神药|拍肩型|牌分析|牌技网|炮的小蜜|陪考枪|喷尿|嫖俄罗|嫖鸡|平惨案|奇淫散|强权政府|清純壆|情聊天室|情妹妹|情视频|情自拍|请集会|群奸暴|群起抗暴|群体性事|任于斯国|柔胸粉|肉洞|肉棍|乳交|三级片|骚妇|骚浪|骚穴|骚嘴|色电影|色妹妹|色视频|色小说|烧公安局|失身水|失意药|兽交|售步枪|售纯度|售单管|售弹簧刀|售防身|售狗子|售虎头|售火药|售假币|售健卫|售军用|售猎枪|售氯胺|售麻醉|售冒名|售枪支|售热武|售三棱|售手枪|售五四|售信用|售一元硬|售子弹|售左轮|丝护士|丝情侣|丝袜保|丝袜恋|丝袜美|丝袜妹|丝袜网|丝足按|死要见毛|速代办|速取证|探测狗|涛共产|特工资|特码|特上门|体透视镜|推油按|脱衣艳|瓦斯手|袜按摩|外透视镜|湾版假|万能钥匙|王立军|王益案|维汉员|维权基|维权人|维权谈|委坐船|温家堡|温切斯特|温影帝|溫家寶|瘟加饱|瘟假饱|乌蝇水|无码专|午夜极|雾型迷|西藏限|希脏|习进平|习晋平|席复活|小穴|校骚乱|协晃悠|写两会|新疆叛|新疆限|行长王益|形透视镜|型手枪|性爱日|性福情|性感少|性推广歌|胸主席|徐玉元|学骚乱|学位證|严晓玲|颜射|要射精了|液体炸|陰唇|陰道|陰戶|淫魔舞|淫情女|淫肉|淫騷妹|淫兽|淫兽学|淫水|淫穴|婴儿命|咏妓|用手枪|幼齿类|与狗性|园发生砍|园砍杀|园凶杀|园血案|找援交|赵紫阳|针刺案|真钱斗地|真钱投注|真善忍|真实文凭|真实资格|证到付款|证件办|证件集团|证生成器|证书办|政府操|政论区|制作证件|着涛哥|自慰用|总会美女|足球玩法|醉钢枪|醉迷药|做爱小|做证件|温家宝|习近平|李克强|江泽民|毛泽东|藏独|视频做|彭丽媛|邓小平|台独"

    words = util.split(words,"|")

    for _,rule in pairs(words) do
        if rule ~="" and ngx.re.match(str,rule,"isjo") then
            return true
        end
    end
    return false
end

local function formatWord(content)

    --- 清洗 content
    local content = content -- 兼容英文 空格 先去掉string.gsub(content,"%s","")
    local lookup = {
        ["零"] = "0",["澪"] = "0",["蕶"] = "0",["O"] = "0",["圈"] = "0",["o"] = "0" ,
        ["一"] = "1",["壹"]="1" ,
        ["二"] = "2",["贰"] = "2",
        ["三"] = "3",["叁"] = "3",["仨"] = "3" ,
        ["四"] = "4",["肆"] = "4",
        ["五"] = "5",["伍"] = "5",
        ["六"] = "6",["陆"] = "6",
        ["七"] = "7",["柒"] = "7",
        ["八"] = "8",["捌"] = "8",["扒"] = "8" ,
        ["九"] = "9",["玖"] = "9",
        ["拾"] = "",["十"] = "" ,
        ["佰"] = "",["百"] = "" ,
        ["仟"] = "",["千"] = "" ,
        ["扣"]  = "q",["釦"] ="q",
        -- other
    }
    -- [:微笑]  表情
    local _n, n, err = ngx.re.gsub(content, "([^a-zA-Z0-9,、，。 !！？:\\[\\]\\.\\x{4e00}-\\x{9fff}\\x{1100}-\\x{11FF}\\x{3130}-\\x{318F}\\x{AC00}-\\x{D7AF}\\x{3040}-\\x{30FF}\\x{31F0}-\\x{31FF}])+", "", "iujo")
    local extra,err = ngx.re.gmatch(_n,"([\\x{4e00}-\\x{9fff}Oo])","iujo")
    local _n_format = _n
    -- ngx.say('_n::'.._n);
    -- ngx.exit(200)
    if extra ~= nil and extra ~= '' then
        while true do
            local m, err = extra()
            if not m then
                break
            end
            if lookup[m[1]] ~= nil and lookup[m[1]] ~= '' then
                _n_format = ngx.re.sub(_n_format,m[1],lookup[m[1]])
            end
        end
    end

    local _n_str, n, err = ngx.re.gsub(_n_format, "([^a-zA-Z0-9])+", "", "isjo")
    local _n_url, n, err = ngx.re.gsub(_n_format, "([^a-zA-Z])+", "", "isjo")
    local _n_chat, n, err = ngx.re.gsub(_n_format, "([^\\x{4e00}-\\x{9fff}])+", "", "iujo")
    --local _n_chat_hy, n, err = ngx.re.gsub(_n_format, "([^\\x{1100}-\\x{11FF}\\x{3130}-\\x{318F}\\x{AC00}-\\x{D7AF}])+", "", "iujo")
    --local _n_chat_ry, n, err = ngx.re.gsub(_n_format, "([^\\x{3040}-\\x{30FF}\\x{31F0}-\\x{31FF}])+", "", "iujo")
    -- 格式化 匹配字符
    local _s,_s_str,_s_url = string.lower(_n),string.lower(_n_str),string.lower(_n_url)

    local data = {
        _all     = _n ,
        _fotall  = _n_format ,
        _lowall  = _s ,
        _lowstr  = _s_str ,
        _lowurl  = _s_url ,
        -- 韩语 日语 敏感词
        --_chathy  = _n_chat_hy,
        --_chatry  = _n_chat_ry,
        _chat    = _n_chat
    }
    return data ,_n
end

local function checkWeight(content)

    local url = require( ngx.var.lua_init )['FORBID_URL']

    if url == nil or url == "" then
        url = "10.138.230.166:9190" -- bjdt
    end

    local params = {
        sKey     = "youxi",
        sToken    = 'youxi',
        content    = content
    }

    ----
    --   预留 开关 跟 阈值 接口
    ---

    local http,util = require "resty.http",require('utils')
    local qs = require( ngx.var.lua_lib_root .. '/common' ).http_build_query(params)
    local ok, code, headers, status, body = http:new():request {
        url = "http://".. url .."/spam/check?"..qs,
        method = "GET", -- POST or GET
    }

    local data,weight = require("json").decode(body),0
    if type(data) == "table" and data['errno'] == 0 and type(data['data']) == 'table' and  data['data']['command'] == -1 then

        local extra,err = ngx.re.gmatch(data['data']['detail']['extra'],"([0-9]+);","i")

        while true do
            local m, err = extra()
            if not m then
                -- no match found (any more)
                break
            end
            -- found a match
            weight = weight + tonumber(m[1])
        end

        --ngx.say( "weigth::"..weight .." _info :" .. data['data']['detail']['extra'] )
    end
    return weight
end

function formatstrlen(str)

    local strlen = #str
    local _,n = str:gsub('[\128-\255]','')
    if tonumber(n) > 0 then
        -- front unicode-16 汉字2个字符  lua是三个字符 在此修正
        strlen = strlen - n/3
    end

    -- ngx.say("str"..strlen)
    return strlen
end

return {
    formatstrlen        =   formatstrlen ,
    htmlspecialchars    =   htmlspecialchars,
    addslashes          =   addslashes,
    utfstrlen           =   utfstrlen,
    detectIP            =   detectIP,
    detectClientIP      =   detectClientIP ,
    detectURL           =   detectURL,
    formatWord          =   formatWord,
    checkWeight         =   checkWeight,
    forbidden           =   forbidden
}
