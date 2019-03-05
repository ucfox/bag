<?php

Yaf_Loader::import("/home/q/php/gsdk_base/guuid.php");
class GuuidModel
{

    private static $ins;

    public function __construct() {
    }

    public static function ins() {
        if(!(self::$ins instanceof self)){
            self::$ins = new self();
        }
        return self::$ins;
    }

    public function get() {
        $id = GuuidClient::ins()->get();
        XLogKit::logger('guuid')->info("[guuid model] id:{$id}");
        return $id;
    }
}
