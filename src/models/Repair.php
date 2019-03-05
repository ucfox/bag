<?php

Class RepairModel {

    const STATUS_INIT     = 0; // 未处理记录
    const STATUS_ORIGIN   = 1; // 原格子回补
    const STATUS_ADD      = 2; // 新格子回补
    const STATUS_ERROR    = 3; // 错误
    const STATUS_NOT      = 4; // 不需要回补

    public function add($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        try {
            $db = Pool::getMysql("master");
        } catch (exception $e) {
            XLogKit::logger('repair')->error("db connect fail add goodsId={$goods['id']}&uid=$uid&num=$num&$uuid=$uuid&hostid=$hostid&roomid=$roomid&expire=$expire");
            return false;
        }
        $sql = $this->_formatAddRepairSql($goods, $uid, $num, $uuid, $hostid, $roomid, $expire);
        if (!$db->insert($sql)) {
            XLogKit::logger('repair')->error("db add fail goodsId={$goods['id']}&uid=$uid&num=$num&$uuid=$uuid&hostid=$hostid&roomid=$roomid&expire=$expire");
            return false;
        }
        return true;
    }

    private function _formatAddRepairSql($goods, $uid, $num, $uuid, $hostid, $roomid, $expire) {
        $now = date("Y-m-d H:i:s", time());
        $status = self::STATUS_INIT;
        return "INSERT INTO `repair` " .
            "(`createtime`, `updatetime`, `goods_id`, `uuid`, `app`, `uid`, `hostid`, `roomid`, `status`, `num`, `expire`) ".
            "VALUES ".
            "('$now', '$now', {$goods['id']}, $uuid, 'pandaren', $uid, $hostid, $roomid, $status, $num, $expire)";
    }

    public function query() {
        try {
            $db = Pool::getMysql();
        } catch (exception $e) {
            XLogKit::logger('repair')->error("db connect fail query");
            return false;
        }

        $status = self::STATUS_INIT;
        $sql = "SELECT * FROM `repair` WHERE status = $status and retry < 5";
        return $db->getAll($sql);
    }

    public function updateStatus($id, $status) {
        try {
            $db = Pool::getMysql("master");
        } catch (exception $e) {
            XLogKit::logger('repair')->error("db connect fail update");
            return false;
        }

        $sql = "UPDATE `repair` SET `status` = $status, `retry` = `retry` + 1 WHERE `id` = $id";
        return $db->update($sql);
    }
}
