<?php
class Tools
{
    public static function xssFilter($val) {
        return htmlspecialchars($val, ENT_QUOTES);
    }

    public static function debug() {
        $args = func_get_args();
        header('content-type: text/html;charset=utf-8');
        echo "<pre>";
        foreach($args as $v) {
            print_r($v);
            echo "\n";
        }
        echo "</pre>";
        exit;

    }

    public static function getIp() {
        return $_SERVER["REMOTE_ADDR"];
    }

    public static function getTtl($expire, $etype) {
        if ($etype == 0) { // 永不过期
            return 0;
        }

        if ($etype == 1) { // 绝对时间过期
            return $expire - time();
        }

        if ($etype == 2) { // 相对时间有效
            return $expire;
        }

        if ($etype == 3) { // 特殊的相对时间
            $time = mktime(23, 59, 59, date("m"), date("d"), date("Y"));
            return $time + $expire;
        }

        return -1;
    }

    public static function getPdft() {
        if (isset($_COOKIE["pdft"])) {
            return self::xssFilter($_COOKIE["pdft"]);
        }
        return "";
    }

    // 计算概率
    // 1 <= $pr <= $pool
    public static function calcProbability($pr, $pool = 1000) {
        return rand(1, $pool) <= $pr;
    }

}
