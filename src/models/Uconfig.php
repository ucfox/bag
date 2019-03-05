<?php

Yaf_Loader::import("/home/q/php/uconfig_sdk/uconfig_sdk.php");
class UconfigModel extends SdkBase
{
    const APP          = "pandaren";

    private $_caller   = "bag";

    private $_UcClient;

    private static $ins;

    public function __construct()
    {
        $this->_logger   = XLogKit::logger("uconfig_api");
        $this->remoteApp = "uconfig";
        $env = $_SERVER['ENV'];
        $this->_UcClient = new UConfigClient($env);
    }

    public static function ins() {
        if(!(self::$ins instanceof self)) {
            self::$ins = new self();
        }
        return self::$ins;
    }

    public function setConfig($uid, $key, $val, $ttl) {
        $res = $this->_UcClient->setConfig(self::APP, $uid, $key, $val, $ttl);
        return $this->formatSDKOutput($res);
    }

    public function incrTtl($uid, $key, $val, $ttl) {
        $res = $this->_UcClient->incrTtl(self::APP, $uid, $key, $val, $ttl);
        return $this->formatSDKOutput($res);
    }
}
