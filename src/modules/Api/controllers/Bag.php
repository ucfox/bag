<?php

Class BagController extends Controller_Api {
    //世界弹幕卡
    const S_CARD = array('world_card'=>array(174,285),'titan_card'=>array(314));
    //限制调用次数
    const LIMIT_NUM = 6;
    // 获取背包物品数量, 已经切换到lua
    public function numAction() {
        $uid = Context::get("uid");
        // 这个参数不能验证，移动端ios傻逼前几期没加这个参数
        $api_version = floatval($this->reqParam("api_version", true));

        $ret = array(
            "total" => "45",
            "use"   => "0"
        );

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], $ret);
        }

        $data = BagModel::ins()->num($uid, $api_version, 0);
        if (!$data) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR], $ret);
        }

        $ret["total"] = (string)$data["total"];
        $ret["use"] = (string)$data["use"];

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $ret);
    }

    // 获取背包物品
    public function getAction() {
        $uid = Context::get("uid");
        // 这个参数不能验证，移动端ios傻逼前几期没加这个参数
        $api_version = floatval($this->reqParam("api_version", true));

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        $data = BagModel::ins()->get($uid, $api_version, 0);
        if (!$data) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }

    //用户某物品的数量
    public function gnumAction(){

        $uid = Context::get("uid");
        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }
        $goodsId = $gid = $this->getParam("goods_id", true);

        $goodsId = explode(',',$goodsId);

        $num = 0;

        $is_wc = false;

        $is_tt = false;

        $ids = array();

        $ret = array();

        $world_num = 0;

        $titan_num = 0;

        $tmp_ids = array();

        array_walk($goodsId, function($v)use (&$ids, &$tmp_ids, &$is_wc, &$is_tt){
            if(isset(self::S_CARD[$v])){
                if($v == "world_card"){
                    $is_wc = true;
                }
                if($v == "titan_card"){
                    $is_tt = true;
                }
                $ids = array_merge($ids,self::S_CARD[$v]);
            }else{
                $tmp_ids[] = $v;
            }
        });

        $ids = array_merge($tmp_ids,$ids);

        $ids = array_unique($ids);

        $count = count($ids);

        if($count > self::LIMIT_NUM){

            XLogKit::logger('gnum')->warn("[gnum action] error ids gt 6 goods_id=${gid}&uid=${uid}");

            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        foreach($ids as $id){

            $id = (int)$id;

            if($id < 0){
                XLogKit::logger('gnum')->warn("[gnum action] goods_id error goods_id=${id}&uid=${uid}");
                continue;
            }

            $data = IbagModel::ins()->goodsNum($uid, $id);

            if ($data === false) {
                continue;
            }

            if(in_array($id,$goodsId)){
                $ret[$id] = $data;
            }

            if($is_wc == true){

                $world_num += in_array($id,self::S_CARD['world_card']) ? $data : 0;

                $ret['world_card'] = $world_num;
            }

            if($is_tt == true){

                $titan_num += in_array($id,self::S_CARD['titan_card']) ? $data : 0;

                $ret['titan_card'] = $titan_num;
            }

            $num += $data;
        }

        $ret['num'] = $num;
        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $ret);
    }

    //世界弹幕卡的个数, 用于直播间发言区块世界弹幕卡个数展现
    public function wcnumAction(){

        $uid = Context::get("uid");

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        //TODO 优化为一个请求
        $goods_174_num = IbagModel::ins()->goodsNum($uid, 174);
        $goods_285_num = IbagModel::ins()->goodsNum($uid, 285);

        if ($goods_174_num === false || $goods_285_num === false) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], ['num'=>intval($goods_174_num) + intval($goods_285_num)]);
    }

    // 消费背包物品
    public function takeAction() {
        $uid = Context::get("uid");

        $goodsId = $this->postParam("goods_id", true);
        $expire  = $this->postParam("expire", true);
        $num     = $this->postParam("num", true);
        $hostid  = $this->postParam("hostid", true);
        $roomid  = $this->postParam("roomid", true);

        $__plat      = $this->reqParam("__plat", true);
        $__version   = $this->reqParam("__version", true);
        $__channel   = $this->reqParam("__channel", true);
        $pt_time     = $this->reqParam("pt_time", true);
        $pt_sign     = $this->reqParam("pt_sign", true);
        $api_version = floatval($this->reqParam("api_version", true));

        XLogKit::logger('take')->info("[take action] __plat:{$__plat} __version:{$__version} __channel:{$__channel} pt_sign:{$pt_sign} pt_time:{$pt_time} api_version:{$api_version} uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");

        // 移动端特殊要的返回结果
        $data = array(
            "goods_id" => (string)$goodsId,
            "expire"   => (string)$expire,
            "num"      => "0",
        );

        if (strpos($goodsId, "xy") !== false) {
            $this->RespJson(RetCode::XY_GOODS, RetCode::$RetMsg[RetCode::XY_GOODS], $data);
        }

        $goodsId = intval($goodsId);
        $expire  = intval($expire);
        $num     = intval($num);
        $hostid  = intval($hostid);
        $roomid  = intval($roomid);

        if ($goodsId <= 0 ||
            $expire  < 0 ||
            $num     <= 0 ||
            $uid     == "" ||
            $api_version < 1) {

            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], $data);
        }
        // 检查token
        if (!Token::checkMb($uid, $pt_sign, $pt_time)) {
            $this->RespJson(RetCode::TOKEN_ERROR, RetCode::$RetMsg[RetCode::TOKEN_ERROR], $data);
        }

        $res = TakeModel::ins()->takeGoods($uid, $goodsId, $expire, $num, $hostid, $roomid, $api_version);
        if ($res["res"] === false) {
            XLogKit::logger('take')->warn("[take action fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));
            $this->RespJson($res["errno"], $res["errmsg"], $data);
        }
        XLogKit::logger('take')->info("[take action succ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));

        $data["num"] = (string)$res["data"];
        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }
}
