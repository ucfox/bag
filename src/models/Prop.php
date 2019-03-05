<?php

Class PropModel {

    const COLOR_SEPAK_CARD = 1; // 彩色弹幕卡
    const CHANGE_NAME_CARD = 2; // 改名卡
    const EXP_CARD         = 3; // 经验卡
    const MAGIC_GIFT       = 4; // 魔术礼物
    const FORBID_ACCESS    = 5; // 禁止从bag消费，由其他系统调用ibag.take消费
    const BADGE_CARD       = 6; // 勋章卡
    const FAN_BADGE_CARD   = 8; // 粉丝勋章卡
    const OCC_CARD         = 7; // 职业卡
    const FANS_EXP_CARD    = 9; // 粉丝经验卡
    const GRANK_CARD = 11; //分手卡
    const SALVO_CARD       = 10; //礼炮

    const SPEAK_COLOR_WORD_VAL = 1;

    //分手卡
    const REMOVE_CARD_ID = 335;
    const REPAIR_CARD_ID = 349;
    // 彩色弹幕卡，去uconfig 设置过期时间
    public function speakColorWord($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        $ttl = Tools::getTtl($goods["use_expire"], $goods["use_etype"]);
        if ($ttl < 0) {
            XLogKit::logger('take')->warn("invalid card with use_expire and use_etype speakColorWord goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }
        // 支持多张卡
        $ttl = $ttl * $num;

        // 假如客户端未更新成新版本，依然会从Uconfig获取弹幕卡周期信息
        // 所以要对旧的弹幕卡做兼容
        if(in_array((int)$goods['id'], [3, 4, 6]) &&
            !UconfigModel::ins()->incrTtl($uid, "color_speak_card", self::SPEAK_COLOR_WORD_VAL, $ttl)) {
            XLogKit::logger('take')->error("use barrage card to uconfig fail speakColorWord goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        // 解析ext字段
        $ext = $goods['ext'];
        if(empty($ext)) {
            XLogKit::logger('take')->warn("invalid card with empty ext field speakColorWord goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }
        $ext_json = json_decode($ext, true);
        if(json_last_error() !== JSON_ERROR_NONE) {
            XLogKit::logger('take')->warn("invalid card with ext field error speakColorWord goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        // 处理conf参数
        $conf = [
            // 'cardid'        => $goods['id'],
            'expire_time'   => $ttl,
        ];
        if(isset($ext_json['style']) &&
            is_array($ext_json['style'])) {
            foreach ($ext_json['style'] as $style => $value) {
                $conf[$style] = $value;
            }
        }

        // 使用道具卡
        $ret = BizModel::ins()->useCard($uid, $conf);
        // 使用道具卡失败
        if(!$ret) {
            XLogKit::logger('take')->error("use barrage card fail speakColorWord goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&ret={$ret}");
            return false;
        }

        return true;

    }

    // 改名卡，调用ruc增加改名次数
    public function changeName($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        return UserModel::ins()->incrModifyNicknameTimes($uid, "pandaren", $uuid, $num);
    }

    // 经验卡
    public function expCard($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {

        $ext = json_decode($goods['ext'],true);

        if (!is_array($ext)){
            $exp = $goods['ext'] * $num;
        }else{
            $exp = $ext["common"] * $num;
        }


        $countModel = new CountModel($uid, $exp);
        $detail['roomid'] = $roomid;
        $detail['hostid'] = $hostid;
        return $countModel->addExp($uid, $exp, $detail);
    }

    // 魔术礼物(暂时只支持真/假佛跳墙)
    public function sendMagicGift($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        XLogKit::logger('take')->info("params sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");

        try {
            $redis = new RedisJanna($_SERVER["REDIS_BAG_CALLNAME"], $_SERVER["REDIS_BAG_TARGET"], $_SERVER["MASTER_REDIS_PWD_BAG"]);
        } catch(Exception $e) {
            XLogKit::logger('take')->error("connect redis failed sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&error=" . $e->getMessage());
            return false;
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

        XLogKit::logger('take')->info("cms config sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&config=" . json_encode($config));

        $maxNum      = isset($config['max_num']) ? (int)$config['max_num'] : 0; // 当天最大数量
        $ratio       = isset($config['ratio'])   ? (int)$config['ratio']   : 0; // 概率系数
        $giftReal    = '5aa7405a3c74f317b96beddb';
        $giftFake    = '5aa741c93c74f317b96bf119';

        $giftid      = $giftFake;
        $ratio       = min(max($ratio, 0), 10000); // 确保该值在0 ~ 10000 之间

        // 计算概率
        if(Tools::calcProbability($ratio, 10000)) {
            XLogKit::logger('take')->info("in probability sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&config=" . json_encode($config));
            // 读取计数器
            $key = RedisKey::todayGiftMaxNum($giftReal);
            if(($curNum = $redis->incr($key)) == 1) {
                // 设置过期时间
                $redis->expireat($key, strtotime('tomorrow'));
            }
            // 检查当天生成数量
            if($_SERVER['ENV'] != 'online' || $curNum <= $maxNum) {
                $giftid = $giftReal;
                XLogKit::logger('take')->info("current number sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&config=" . json_encode($config) . "&number={$curNum}");
            } else {
                $redis->decr($key);
            }
        } else {
            XLogKit::logger('take')->info("not in probability sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&config=" . json_encode($config));
        }

        XLogKit::logger('take')->info("the giftid sendMagicGift goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&config=" . json_encode($config) . "&giftid={$giftid}");

        $giftModel = new GiftModel();
        return $giftModel->send(['pid' => $giftid], $uid, $num, $uuid, $hostid, $roomid, $expire, false);
    }

    //勋章
    public function sendBadge($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {


        $ext = json_decode($goods['ext'],true);

        if (!is_array($ext)){
            $badge_id = $goods['ext'];
        }else{
            $badge_id = $ext["common"];
        }

        if(empty($badge_id)){
                XLogKit::logger('take')->warn("badge_id is empty goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
                return false;
        }

        $ttl = $goods["use_expire"];
        if ($ttl < 0) {
            XLogKit::logger('take')->warn("invalid badge with use_expire and use_etype badge goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        $endTime = $ttl > 0 ? (time() + $ttl)*1000 : 0;

        $badgeModel = new BadgeModel();
        $info =  $badgeModel->addByBcode($uid, $badge_id, $endTime);


        if(!isset($info['errno'])){
            XLogKit::logger('take')->error(Sprintf("[%s], BADGE_ADD_SYS_ERROR. call fail:%s ", __METHOD__, json_encode($info)));
            return false;
        }

        if( $info['errno']!==0 ){
            XLogKit::logger('take')->error(Sprintf("[%s],BADGE_ADD_ERROR. call return false: %s ", __METHOD__, json_encode($info)));
            return false;
        }


        XLogKit::logger('take')->info("sendBadge goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");

        return true;
    }
    // 经验卡
    public function sendFansExpCard($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        $ext = json_decode($goods['ext'],true);

        if (!is_array($ext)){
            $exp = $goods['ext'] * $num;
        }else{
            $exp = $ext["common"] * $num;
        }
        if ($exp  <= 0) {
            XLogKit::logger('take')->warn("invalid fansExpCard with exp goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        $dvaModel = new DvaModel();
        return $dvaModel->addFansExp($exp, $hostid, $uid);
    }

    //call 粉丝勋章专属
    public function sendFansBadge($goods, $uid, $num, $uuid, $hostid, $roomid, $expire){

        XLogKit::logger('take')->info("sendFansBadge goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");

        return true;
    }
    //职业卡

    public function sendOccupation($goods, $uid, $num, $uuid, $hostid, $roomid, $expire){

        $arr_cate = ['fashi', 'qishi', 'mushi', 'jueshi'];
        $ext = json_decode($goods['ext'],true);

        $cate = isset($ext['cate']) ? $ext['cate'] : '';

        $level = isset($ext['level']) ? (int)$ext['level'] : 0;

        $howlong = isset($ext['use_expire']) ? (int)$ext['use_expire'] : 7;

        $banner = isset($ext['banner']) ? $ext['banner'] : 1;
        $luckbag = isset($ext['luckbag']) ? $ext['luckbag'] : 1;
        $cost = 0;


        if($roomid <=0 ){
            XLogKit::logger('take')->warn("param roomid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");
            return false;
        }

        if($hostid <= 0){
            XLogKit::logger('take')->warn("param hostid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");
            return false;
        }


        if(!in_array($cate, $arr_cate)){
                XLogKit::logger('take')->warn("param occ_cate is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");
                return false;
        }

        if($level < 1){
                XLogKit::logger('take')->warn("param occ_level is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");

                return false;
        }

        $occModel = OccupationModel::ins();

        $info =  $occModel->addByCate(['rid'=>$uid,'cate'=>$cate,'level'=>$level,'cost'=>$cost,'roomid'=>$roomid,'hostid'=>$hostid,'howlong'=>$howlong,'banner'=>$banner,'luckbag'=>$luckbag]);

        if(!is_array($info)){
                XLogKit::logger('take')->warn("[addByCate ret false]  goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}&howlong={$howlong}&banner={$banner}&luckbag={$luckbag}");
                return false;
        }
        return true;
    }
    //分手卡
    public function sendRemove($goods, $uid, $num, $uuid, $hostid, $roomid, $expire){
        if($roomid <=0 ){
            XLogKit::logger('take')->warn("param roomid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");
            return false;
        }

        if($hostid <= 0){
            XLogKit::logger('take')->warn("param hostid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}&cate={$cate}&level={$level}");
            return false;
        }

        if($goods['id'] == self::REMOVE_CARD_ID){

            $data = GrankModel::ins()->addBlackList($hostid,$uid);
        }
        if($goods['id'] == self::REPAIR_CARD_ID){

            $data = GrankModel::ins()->remBlackList($hostid,$uid);
        }

        if(empty($data)){
            return false;
        }

        if($data['data'] != 1){
            XLogKit::logger('take')->warn("sendRemove ret " . json_encode($data) . " goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
        }
        return true;

    }
    // 礼炮狂热卡
    public function sendSalvoCard($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        if($roomid <=0 ){
            XLogKit::logger('take')->warn("param roomid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        if($hostid <= 0){
            XLogKit::logger('take')->warn("param hostid is error goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        $holidayModel = new HolidayModel();

        $ret = $holidayModel->salvoHot($hostid, $uid);

        if($ret == false){
            XLogKit::logger('take')->warn("[salvoHot ret false]  goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");
            return false;
        }

        XLogKit::logger('take')->info("sendSalvoCard goods=" . json_encode($goods) . "&uid={$uid}&num={$num}&uuid={$uuid}&hostid={$hostid}&roomid={$roomid}&expire={$expire}");

        return true;
    }
}
