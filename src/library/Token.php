<?php
Class Token {
    private static $token_key = '84D988A6-79F3-3975-01FE-BE29AEE5360F';
    private static $mtoken_key = '84D988A6-79F3-3975-01FE-PANDAEMOBILE';
    const TTL = 43200; // 3600 * 12, 12小时过期

    public static function checkPc($uid, $token) {
         $s_token = md5($uid.'|'.self::$token_key);
         return $token === $s_token;
    }

    public static function checkMb($uid, $token, $ttl) {
        if(time() - $ttl > self::TTL){
            return false;
        }
        $s_token = md5($uid.'||'.$ttl.'||'.self::$mtoken_key);
        return $token === $s_token;
    }
}

