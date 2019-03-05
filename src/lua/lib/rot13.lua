Rot13 = {}

function encrypt(t)
    return (string.gsub(t, "[%a]",function (char)
        local bUpper = (char < 'a')
        local b = string.byte(string.upper(char)) - 65 -- 0 to 25
        b = math.mod(b  + 13, 26)
        if bUpper then
            return string.char(b + 65)
        else
            return string.char(b + 97)
        end
    end
    ))
end

Rot13.encrypt = encrypt
return Rot13




