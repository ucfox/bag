<?php

Class XyModel {

    private static $ins;

    private $_logger     = "xy";
    private $_address    = "parlist.xingyan.pdtv.io:8360";
    private $_httpclient = null;
    private $_caller     = "bag";
    private $_sign       = null;

    const API_TIME_OUT = 1;

    public function __construct() {
        $this->_sign = $_SERVER["XY_SIGN"];
        $env = $_SERVER['ENV'];
        if($env != "online"){
            $this->_caller = "test";
        }

        $this->_httpclient = new HttpRequest($this->_logger);
    }

    public static function ins() {
        if(!(self::$ins instanceof self)){
            self::$ins = new self();
        }
        return self::$ins;
    }

    private function _getBaseUrl() {
        return "http://{$this->_address}";
    }

    private function _BuildUrl($apiurl) {
        return $this->_getBaseUrl() . $apiurl . "&_caller=" . $this->_caller;
    }

    public function get($uid) {
        $params = array(
            "rid"  => $uid,
            "_sign" => $this->_sign,
        );
        $url = $this->_BuildUrl("/list?" . http_build_query($params));
        $res = $this->_httpclient->http($url, "", array(), self::API_TIME_OUT);
        $data = json_decode($res, true);
        if ($data && isset($data['errno']) && $data['errno'] == 0) {
            return $data["data"];
        }
        XLogKit::logger($this->_logger)->error("[get fail] res:$res");
        return array();
    }
}
