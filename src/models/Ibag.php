<?php

Yaf_Loader::import("/home/q/php/ibag_sdk/ibag_sdk.php");
class IbagModel extends SdkBase
{
    const APP          = "pandaren";
    const SCOPE_ROOM   = "1";

    private $_bagClient;
    private $_goodsClient;

    private static $ins;

    public function __construct($caller = 'bag') {
        $this->_logger   = XLogKit::logger("ibag_api");
        $this->remoteApp = "ibag";
        $env = $_SERVER["ENV"];
        $this->_bagClient = new BagClient($caller, $env, "ibag_api");
        $this->_goodsClient = new GoodsClient($caller, $env, "ibag_api");
    }

    public static function ins($caller = 'bag') {
        if(!(self::$ins instanceof self)){
            self::$ins = new self($caller);
        }
        return self::$ins;
    }

    public function get($uid) {
        $res = $this->_bagClient->get(self::APP, $uid);
        return $this->formatSDKOutput($res);
    }

    public function num($uid) {
        $res = $this->_bagClient->num(self::APP, $uid);
        return $this->formatSDKOutput($res);
    }

    public function take($uid, $goodsId, $expire, $uuid, $num, $ext='') {
        $res = $this->_bagClient->take(self::APP, $uid, $goodsId, $expire, $uuid, $num, $ext);
        if(!$res || $res['errno'] != 0) {
            $this->_logger->warn("[from ".$this->remoteApp." sdk] [errno:".$res['errno']."] [errmsg:".$res['errmsg']."]");
        }
        return $res;
    }

    public function autoTake($uid, $goodsId, $uuid, $num, $ext='') {
        $res = $this->_bagClient->autoTake(self::APP, $uid, $goodsId, $uuid, $num, $ext);
        if(!$res || $res['errno'] != 0) {
            $this->_logger->warn("[from ".$this->remoteApp." sdk] [errno:".$res['errno']."] [errmsg:".$res['errmsg']."]");
        }
        return $res;
    }

    public function add($uid, $goodsId, $uuid, $num) {
        $res = $this->_bagClient->add(self::APP, $uid, $goodsId, $num, $uuid);
        if(!$res || $res['errno'] != 0) {
            $this->_logger->warn("[from ".$this->remoteApp." sdk] [errno:".$res['errno']."] [errmsg:".$res['errmsg']."]");
        }
        return $res;
    }

    public function repair($uid, $goodsId, $expire, $uuid, $num) {
        $res = $this->_bagClient->repair(self::APP, $uid, $goodsId, $expire, $uuid, $num);
        if(!$res || $res['errno'] != 0) {
            $this->_logger->warn("[from ".$this->remoteApp." sdk] [errno:".$res['errno']."] [errmsg:".$res['errmsg']."]");
        }
        return $res;
    }

    public function getGoods($goodsId) {
        $res = $this->_goodsClient->get($goodsId);
        return $this->formatSDKOutput($res);
    }

    public function goodsNum($uid, $goodsId){

        $res = $this->_bagClient->gnum(self::APP, $uid, $goodsId);
        return $this->formatSDKOutput($res);
    }
}
