<?php

/**
 * 脚本用于给指定用户增加物品
 * 当前更新: 2018-01-18 to 2018-01-23
 * /home/q/tools/pylon_rigger/rigger php -s crontab -f /home/q/system/bag/src/worker/bag_add_cron.php
 */

// 初始化文件
require_once 'init.php';

/**
 * 计划任务
 */
class Task
{

    /**
     * task name
     */
    const TASK_NAME = 'bag_add_cron';

    /**
     * @var float
     */
    private $now = null;

    /**
     * @var int
     */
    private $time = null;

    /**
     * Repair constructor.
     */
    public function __construct()
    {
        // 当前时间
        $this->now  = microtime(true);
        $this->time = time();
    }

    //脚本，修复180414 bjtb迁移 未发count sdk 导致的经验卡消费失败问题, 回补经验卡
    public function run180414()
    {

        // start
        XLogKit::logger(self::TASK_NAME)->info('start0414');

        //uid,goodsId,num
        $file = explode(PHP_EOL,file_get_contents("/tmp/bag_180413_ques.txt"));

        foreach($file as $row)
        {
            $info = explode(',', $row);
            $goodsId = $info[1];
            $num     = $info[2];
            $uids    = [
                           $info[0]
                       ];

            // Add
            $result = $this->addToBag($uids, $goodsId, $num);

            if(is_bool($result)) {
                // Failed
                XLogKit::logger(self::TASK_NAME)->error('AddToBag failed uids=' . json_encode($uids) . "&goodsId={$goodsId}&num={$num}");
            } else {
                // Something successed or failed
                XLogKit::logger(self::TASK_NAME)->info('AddToBag successed uids=' . json_encode($uids) . "&goodsId={$goodsId}&num={$num}&result=" . json_encode($result));
            }
        }

        // stop
        XLogKit::logger(self::TASK_NAME)->info('stop0414');
    }

    /**
     * 执行操作
     */
    public function run180118()
    {
        $beginTime = strtotime('2018-01-18 00:00:00', $this->time);
        $endTime   = strtotime('2018-01-23 23:59:59', $this->time);
        // 时间范围判断
        if($this->time < $beginTime || $this->time > $endTime) {
            XLogKit::logger(self::TASK_NAME)->warn("Time out: {$beginTime} to {$endTime}");
            exit;
        }

        // start
        XLogKit::logger(self::TASK_NAME)->info('start');

        // 参数
        $goodsId = 118;
        $num     = 12;
        $uids    = [
                       '97928948'
                   ];

        // Add
        $result = $this->addToBag($uids, $goodsId, $num);

        if(is_bool($result)) {
            // Failed
            XLogKit::logger(self::TASK_NAME)->error('AddToBag failed uids=' . json_encode($uids) . "&goodsId={$goodsId}&num={$num}");
        } else {
            // Something successed or failed
            XLogKit::logger(self::TASK_NAME)->info('AddToBag successed uids=' . json_encode($uids) . "&goodsId={$goodsId}&num={$num}&result=" . json_encode($result));
        }

        // stop
        XLogKit::logger(self::TASK_NAME)->info('stop');
    }

    /**
     * 向用户背包增加物品
     *
     * @param array $uids
     * @param int   $goodsId
     * @param int   $num
     * @return array|bool
     */
    protected function addToBag(array $uids, $goodsId, $num = 1)
    {
        // 用户列表不能为空
        if(count($uids) < 1) {
            return false;
        }
        // 物品ID/数量单位需要合法
        if($goodsId < 1 || $num < 1) {
            return false;
        }

        // uuid
        $guuidModel = new GuuidModel();
        // ibag model
        $ibagModel = new IbagModel("xuhan_email");

        // 执行结果
        $result = [];
        // 批量处理用户
        foreach($uids as $uid) {
            // 获取uuid
            $uuid = $guuidModel->get();
            if(empty($uuid)) {
                for($i = 0; $i < 2; $i++) {
                    $uuid = $guuidModel->get();
                    if(empty($uuid)) {
                        // 500 ms
                        usleep(500000);
                    } else {
                        break;
                    }
                }
                // Fail to get uuid
                if(empty($uuid)) {
                    XLogKit::logger(self::TASK_NAME)->error("Invalid uuid AddToBag failed uid={$uid}&goodsId={$goodsId}&num={$num}&uuid={$uuid}");
                    $result[$uid] = false;
                    continue;
                }
            }

            $res = $ibagModel->add($uid, $goodsId, $uuid, $num);
            if(!$res || $res['errno'] != 0) {
                // Failed
                XLogKit::logger(self::TASK_NAME)->error("AddToBag failed uid={$uid}&goodsId={$goodsId}&num={$num}&uuid={$uuid}&res=" . json_encode($res));
                $result[$uid] = false;
            } else {
                // Success
                XLogKit::logger(self::TASK_NAME)->info("AddToBag successed uid={$uid}&goodsId={$goodsId}&num={$num}&uuid={$uuid}&res=" . json_encode($res));
                $result[$uid] = true;
            }
        }

        return $result;

    }

    public function run180622(){

        $path =  dirname(APP_PATH) . '/tmp/';
        $conf = array(
            'v1' => array('goods_id'=>'174','num'=>5),
            'v2' => array('goods_id'=>'174','num'=>3),
            'v3' => array('goods_id'=>'174','num'=>1),
            'v4' => array('goods_id'=>'132','num'=>2),
            'v5' => array('goods_id'=>'132','num'=>1),
            'v6' => array('goods_id'=>'131','num'=>2),
            'v7' => array('goods_id'=>'131','num'=>1),
        );

 //       $_SERVER['ENV'] = 'beta';
        foreach($conf as $k=>$v){
            $uids = [];
            $file = $path . $k . '.txt';
            $fp = fopen($file,'r');
            while(!feof($fp)){
                $uid = fgets($fp,1024);
                $uid = trim($uid);
                if($uid){
                    $uids[] = intval($uid);
                }
            }
            $this->addToBag($uids,$v['goods_id'],$v['num']);
        }
    }

    public function __destruct() {
        $stime = microtime(true) - $this->now;
        XLogKit::logger(self::TASK_NAME)->info("usetime: {$stime}");
    }

}

/**
 * Entry point
 *
 * @param int       $argc
 * @param string[]  $argv
 */
function main($argc, $argv) {
    $task = new Task();
//    $task->run180622();
}

