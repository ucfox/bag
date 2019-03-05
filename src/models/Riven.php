<?php

Class RivenModel {

    private static $ins;

    private $_logger     = null;
    private $_env        = "";
    private $_address    = "riven.pdtv.io:8360";
    private $_httpclient = null;
    private $_caller     = "bag";

    public function __construct($env = '') {
        $this->_logger     = XLogKit::logger("riven");
        $this->_env        = $env ? : $_SERVER['ENV'];
        $this->_httpclient = new HttpRequest('riven');
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
        return $this->_getBaseUrl() . $apiurl;
    }

    public function sendGlobalMsg($msgType,$data) {
        $msg = [
            'type'=>$msgType,
            'time'=>time(),
            'data'=>$data,
        ];
        $msg = json_encode($msg,JSON_UNESCAPED_UNICODE);
        $url = $this->_BuildUrl("/msg/send_global_msg?app=pcgame_mmlive&priority=4");
        $rsStr = $this->_httpclient->http($url,$msg,['Content-Type'=>'application/json']);
        $rs = json_decode($rsStr, true);

        if ($rs && isset($rs['errno']) && $rs['errno'] == 0) {
            $this->_logger->info(__FUNCTION__.", data: ".$msg."   ret:".$rsStr);
            return $rs["data"];
        }else{
            $this->_logger->error(__FUNCTION__.", data: ".$msg."   ret:".$rsStr);
        }
        return false;
    }

}
