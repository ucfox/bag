<?php

Class BagModel {

    const MY_TYPE = 0;
    const XY_TYPE = 1;

    private static $ins;
    public static function ins() {
        if(!(self::$ins instanceof self)) {
            self::$ins = new self();
        }
        return self::$ins;
    }

    // 获得背包物品详情
    public function get($uid, $apiVersion, $pcVersion) {
        $data = IbagModel::ins()->get($uid);
        if (!$data) {
            return false;
        }
        $data =  $this->_formatBag($data);
        // 增加xy的物品逻辑
        if ($apiVersion > 1.1 || $pcVersion > 2) {
            $data = $this->_formatXy($uid, $data);
        }

        // 这里一定要是空对象，要不然移动端就gg了
        if (empty($data["goods"])) {
            $data["goods"] = (object)array();
        }
        return $data;
    }

    // 格式化一下底层的背包数据，封装一下，只返回有用的
    private function _formatBag($data) {
        $bag = array();
        $bag["total"] = (string)$data["total"];
        $bag["app"] = (string)$data["app"];
        $bag["goods"] = array();

        // 封装一下goods
        foreach($data["goods"] as $k => $goods) {
            $tmp = array();
            $tmp['confirm'] = false;
            $tmp['confirm_msg'] = '';
            $ext = isset($goods['ext']) && !empty($goods['ext']) ? json_decode($goods['ext'],true) : array();
            if(isset($ext['confirm'])&&isset($ext['confirm_msg'])){
                $tmp['confirm'] = $ext['confirm'];
                $tmp['confirm_msg'] = $ext['confirm_msg'];
            }
            $tmp["id"] = (string)$goods["id"];
            $tmp["pid"] = (string)$goods["pid"];
            $tmp["ptype"] = (string)$goods["ptype"];
            $tmp["scope"] = (string)$goods["scope"];
            $tmp["name"] = (string)$goods["name"];
            $tmp["instroduction"] = (string)$goods["instroduction"];
            $tmp["use_expire"] = (string)$goods["use_expire"];
            $tmp["use_etype"] = (string)$goods["use_etype"];
            $tmp["picture"] = json_decode($goods["picture"], true);
            $tmp["system_type"] = (string)self::MY_TYPE;
            $bag["goods"][$k] = $tmp;
        }
        // 封装一下details
        $bag["details"] = array();
        foreach($data["details"] as $k => $detail) {
            $tmp = array();
            $tmp["bag_id"] = (string)$detail["bag_id"];
            $tmp["goods_id"] = (string)$detail["goods_id"];
            $tmp["expire"] = (string)$detail["expire"];
            $tmp["num"] = (string)$detail["num"];
            $tmp["ext"] = (string)$detail["ext"];
            $bag["details"][$k] = $tmp;
        }

        return $bag;
    }

    private function _formatXy($uid, $data) {
        $xyData = XyModel::ins()->get($uid);
        if (empty($xyData)) {
            return $data;
        }
        foreach($xyData as $k => $v) {
            $xyId = "xy_" . $v["goods_id"];
            if (!isset($data["goods"][$xyId])) {
                $data["goods"][$xyId] = array(
                    "id"            => (string)$xyId,
                    "pid"           => "1",
                    "ptype"         => "3",
                    "scope"         => "0",
                    "name"          => (string)$v["name"],
                    "instroduction" => (string)$v["describe"],
                    "use_expire"    => "0",
                    "use_etype"     => "0",
                    "picture"       => array(
                        "mb"        => $v["icon_234_234"],
                        "pc"        => $v["icon_66_66"],
                    ),
                    "system_type"   => (string)self::XY_TYPE,
                );
            }
            $expire = $v["effective_date"] == "0" ? $v["effective_date"] : strtotime($v["effective_date"]);
            $data["details"][] = array(
                "bag_id"   => "0",
                "goods_id" => (string)$xyId,
                "expire"   => (string)$expire,
                "num"      => (string)$v["number"],
            );
        }
        return $data;
    }

    // 背包格子数
    public function num($uid, $apiVersion, $pcVersion) {
        // 这个接口就不封装了，我觉得几百年不会变
        $data = IbagModel::ins()->num($uid);
        if (!$data) {
            return false;
        }
        // 增加xy的物品逻辑
        if ($apiVersion > 1.1 || $pcVersion > 2) {
            $xyData = XyModel::ins()->get($uid);
            if (!empty($xyData)) {
                $num = count($xyData);
                $data["use"] += $num;
            }
        }
        return $data;
    }

}
