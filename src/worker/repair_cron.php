<?php
// 这个脚本脚本修复背包物品使用失败的情况
// 一期不考虑彩色弹幕卡和改名卡的的回补
require_once "init.php";
function main($argc, $argv) {
    $obj = new Repair();
    $obj->execute();
}
Class Repair {
    private $now = NULL;
    private $repairModel = NULL;

    private $_procFunc = array(
        "gift"                      => "_procGift",       // 礼物
        PropModel::COLOR_SEPAK_CARD => "_procSpeakColor", // 彩色弹幕卡
        PropModel::CHANGE_NAME_CARD => "_procChangeName", // 改名卡
    );

    public function __construct() {
        $this->now = time();
        $this->repairModel = new RepairModel();
    }
    public function execute() {
        XLogKit::logger('repair_cron')->info("start");
        $list = $this->repairModel->query();
        if (empty($list)) {
            XLogKit::logger('repair_cron')->info("empty list");
            return true;
        }

        foreach($list as $k => $record) {
            $goods = IbagModel::ins()->getGoods($record["goods_id"]);
            if ($goods === false) {
                XLogKit::logger('repair_cron')->warn("goods not exists goods_id=" . $record["goods_id"]);
                // 增加retry次数
                $this->repairModel->updateStatus($record["id"], RepairModel::STATUS_INIT);
                continue;
            }

            $funcName = $this->_getProcFuncName($goods);
            if (empty($funcName)) {
                XLogKit::logger('repair_cron')->warn("funcName empty goods=" . json_encode($goods));
                // 增加retry次数
                $this->repairModel->updateStatus($record["id"], RepairModel::STATUS_NOT);
                continue;
            }

            if(!$this->$funcName($record, $goods)) {
                XLogKit::logger('repair_cron')->info("proc fail funcName=$funcName&record=" . json_encode($record). "&goods=" . json_encode($goods));
            } else {
                XLogKit::logger('repair_cron')->info("proc success funcName=$funcName&record=" . json_encode($record). "&goods=" . json_encode($goods));
            }
        }
    }

    // 返回处理的函数名
    private function _getProcFuncName($goods) {
        if ($goods["ptype"] == TakeModel::PTYPE_GIFT ) {
            return $this->_procFunc["gift"];
        }
        if ($goods["ptype"] == TakeModel::PTYPE_PROP && isset($this->_procFunc[$goods["pid"]])) {
            return $this->_procFunc[$goods["pid"]];
        }
        return false;
    }

    // 礼物补单
    private function _procGift($record, $goods) {
        // 从magi查询状态
        $magiStatus = $this->_getMagiStatus($record);
        // 返回true跳过当前记录
        if (!$magiStatus) {
            return false;
        }
        // 未来一个小时之内物品不会过期, 原格子回补
        $uuid = GuuidModel::ins()->get();
        if (empty($uuid)) {
            return false;
        }

        // 原格子回补: 有效期大于当前时间+1小时
        if ($record["expire"] > ($this->now + 3600)) {
            return $this->_repairGoods($record, $uuid);
        }
         // 添加物品
        return $this->_addGoods($record, $uuid);
    }

    // 从magi查询状态，magiStatus为3时，该订单是成功的；当magiStatus为2时该订单是失败的，可以给用户回补；其他值代表未知(不知道成功还是失败)，可以持续重试
    private function _getMagiStatus($record) {
        $magiStatus = MagiModel::ins()->sendGiftStatus($record["uuid"]);
        // 可以补
        if ($magiStatus === MagiModel::SEND_FAIL) {
            return true;
        }
        // 不需要回补
        if ($magiStatus === MagiModel::SEND_SUCC) {
            // 更新记录
            if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_NOT)) {
                XLogKit::logger('repair_cron')->error("update status fail status= " . RepairModel::STATUS_NOT . "&record=". json_encode($record));
            }
            return false;
        }
        // 未知状态，增加retry, 下次轮训再查一次
        if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_INIT)) {
            XLogKit::logger('repair_cron')->error("update status fail status= " . RepairModel::STATUS_INIT . "&record=". json_encode($record));
        }

        return false;
    }

    // 彩色弹幕卡补单 todo
    private function _procSpeakColor($record, $goods) {
        return RepairModel::STATUS_ERROR;
    }
    // 改名卡补单 todo
    private function _procChangeName($record, $goods) {
        return RepairModel::STATUS_ERROR;
    }

    private function _repairGoods($record, $uuid) {
        $res = IbagModel::ins('repair')->repair($record["uid"], $record["goods_id"], $record["expire"], $uuid, $record["num"]);
        // 记录日志，更新数据状态
        if ($res && isset($res["errno"]) &&$res["errno"] == 0) {
            XLogKit::logger('repair_cron')->info("repair success uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
            // 更新记录
            if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_ORIGIN)) {
                XLogKit::logger('repair_cron')->error("update status fail status=" . RepairModel::STATUS_ORIGIN . "&uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
            }
            return true;
        }
        // 更新记录
        if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_ERROR)) {
            XLogKit::logger('repair_cron')->error("update status fail status=". RepairModel::STATUS_ERROR . "&uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
        }
        XLogKit::logger('repair_cron')->error("repair fail uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
        return false;
    }

    private function _addGoods($record, $uuid) {
        $res = IbagModel::ins('repair')->add($record["uid"], $record["goods_id"], $uuid, $record["num"]);
        // 记录日志，更新数据状态
        if ($res && isset($res["errno"]) &&$res["errno"] == 0) {
            // 更新记录
            if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_ADD)) {
                XLogKit::logger('repair_cron')->error("update status fail status=" . RepairModel::STATUS_ADD . "&uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
            }
            XLogKit::logger('repair_cron')->info("add success uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
            return true;
        }
        // 更新记录
        if (!$this->repairModel->updateStatus($record["id"], RepairModel::STATUS_ERROR)) {
            XLogKit::logger('repair_cron')->error("update status fail status=". RepairModel::STATUS_ERROR . "&uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
        }
        XLogKit::logger('repair_cron')->error("add fail uuid=$uuid&record=". json_encode($record) . "&res=" . json_encode($res));
        return false;
    }

    public function __destruct() {
        $stime = time() - $this->now;
        XLogKit::logger('repair_cron')->info("usetime: ". $stime);
    }
}
