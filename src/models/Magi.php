<?php

Class MagiModel {

    private static $ins;

    private $_logger     = "magi";
    private $_env        = "";
    private $_address    = "magi.pdtv.io:8360";
    private $_httpclient = null;
    private $_caller     = "gift_web";

    const SEND_FAIL = "2";
    const SEND_SUCC = "3";

    public function __construct($env = '') {
        $this->_env        = $env ? : $_SERVER['ENV'];
        $this->_httpclient = new HttpRequest($this->_logger);
    }

    public static function ins() {
        if(!(self::$ins instanceof self)){
            self::$ins = new self();
        }
        return self::$ins;
    }

    private function _getBaseUrl() {
        if ($this->_env === 'online') {
            $prefix = '';
        } else {
            $prefix = 'beta.';
        }

        return "http://{$prefix}{$this->_address}";
    }

    private function _BuildUrl($apiurl) {
        return $this->_getBaseUrl() . $apiurl . "&__plat=" . $this->_caller;
    }

    public function sendGiftStatus($uuid) {
        $params     = array(
            "uuid" => $uuid,
        );
        $url = $this->_BuildUrl("/gift/send_free_uuid_query?" . http_build_query($params));
        $data = $this->_httpclient->http($url);
        $data = json_decode($data, true);
        if ($data && isset($data['errno']) && $data['errno'] == 0) {
            return $data["data"];
        }
        return false;
    }

}
