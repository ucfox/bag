<?php

// 消费背包物品逻辑
Class TakeModel {

    private static $ins;

    const PTYPE_GIFT = "0";
    const PTYPE_PROP = "1";

    // 排麦
    const SUPPORT_PAIMAI_VERSION = 100.0; // 这里临时写100,移动端上线再改
    const PAIMAI_ROOM = 798;


    // 背包物品调用的类
    private $_classConf = array(
        self::PTYPE_GIFT => "GiftModel", // 礼物
        self::PTYPE_PROP => "PropModel", // 道具
    );

    // 背包道具调用的函数, key 为prop id
    private $_propFunc = array(
        PropModel::COLOR_SEPAK_CARD => "speakColorWord", // 彩色弹幕卡
        PropModel::CHANGE_NAME_CARD => "changeName",     // 改名卡
        PropModel::EXP_CARD         => "expCard",        // 经验卡
        PropModel::MAGIC_GIFT       => "sendMagicGift",  // 魔术礼物
        PropModel::BADGE_CARD       => "sendBadge",  // 勋章
        PropModel::FAN_BADGE_CARD   => "sendFansBadge", //粉丝勋章专属
        PropModel::OCC_CARD         => "sendOccupation", // 职业
        PropModel::FANS_EXP_CARD    => "sendFansExpCard", // 粉丝贡献卡
        PropModel::GRANK_CARD => "sendRemove", //消除排行卡
        PropModel::SALVO_CARD       => "sendSalvoCard", //礼炮狂热卡
    );

    // api_version支持的道具
    private $_propSupport = array(
        "1.0" => array(PropModel::COLOR_SEPAK_CARD),
        "1.1" => array(PropModel::COLOR_SEPAK_CARD, PropModel::CHANGE_NAME_CARD, PropModel::EXP_CARD),
    );

    public static function ins() {
        if(!(self::$ins instanceof self)) {
            self::$ins = new self();
        }
        return self::$ins;
    }

    public function takeGoods($uid, $goodsId, $expire, $num, $hostid, $roomid, $api_version = "pc", $type = 1) {
        // 有效期判断
        if ($expire != "0" && $expire < time()) {
            XLogKit::logger('take')->warn("[take model expire out of date] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::EXPIRE_OUT_OF_DATE, "errmsg" => RetCode::$RetMsg[RetCode::EXPIRE_OUT_OF_DATE]);
        }

        // 获取物品属性
        // TODO 获取用户背包详情
        $goods = IbagModel::ins()->getGoods($goodsId);
        if ($goods === false) {
            XLogKit::logger('take')->warn("[take model get goods fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::GOODS_NOT_EXISTS, "errmsg" => RetCode::$RetMsg[RetCode::GOODS_NOT_EXISTS]);
        }

        // 特殊逻辑，移动端排麦特殊礼物处理
        if ($goods["ptype"] == self::PTYPE_GIFT && $api_version != "pc" && $roomid == self::PAIMAI_ROOM && $api_version < self::SUPPORT_PAIMAI_VERSION) {
            XLogKit::logger('take')->warn("[take model mobile paimai fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} api_version:{$api_version}");
            return array("res" => false, "errno" => RetCode::ROOM_NOT_SUPPORT, "errmsg" => RetCode::$RetMsg[RetCode::ROOM_NOT_SUPPORT]);
        }
        // 特殊处理，排麦房间礼物不让送
        if ($goods["ptype"] == self::PTYPE_GIFT && $roomid == self::PAIMAI_ROOM) {
            XLogKit::logger('take')->warn("[take model paimai not gift] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} api_version:{$api_version}");
            return array("res" => false, "errno" => RetCode::PAIMAI_NOT_GIFT, "errmsg" => RetCode::$RetMsg[RetCode::PAIMAI_NOT_GIFT]);
        }

        // 道具版本兼容
        if ($api_version != "pc" && $goods["ptype"] == self::PTYPE_PROP && isset($this->_propSupport[$api_version]) && !in_array($goods["pid"], $this->_propSupport[$api_version])) {
            XLogKit::logger('take')->warn("[take model prop support] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} api_version:{$api_version}");
            return array("res" => false, "errno" => RetCode::PROP_VERSION_UPGRADE, "errmsg" => RetCode::$RetMsg[RetCode::PROP_VERSION_UPGRADE]);
        }
        if ($api_version != "pc" && $goods["ptype"] == self::PTYPE_PROP && $goods["pid"] == PropModel::COLOR_SEPAK_CARD && $api_version < 1.3 && !in_array($goodsId, array(3, 4, 6)) ) {
            XLogKit::logger('take')->warn("[take model prop fall behind] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} api_version:{$api_version}");
            return array("res" => false, "errno" => RetCode::PROP_VERSION_FALL_BEHIND, "errmsg" => RetCode::$RetMsg[RetCode::PROP_VERSION_FALL_BEHIND]);
        }
        // api_version<1.4, 不能使用世界消息卡
        if ($api_version != "pc" && $goods["ptype"] == self::PTYPE_PROP && $goods["pid"] == PropModel::FORBID_ACCESS && $api_version < 1.4 ) {
            XLogKit::logger('take')->warn("[take model prop support] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} api_version:{$api_version}");
            return array("res" => false, "errno" => RetCode::PROP_VERSION_UPGRADE, "errmsg" => RetCode::$RetMsg[RetCode::PROP_VERSION_UPGRADE]);
        }

        // 不支持在此系统消费， 如520活动世界消息卡
        if($goods["ptype"] == self::PTYPE_PROP && $goods["pid"] == PropModel::FORBID_ACCESS){
            $retcode = isset(RetCode::FORBID_ACCESS_GOODS[$goodsId]) ? RetCode::FORBID_ACCESS_GOODS[$goodsId] : RetCode::FORBID_ACCESS_GOODS[0];
            return array("res" => false, "errno" => $retcode, "errmsg" => RetCode::$RetMsg[$retcode]);
        }

        // 如果该物品的scope是房间内使用，则必须传roomid和hostid
        if (($goods["scope"] == IbagModel::SCOPE_ROOM) && (empty($hostid) || empty($roomid))) {
            XLogKit::logger('take')->warn("[take model hostid or roomid empty ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::INVALID_PARAMS, "errmsg" => RetCode::$RetMsg[RetCode::INVALID_PARAMS]);
        }

        // 验证hostid 和roomid是否合法
        if ($goods["scope"] == IbagModel::SCOPE_ROOM) {
            $cond = array("filter" => "instroduction", "roomid" => $roomid);
            $room = RoomModel::ins()->getRoomByCond($cond);
            if (empty($room) || $room["hostid"] != $hostid) {
                XLogKit::logger('take')->warn("[take model room empty] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
                return array("res" => false, "errno" => RetCode::INVALID_PARAMS, "errmsg" => RetCode::$RetMsg[RetCode::INVALID_PARAMS]);
            }
        }

        // 特殊判断如果是礼物，不能自己给自己送
        if ($goods["ptype"] == self::PTYPE_GIFT && $uid == $hostid) {
            XLogKit::logger('take')->warn("[take model gift can't send yourself ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::GIFT_SELF, "errmsg" => RetCode::$RetMsg[RetCode::GIFT_SELF]);
        }

        // 会掉落礼物的道具，不能自己赠送给自己
        if ($goods["ptype"] == self::PTYPE_PROP && $goods["pid"] == PropModel::MAGIC_GIFT && $uid == $hostid) {
            XLogKit::logger('take')->warn("[take model magic gift can't send yourself ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::GIFT_SELF, "errmsg" => RetCode::$RetMsg[RetCode::GIFT_SELF]);
        }

        //类英雄(会员)职业体系, 若是类似goods_id的，带上制作人
        if(in_array($goodsId , GiftModel::GOODS_IDS_EXT_NAME_PREFIX )){

            $userBag = IbagModel::ins()->get($uid);
            if(!isset($userBag['details']) ){
                XLogKit::logger('take')->error("[take model call api bag/get return error] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, res:".json_encode($userBag));
                return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
            }
            if( !isset($userBag['goods'][$goodsId]) || empty($userBag['details']) ){
                XLogKit::logger('take')->info("[take model call api bag/get return goodsId non-exists] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, res:".json_encode($userBag));
                return array("res" => false, "errno" => RetCode::NUM_NOT_ENOUGH, "errmsg" => RetCode::$RetMsg[RetCode::NUM_NOT_ENOUGH]);
            }

            $gift_ext_name_prefix = '';
            foreach($userBag['details'] as $v){
                if(!(in_array($v['goods_id'] ,GiftModel::GOODS_IDS_EXT_NAME_PREFIX) && $v['expire'] == $expire)){
                    continue;
                }
                if(empty($v['ext'])){
                    XLogKit::logger('take')->warn("[take model call api bag/get return details ext non-exists] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, res:".json_encode($userBag));
                    continue;
                }
                $v_ext_arr = json_decode($v['ext'], true);
                if(empty($v_ext_arr['pid'])){
                    XLogKit::logger('take')->warn("[take model call api bag/get return details ext non-exists] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, ext:{$v['ext']}, ");
                    continue;
                }
                //pid的name
                $user_obj = new UserModel();
                $pid_user_info = $user_obj->getSimpleInfoById($v_ext_arr['pid']);
                if(empty($pid_user_info['nickName'])){
                    XLogKit::logger('take')->warn("[take model call api ruc return nickName non-exists] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, res:".json_encode($pid_user_info));
                    continue;

                }
                $gift_ext_name_prefix = $pid_user_info['nickName'].'制作的';
            }
            Context::set('gift_ext_name_prefix', $gift_ext_name_prefix);
        }
        //英雄体系卡
        if($goods['pid'] == PropModel::OCC_CARD){
            $occModel = OccupationModel::ins();

            $ret = $occModel->getOccInfo(array('rid'=>$uid));
            $goodsInfo = json_decode($goods['ext'],true);
            $level = (int)$goodsInfo['level'];

            if(!is_array($ret)){

                XLogKit::logger('take')->warn("[take model call api getOccInfo return false]  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
            }
            if(!empty($ret)){


                if( $ret[$uid]['level'] > $goodsInfo['level'] ){

                    XLogKit::logger('take')->warn("param occ_level is too small,  goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&&level={$level}");
                    return ["res" => false, 'errno' => RetCode::OCC_LEVEL_FORBIN, 'errmsg' => RetCode::$RetMsg[RetCode::OCC_LEVEL_FORBIN]];
                }
            }
        }

        if(PropModel::FAN_BADGE_CARD == $goods['pid'] || PropModel::FANS_EXP_CARD == $goods['pid']){

            $ret = DvaModel::ins()->fansMedal($hostid);

            if(empty($ret) || !isset($ret['medal']) || !$ret['medal']){

                XLogKit::logger('take')->info("[take model call api fans/medal_text return medal empty] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}, res:".json_encode($ret));
                return array("res" => false, "errno" => RetCode::FANS_MEDAL_SUPPORT, "errmsg" => RetCode::$RetMsg[RetCode::FANS_MEDAL_SUPPORT]);

            }
        }
        if(PropModel::SALVO_CARD == $goods['pid']){
            $ret = HolidayModel::ins()->getSalvoInfo($hostid);
            if(!is_array($ret)){

                XLogKit::logger('take')->warn("[take model call api holiday.pdtv.io/game_pk_check_wildcard return {$ret}]  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
            }

            if(!isset($ret['stat']) || $ret['stat'] != 1){

                $r = json_encode($ret);

                XLogKit::logger('take')->warn("[take model call api holiday.pdtv.io/game_pk_check_wildcard return {$r}]  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                $stat = !isset($ret) || $ret['stat'] == 0 ? 0 : 1;

                return array("res" => false, "errno" => RetCode::SALVO_FORBIN, "errmsg" => RetCode::$RetMsg[RetCode::SALVO_FORBIN][$stat]);
            }

        }
        //竹仙宝藏活动大于十级 无法使用
        if($goods['id'] == 261){
           $ret = HolidayModel::ins()->getBoxInfo($hostid);

            if(!is_array($ret)){

                XLogKit::logger('take')->warn("[take model call api holiday.pdtv.io/ajax_get_level_info return false]  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
            }

           $box_flag = isset($ret['level']) && $ret['level'] <= 10 ? true : false;
           if(!$box_flag){

                    XLogKit::logger('take')->warn("host box level more than 10,  goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");
                    return ["res" => false, 'errno' => RetCode::BOX_LEVEL_FORBIN, 'errmsg' => RetCode::$RetMsg[RetCode::BOX_LEVEL_FORBIN]];
           }
        }
        //消除排行状态，分手卡
        if($goods['pid'] == PropModel::GRANK_CARD){
            $ret = GrankModel::ins()->blacklist($hostid);

            if(!is_array($ret)){
                XLogKit::logger('take')->warn("[take model call api blacklist return false]  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
            }

            if(in_array($uid,$ret) && $goods['id'] == PropModel::REMOVE_CARD_ID){

                XLogKit::logger('take')->warn("[take model call api blacklist return ".json_encode($ret)."] has blacklist  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::REMOVE_GRANK_FORBIN, "errmsg" => RetCode::$RetMsg[RetCode::REMOVE_GRANK_FORBIN]);
            }

            if(!in_array($uid,$ret) && $goods['id'] == PropModel::REPAIR_CARD_ID){

                XLogKit::logger('take')->warn("[take model call api blacklist return ".json_encode($ret)."] has not in blacklist  goods=" . json_encode($goods). "&uid={$uid}&num={$num}&hostid={$hostid}&roomid={$roomid}");

                return array("res" => false, "errno" => RetCode::REPAIR_GRANK_FORBIN, "errmsg" => RetCode::$RetMsg[RetCode::REPAIR_GRANK_FORBIN]);
            }
        }

        // 获取uuid 和加日志 add log
        $uuid = GuuidModel::ins()->get();
        if (empty($uuid)) {
            XLogKit::logger('take')->warn("[take model get guuid fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid}");
            return array("res" => false, "errno" => RetCode::GUUID_FAIL, "errmsg" => RetCode::$RetMsg[RetCode::GUUID_FAIL]);
        }

        // 扣背包库存
        if($type != 2){
            $take = IbagModel::ins()->take($uid, $goodsId, $expire, $uuid, $num, json_encode(['hostid'=>$hostid, 'roomid'=>$roomid]));
        }else{
            $take = IbagModel::ins()->autoTake($uid, $goodsId, $uuid, $num, json_encode(['hostid'=>$hostid, 'roomid'=>$roomid]));
        }
        if ($take && !$take["errno"]) {

            $callerstrace = array();
            if(isset($take['data']['expend']) && !empty($take['data']['expend'])){
                foreach($take["data"]["expend"] as $t){
                    $s = array("caller"=>$t["Caller"],"demand"=>$t["Demand"],"num"=>$t["Num"]);
                    $callstrace[] = $s;
                }
                $callerstrace['trace'] = $callstrace;
                $callerstrace['caller'] = 'bag';
                $callerstrace['demand'] = 'bag';
            }
            context::set("callerstrace",json_encode($callerstrace));
            $total = $take["data"]["res"];
        } else {
            // 超扣情况
            if ($take["errno"] == RetCode::NUM_NOT_ENOUGH) {
                XLogKit::logger('take')->warn("[take model decr bag fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} uuid:{$uuid} hostid:{$hostid} roomid:{$roomid}");
                return array("res" => false, "errno" => RetCode::NUM_NOT_ENOUGH, "errmsg" => RetCode::$RetMsg[RetCode::NUM_NOT_ENOUGH]);
            }
            // 物品不存在的情况
            if ($take["errno"] == RetCode::GOODS_NOT_EXISTS_IN_BAG) {
                XLogKit::logger('take')->warn("[take model goods not exists in bag] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} uuid:{$uuid} hostid:{$hostid} roomid:{$roomid}");
                return array("res" => false, "errno" => RetCode::GOODS_NOT_EXISTS_IN_BAG, "errmsg" => RetCode::$RetMsg[RetCode::GOODS_NOT_EXISTS_IN_BAG]);
            }
            // 其他错误
            XLogKit::logger('take')->error("[take model ibag fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} uuid:{$uuid} hostid:{$hostid} roomid:{$roomid} errno:" . $take["errno"] . " errmsg:" . $take["errmsg"]);
            return array("res" => false, "errno" => RetCode::BAG_DECR_FAIL, "errmsg" => RetCode::$RetMsg[RetCode::BAG_DECR_FAIL]);
        }

        // 根据pid和ptype调用不同的类和函数
        $class = $this->_classConf[$goods["ptype"]];
        if (empty($class)) {
            XLogKit::logger('take')->warn("[take model fail class not found] uid:{$uid} goodsId:{$goodsId} class:{$class} ptype:" . $goods["ptype"]);
            return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $func = $this->_getFunc($goods["ptype"], $goods["pid"]);

        if (empty($func)) {
            XLogKit::logger('take')->warn("[take model fail func not found] uid:{$uid} goodsId:{$goodsId} func:{$func} pid:" . $goods["pid"]);
            return array("res" => false, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        XLogKit::logger('take')->info("[take model execute] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} func:$func hostid:{$hostid} roomid:{$roomid}");
        $obj = new $class();
        $res = $obj->$func($goods, $uid, $num, $uuid, $hostid, $roomid, $expire);
        if ($res === false) {
            return array("res" => $res, "errno" => RetCode::SYSTEM_ERROR, "errmsg" => RetCode::$RetMsg[RetCode::SYSTEM_ERROR], "data" => $total);
        }
        return array("res" => $res, "errno" => RetCode::SUCCESS, "errmsg" => RetCode::$RetMsg[RetCode::SUCCESS], "data" => $total);
    }

    // 获取礼物和道具需要调用的函数
    private function _getFunc($ptype, $pid) {
        if ($ptype == self::PTYPE_GIFT) {
            return "send";
        }
        if ($ptype == self::PTYPE_PROP) {
            return $this->_propFunc[$pid];
        }

        return NULL;
    }
}

