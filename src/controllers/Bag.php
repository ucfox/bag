<?php

Class BagController extends Controller_Index {

    public function numAction() {
        $uid = Context::get("uid");
        $version = intval($this->getParam("pc_version", true));

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        $data = BagModel::ins()->num($uid, 0, $version);
        if (!$data) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }

    public function gnumAction(){

        $uid = Context::get("uid");
        $goodsId = (int)$this->getParam("goods_id", true);

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        if($goodsId < 1){

            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");

        }
        $data = IbagModel::ins()->goodsNum($uid, $goodsId);

        if ($data === false) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
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

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], intval($goods_174_num) + intval($goods_285_num));
    }

    // 获取背包物品
    public function getAction() {
        $uid = Context::get("uid");
        $version = intval($this->getParam("pc_version", true));

        if (empty($uid)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        $data = BagModel::ins()->get($uid, 0, $version);
        if (!$data) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }
        // 临时兼容
        if ($version < 2) {
            $data["total"] = 15;
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }


    // 消费背包物品
    public function takeAction() {
        $uid = Context::get("uid");

        $goodsId       = intval($this->postParam("goods_id", true));
        $expire        = intval($this->postParam("expire", true));
        $num           = intval($this->postParam("num", true));
        $hostid        = intval($this->postParam("hostid", true));
        $roomid        = intval($this->postParam("roomid", true));
        $giftExtRoomid = intval($this->postParam("gift_ext_roomid", true));
        $type          = intval($this->postParam("type", true, 1));
        $token         = $this->reqParam("token", true);

        // gift_ext_roomid 特殊写到上下文里
        Context::set("gift_ext_roomid", $giftExtRoomid);

        XLogKit::logger('take')->info("[take action] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} gift_ext_roomid:{$giftExtRoomid} token:{$token} type:{$type}");

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
        if (!Token::checkPc($uid, $token)) {
            $this->RespJson(RetCode::TOKEN_ERROR, RetCode::$RetMsg[RetCode::TOKEN_ERROR], $data);
        }

        $res = TakeModel::ins()->takeGoods($uid, $goodsId, $expire, $num, $hostid, $roomid,'pc',$type);
        if ($res["res"] === false) {
            XLogKit::logger('take')->warn("[take action fail] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));
            $this->RespJson($res["errno"], $res["errmsg"], $data);
        }
        XLogKit::logger('take')->info("[take action succ] uid:{$uid} goodsId:{$goodsId} expire:{$expire} num:{$num} hostid:{$hostid} roomid:{$roomid} res:" . json_encode($res));

        $data["num"] = (int)$res["data"];
        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }

    public function convertAction(){
        $uid = Context::get("uid");
        if (empty($uid)) {
            $this->RespError(RetCode::USER_NOT_LOGIN);
        }
        $rows = array_column(json_decode(file_get_contents('/home/q/cmstpl/pandatv/cmsconfig/bag/goods_convert.json'),1),null,'convert_id');
        $convertId = intval($this->postParam("convert_id", true));
        XLogKit::logger('convert')->info("[convert init] convertId:{$convertId} uid:{$uid}");
        if(!isset($rows[$convertId])){
            $this->RespError("兑换项不存在");
        }
        $row = $rows[$convertId];
        if($row['s_num']<1){
            $this->RespError("兑换项异常");
        }

        //个人背包内物品数量是否足够
        $gnum = intval(IbagModel::ins()->goodsNum($uid, $row['s_id']));
        if($gnum<$row['s_num']){
            // $this->RespError("背包内的{$row['s_name']}不足{$row['s_num']}个，不能兑换。");
            $this->RespError("未达到兑换条件哦，快去试试手气吧~");
        }

        //取出源物品 s:source
        $takeUuid = GuuidModel::ins()->get();
        $takeRet = IbagModel::ins()->autoTake($uid,$row['s_id'],$takeUuid,$row['s_num']);
        if(!isset($takeRet['errno'])||$takeRet['errno']!=0){
            XLogKit::logger('convert')->warn("[take action fail] convertId:{$convertId} uid:{$uid} goodsId:{$row['s_id']} num:{$row['s_num']} ret:" . json_encode($takeRet));
            $this->RespError("扣除源物品({$row['s_name']}*{$row['s_num']})失败，请重试");
        }

        //兑换目标物品 d:destination
        $addUuid = GuuidModel::ins()->get();
        $addRet = IbagModel::ins()->add($uid,$row['d_id'],$addUuid,$row['d_num']);
        if(!isset($addRet['errno'])||$addRet['errno']!=0){
            XLogKit::logger('convert')->error("[add action fail] convertId:{$convertId} uid:{$uid} goodsId:{$row['d_id']} num:{$row['d_num']} ret:" . json_encode($addRet));
            $this->RespError("兑换物品({$row['d_name']}*{$row['d_num']})失败，请重试");
        }

        //发送横幅
        $bannerId = trim($row['banner_id']);
        if($bannerId){
            $msgDdata['from']['nickName'] = Context::get('nickName');
            $msgDdata['to']['roomid'] = $msgDdata['to']['toroom'] = '';
            $msgDdata['content']['giftid'] = $bannerId;
            $msgDdata['content']['giftname'] = $row['d_name'];
            RivenModel::ins()->sendGlobalMsg(3303,$msgDdata);
        }
        $this->RespSuccess();
    }
}
