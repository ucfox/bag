<?php

/**
 * 首充控制器
 */
Class FirstController extends Controller_Intenal
{

    /**
     * 日志
     * @var Logger
     */
    protected $logger = null;

    private $msg = '';
    /**
     * Initialize controller
     */
    public function init()
    {
        $this->logger = XLogKit::logger('first');
    }

    /**
     * The first charge
     *
     * @param int @rid
     * @param int @packlimit
     */
    public function chargeAction()
    {
        // params
        $rid       = (int)$this->postParam("rid",        true);
        $packlimit = (int)$this->postParam("packlimit",  true);
        $type = $this->postParam("type",  true, 'first_charge');
        //$comb = $this->postParam("comb", true, 0);
        $lev = (int)$this->postParam("lev",  true, 1);

        $this->logger->info("chargeAction params: rid={$rid}&packlimit={$packlimit}&lev={$lev}&type={$type}");

        if (empty($rid) || empty($packlimit)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS]);
        }

        try {
            $firstModel = new FirstModel();
        } catch(Exception $e) { // 数据库连接异常
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }
        // 检查首充是否存在记录
        switch($firstModel->exists($rid, $packlimit)) {
            case 1:
                $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS]);
            case 0:
                break;
            case -1:
            default:
                $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }
        // 添加首充记录
        $goods = $this->getFirstGoods($type, $lev);
        $ret   = $firstModel->charge($rid, $packlimit, $goods);

        // Send message
        if($ret) {
        //    $list = PackageModel::ins()->getCombGoods($comb,$lev);
           // $msg = isset($list['name']) ? $list['name'] : "7天炫彩弹幕卡";
            $msg = $this->msg;
            $this->sendMsg($rid, $lev, $msg);
            $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS]);
        }

        $this->RespJson(RetCode::GOODS_ADD_FAIL, RetCode::$RetMsg[RetCode::GOODS_ADD_FAIL]);

    }

    /**
     * 修复首充记录
     *
     * @param int @rid
     * @param int @packlimit
     */
    public function repairAction()
    {
        // params
        $rid       = (int)$this->postParam("rid",        true);
        $packlimit = (int)$this->postParam("packlimit",  true);
        $type = $this->postParam("type",  true, 'first_charge');
        $lev = (int)$this->postParam("lev",  true, 1);

        $this->logger->info("chargeAction params: rid={$rid}&packlimit={$packlimit}");

        if (empty($rid) || empty($packlimit)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS]);
        }

        try {
            $firstModel = new FirstModel();
        } catch(Exception $e) { // 数据库连接异常
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        // 修复首充记录
        $goods = $this->getFirstGoods($type, $lev);
        $ret   = $firstModel->repair($rid, $packlimit, $goods);

        // ret
        if($ret) {
            $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS]);
        }

        $this->RespJson(RetCode::GOODS_ADD_FAIL, RetCode::$RetMsg[RetCode::GOODS_ADD_FAIL]);

    }

    /**
     * 获取首充赠送礼物列表
     *
     * @return int[]
     */
    protected function getFirstGoods($type, $lev)
    {
        $res = array();
        $goods_list = PackageModel::ins()->get($type, $lev);
        $this->msg = isset($goods_list['send_msg']) ? $goods_list['send_msg']: '';
        foreach($goods_list['list'] as $v){
            if(isset($v['goods_id'])){
                //type  1:主站 2:星颜
                $type = isset($v['type']) ? $v['type'] : 1;
                $res[$type][$v['goods_id']]['num'] = $v['num'];
                if($type == 2){
                    $res[$type][$v['goods_id']]['xy_type'] = $v['xy_type'];
                    $res[$type][$v['goods_id']]['num'] = $v['num'];
                    $res[$type][$v['goods_id']]['effect_date'] = $v['effect_date'];
                    $res[$type][$v['goods_id']]['effect_time'] = $v['effect_time'];
                }
            }
        }

        // combile
        return $res;
    }

    /**
     * 推送提示消息
     *
     * @param int $uid
     * @param int $num
     *
     * @return bool
     */
    protected function sendMsg($uid, $lev, $msg = "")
    {
        //todo 稍后重配
        $now = date('Y-m-d');

        if($msg){
            $content = $msg . $now;
        }else{
            return true;
        }
        return Message::sendMessageToUser('first', $uid, 5, '获得首次充值礼包通知', $content);
    }

}
