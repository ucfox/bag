<?php

Class GiftModel {

    private static $ins;

    private $_logger     = "mallapi";
    private $_env        = "";
    private $_address    = "mall.pdtv.io:8360";
    private $_httpclient = null;
    private $_caller     = "gift_web";
    const RETRY_SUCCESS  = 4009;
    const GOODS_IDS_EXT_NAME_PREFIX = [193,195,196]; //类英雄会员职业体系，需要带着制作人给mall的goods_id 列表

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

    public function __construct($env = '') {
        $this->_env        = $env ? : $_SERVER['ENV'];
        $this->_httpclient = new HttpRequest($this->_logger);
    }

    public function send($goods, $uid, $num, $uuid, $hostid, $roomid, $expire, $repair = true) {

        $callerstrace = Context::get("callerstrace");
        $params       = array(
            "giftid"      => $goods["pid"],
            "rid"         => $uid,
            "uuid"        => $uuid,
            "count"       => $num,
            "hostid"      => $hostid,
            "callerstrace" => $callerstrace,
            "pdft"        => Tools::getPdft(),
            "ip"          => Tools::getIp(),
            "ext_room_id" => intval(Context::get("gift_ext_roomid")),
            "ext_name_prefix" => (Context::get('gift_ext_name_prefix') ?? ''),
        );
        $url = $this->_BuildUrl("/gift_inside_send?" . http_build_query($params));
        $retry = 1;
        do {
            $res = $this->_httpclient->http($url);
            $data = json_decode($res, true);

            XLogKit::logger($this->_logger)->info("[send gift ] params:" . http_build_query($params));
            if ($data && isset($data['errno']) && ($data['errno'] == 0 || $data['errno'] == self::RETRY_SUCCESS)) {
                return true;
            }
            XLogKit::logger($this->_logger)->info("[send gift fail] retry:$retry res:$res");
            $retry++;
            // 100毫秒
            usleep(100000);
        } while($retry <= 3);

        if($repair) {
            // 自动补单
            $repairModel = new RepairModel();
            $repairModel->add($goods, $uid, $num, $uuid, $hostid, $roomid, $expire);
        }

        XLogKit::logger($this->_logger)->error("[send gift fail] res:$res");
        return false;
    }

}
