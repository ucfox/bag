<?php

/**
 * 首充记录
 */
Class FirstModel
{

    /**
     * 首充初始化状态
     */
    const STATUS_INIT = 0;

    /**
     * 礼包返回完成状态
     */
    const STATUS_FINISH = 1;

    /**
     * 礼包返回失败
     */
    const STATUS_FAIL = 4;

    /**
     * 数据表
     * @var string
     */
    private $table = 'first_charge';

    /**
     * Database
     * @var Mysql
     */
    private $db = null;

    /**
     * 日志
     * @var Logger
     */
    protected $logger = null;

    /**
     * constructor
     */
    public function __construct()
    {
        $this->logger = XLogKit::logger('first');

        try {
            $this->db = Pool::getMysql('master');
        } catch(Exception $e) {
            $this->logger->error('db connect fail ' . __METHOD__ . ': ' . $e->getMessage());
            throw new Exception('db connect fail');
        }
    }

    /**
     * 检查记录是否存在
     *
     * @param int @uid
     * @param int @packlimit
     *
     * @return int
     */
    public function exists($uid, $packlimit)
    {
        // query
        $total = $this->db->getOne("SELECT COUNT(1) AS `total` FROM {$this->table} WHERE uid = {$uid} AND packlimit = {$packlimit}");
        if (is_null($total)) {
            $this->logger->error("db select fail exists uid={$uid}&packlimit={$packlimit}");
            return -1;
        }
        return ((int)$total) > 0 ? 1 : 0;
    }

    /**
     * 带状态检查记录是否存在
     *
     * @param int $uid
     * @param int $packlimit
     * @param int $status
     *
     * @return int
     */
    public function existsWithStatus($uid, $packlimit, $status)
    {
        // query
        $total = $this->db->getOne("SELECT COUNT(1) AS `total` FROM {$this->table} WHERE uid = {$uid} AND packlimit = {$packlimit} AND status = {$status}");
        if (is_null($total)) {
            $this->logger->error("db select fail existsWithStatus uid={$uid}&packlimit={$packlimit}");
            return -1;
        }
        return ((int)$total) > 0 ? 1 : 0;
    }

    /**
     * 获取首充记录数据
     *
     * @param int $uid
     * @param int $packlimit
     *
     * @return array
     */
    public function get($uid, $packlimit)
    {
        return $this->db->getRow("SELECT * FROM {$this->table} WHERE uid = {$uid} AND packlimit = {$packlimit}");
    }

    /**
     * 添加一条首充记录
     *
     * @param int $uid
     * @param int $packlimit
     * @param int $status       记录初始化状态，无特殊需求默认即可
     *
     * @return bool
     */
    public function add($uid, $packlimit, $status = self::STATUS_INIT)
    {
        $now = date('Y-m-d H:i:s');
        $sql = "INSERT INTO {$this->table} (createtime, updatetime, uid, packlimit, status) VALUES ('{$now}', '{$now}', {$uid}, {$packlimit}, {$status})";
        // insert
        $id = $this->db->insert($sql);
        if($id == false) {
            $this->logger->error("db insert fail add uid={$uid}&packlimit={$packlimit}&status={$status}");
        }
        return $id;
    }

    /**
     * 更新首充礼物记录
     *
     * @param int   $id     记录ID
     * @param int[] $goods  礼物ID列表
     * @param int   $status 该操作最终状态，成功或失败
     *
     * @return bool
     */
    public function updateGoods($id, $goods, $status)
    {
        $sql = "UPDATE {$this->table} SET status = {$status}, data = '" . implode(',', $goods) . "' WHERE id = {$id}";
        $rows = $this->db->update($sql);

        if($rows == false) {
            $this->logger->error("db update fail updateGoods uid={$uid}&goods=" . implode(',', $goods) . "&status={$status}");
        }

        return $rows;
    }

    // /**
    //  * 删除首充记录
    //  *
    //  * @param int $uid
    //  * @param int $packlimit
    //  *
    //  * @return bool|mixed
    //  */
    // public function delete($uid, $packlimit)
    // {
    //     return $this->db->del("DELETE FROM {$this->table} WHERE uid = {$uid} AND packlimit = {$packlimit}");
    // }

    /**
     * 首次充值逻辑
     *
     * @param int $uid
     * @param int $packlimit
     * @param array $goods
     *
     * @return bool
     */
    public function charge($uid, $packlimit, array $goods)
    {
        // 添加首充记录
        $id = $this->add($uid, $packlimit, self::STATUS_INIT);
        $count = $sCount = 0;
        $succIds = $succXyIds = array();
        if($id == false) {
            return false;
        }
        $count = $xyCount = 0;
        if(isset($goods[1]) && !empty($goods[1])){

            $succIds = $this->addToIbag($uid, $goods[1]);
            $count = count($goods[1]);
        }
        if(isset($goods[2]) && !empty($goods[2])){
            $xyCount = count($goods[2]);
            $succXyIds = $this->addToXyBag($uid,$goods[2]);
        }

        $count = $count + $xyCount;
        $sCount = count($succIds) + count($succXyIds);
        if ($count == $sCount) {
            $status = FirstModel::STATUS_FINISH;
            $this->logger->info("first charge gift add successed: " . implode(',', $succIds));
        } else {
            $status = FirstModel::STATUS_FAIL;
            $this->logger->warn("first charge gift add failed");
        }

        $succIds = array_merge($succIds,$succXyIds);

        $this->updateGoods($id, $succIds, $status);

        return $status == self::STATUS_FINISH;
    }

    /**
     * 修复未成功的操作
     *
     * @param int $uid
     * @param int $packlimit
     *
     * @return bool
     */
    public function repair($uid, $packlimit, array $goods)
    {
        $exists = $this->exists($uid, $packlimit);
        if($exists === 0 || $exists === -1) {
            return false;
        }

        $record = $this->get($uid, $packlimit);
        $status  = FirstModel::STATUS_FINISH;
        $succIds = $succXyIds = array();

        $id   = $record['id'];
        $uid  = $record['uid'];
        $data = $record['data'];

        if((int)$record['status'] == self::STATUS_FINISH) {
            return true;
        }

        if($data == '') {
            $filter = [];
        } else {
            $filter = explode(',', $data);
        }
        $igoods = isset($goods[1]) ? $goods[1] : array();
        $xygoods = isset($goods[2]) ? $goods[2] : array();


        foreach($filter as $gid){
            if(isset($igoods[$gid])){
                unset($igoods[$gid]);
                continue;
            }
            if(strpos('xy',$gid) === 0 ){
                $xyid = substr($gid,2);
                if(isset($xygoods[$xyid])){
                    unset($xygoods[$xyid]);
                }
            }

        }


        if(count($igoods) > 0) {
            // 需要添加物品
            $succIds = $this->addToIbag($uid, $igoods);
        }
        if(count($xygoods)>0){
            $succXyIds = $this->addToXyBag($uid, $xygoods);
        }
        $succGoods = array_merge($succIds,$succXyIds);
        if ((count($succIds) == count($igoods)) && (count($succXyIds) == count($xygoods))) {
            $status = FirstModel::STATUS_FINISH;

            $this->logger->info("first charge gift add successed: " . implode(',', $succGoods));
        } else {
            $status = FirstModel::STATUS_FAIL;
            $this->logger->warn("first charge gift add failed");
        }

        $succIds = array_merge($succIds,$succXyIds);
        $succIds = array_merge($filter, $succIds);

        $this->updateGoods($id, $succIds, $status);

        return $status == self::STATUS_FINISH;
    }

    /**
     * 向ibag添加物品，带日志。返回添加成功物品列表
     *
     * @param int   $uid
     * @param array $goods
     *
     * @return int[]
     */
    protected function addToIbag($uid, array $goods)
    {
        // modles
        $ibagModel  = new IbagModel();
        $guuidModel = new GuuidModel;

        $succIds   = [];

        foreach($goods as $gid => $data) {
            $num = $data['num'];
            $retry = 0;
            $uuid  = $guuidModel->get();

            // if the uuid is zero
            if (empty($uuid)) {
                for($i = 0; $i < 2; $i++) {
                    $uuid  = $guuidModel->get();
                    if(empty($uuid)) {
                        // 500 ms
                        usleep(500000);
                    } else {
                        break;
                    }

                }
            }
            // throw failed
            if(empty($uuid)) {
                $this->logger->warn("uuid generate fail goodsId={$gid}&uid={$uid}&num={$num}&uuid={$uuid}");
                return $succIds;
            }

            for(; $retry <= 3; $retry++) {
                // 添加物品
                $res = $ibagModel->add($uid, $gid, $uuid, $num);
                if($res && $res['errno'] == 0) {
                    // 添加成功

                    $succIds[] = $gid;

                    break;
                } else {
                    $this->logger->warn("ibag add fail goodsId={$gid}&uid={$uid}&num={$num}&uuid={$uuid}&retry={$retry}&ret_errno={$res['errno']}&ret_errmsg={$res['errmsg']}");
                }
                // 500 ms
                usleep(500000);
            }

            // 添加失败
            if ($retry > 3) {
                $this->logger->error("ibag add fail goodsId={$gid}&uid={$uid}&num={$num}&uuid={$uuid}&retry={$retry}");
                continue;
            }

            $this->logger->info("ibag add success goodsId={$gid}&uid={$uid}&num={$num}&uuid={$uuid}&retry={$retry}");
        }

        // 返回添加成功物品
        return $succIds;

    }

    protected function addToXyBag($uid, array $goods)
    {
        // modles
        $XypackageModel  = new XypackageModel();

        $succIds   = [];

        foreach($goods as $gid => $goods) {

            // 添加物品
            $res = $XypackageModel->addunused($uid, $gid, $goods['num'], $goods['xy_type'], $goods['effect_date'], $goods['effect_time']);
            if($res && $res['errno'] == 0) {
                // 添加成功

                $succIds[] = "xy_" . $gid;

                break;
            } else {
                $this->logger->warn("xybag add fail goodsId={$gid}&uid={$uid}&num={$goods['num']}&type={$goods['xy_type']}&effect_date={$goods['effect_date']}&effect_time={$goods['effect_time']}&ret_errno={$res['errno']}&ret_errmsg={$res['errmsg']}");
            }
        }


        $this->logger->info("xybag add success goodsId={$gid}&uid={$uid}&num={$goods['num']}&type={$goods['xy_type']}&effect_date={$goods['effect_date']}&effect_time={$goods['effect_time']}");

        // 返回添加成功物品
        return $succIds;

    }

}
