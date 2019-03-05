<?php

Class BagController extends Controller_Intenal {

    // 监控背包信息
    public function monitorAction() {
        // 监控信息
        $info = [];
        $info = array_merge($info, $this->getInfoOfTodayFTQ());

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $info);
    }

    // 获取佛跳墙监控信息
    protected function getInfoOfTodayFTQ()
    {
        try {
            $redis = new RedisJanna($_SERVER["REDIS_BAG_CALLNAME"], $_SERVER["REDIS_BAG_TARGET"], $_SERVER["MASTER_REDIS_PWD_BAG"]);
        } catch(Exception $e) {
            return -1;
        }

        // 读取cms
        $data = @file_get_contents("/home/q/cmstpl/bag/april_fool_day.json");
        if(empty($data)) {
            $config = [];
        } else {
            $config = json_decode($data, true);
            if(json_last_error() != JSON_ERROR_NONE || !is_array($config)) {
                $config = [];
            }
        }

        $maxNum     = isset($config['max_num']) ? (int)$config['max_num'] : 0; // 当天最大数量
        $ratio      = isset($config['ratio'])   ? (int)$config['ratio']   : 0; // 概率系数
        $giftReal   = '5aa7405a3c74f317b96beddb';
        $giftFake   = '5aa741c93c74f317b96bf119';
        $key        = RedisKey::todayGiftMaxNum($giftReal);

        return [
            'today_ftq_count' => (int)($redis->get($key) ?? 0),
            'today_ftq_ratio' => $ratio,
            'max_ftq_count'   => $maxNum,
        ];

    }
    // 消费背包物品
    public function takeAction() {

        $goodsId       = intval($this->postParam("goods_id", true));
        $expire        = intval($this->postParam("expire", true));
        $num           = intval($this->postParam("num", true));
        $hostid        = intval($this->postParam("hostid", true));
        $roomid        = intval($this->postParam("roomid", true));
        $giftExtRoomid = intval($this->postParam("gift_ext_roomid", true));
        $uid           = intval($this->postParam("uid", true));
        $room_type     = intval($this->postParam("room_type",true,1));
        $caller        = $this->reqParam("caller", true);

        // gift_ext_roomid 特殊写到上下文里
        Context::set("gift_ext_roomid", $giftExtRoomid);

        Context::set("room_type", $room_type);
        XLogKit::logger('take')->info("[take action] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} gift_ext_roomid:{$giftExtRoomid} caller:{$caller}");
        // 前端返回剩余数量
        $data["num"] = 0;

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
            $uid     == "") {

            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], $data);
        }

        // 检查token
        //if (!Token::checkPc($uid, $token)) {
        //    $this->RespJson(RetCode::TOKEN_ERROR, RetCode::$RetMsg[RetCode::TOKEN_ERROR], $data);
        //}

        $res = TakeModel::ins()->takeGoods($uid, $goodsId, $expire, $num, $hostid, $roomid);
        if ($res["res"] === false) {
            XLogKit::logger('take')->warn("[take action fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));
            $this->RespJson($res["errno"], $res["errmsg"], $data);
        }
        XLogKit::logger('take')->info("[take action succ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));

        $data["num"] = (int)$res["data"];
        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }

    function propCateListAction(){
        $arr = array(
            PropModel::COLOR_SEPAK_CARD => "彩色弹幕卡", // 彩色弹幕卡
            PropModel::CHANGE_NAME_CARD => "改名卡",     // 改名卡
            PropModel::EXP_CARD         => "经验卡",        // 经验卡
            PropModel::MAGIC_GIFT       => "魔术礼物",  // 魔术礼物
            PropModel::FORBID_ACCESS    => "世界弹幕卡类",//世界弹幕卡
            PropModel::BADGE_CARD       => "用户勋章卡",  // 勋章
            PropModel::FAN_BADGE_CARD   => "粉丝勋章卡", //粉丝勋章专属
            PropModel::OCC_CARD         => "英雄职业卡", // 职业
            PropModel::FANS_EXP_CARD    => "粉丝经验卡", // 粉丝贡献卡
            PropModel::GRANK_CARD       => "分手卡类", //消除排行卡
            PropModel::SALVO_CARD       => "礼炮狂热卡", //礼炮狂热卡
        );

        $this->RespJson(RetCode::SUCCESS,RetCode::$RetMsg[RetCode::SUCCESS],$arr);
    }

}
