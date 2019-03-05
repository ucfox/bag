<?php

Yaf_Loader::import("/home/q/php/villa_sdk/villa_sdk.php");
class RoomModel extends SdkBase
{
    const APP          = "pandaren";
    const SCOPE_ROOM   = "1";

    private $_caller   = "bag";

    private $_roomClient;

    private static $ins;

    public function __construct()
    {
        $this->_logger   = XLogKit::logger("villa_api");
        $this->remoteApp = "villa";

        $env = ($_SERVER['ENV'] == 'online' || $_SERVER['ENV'] == 'beta') ? 'online' : 'demo';
        if ($_SERVER["ENV"] == "beta") {
            $env = (isset($_SERVER['ISOLATION']) && $_SERVER['ISOLATION']==true) ? 'online' : 'beta';
        }
        $this->_roomClient = new RoomClient($this->_caller, $env, "villa_api");
    }

    public static function ins(){
        if(!(self::$ins instanceof self)){
            self::$ins = new self();
        }
        return self::$ins;
    }

    public function getRoomByCond($cond) {
        $res = $this->_roomClient->getRoomByCond($cond);
        return $this->formatSDKOutput($res);
    }
}

