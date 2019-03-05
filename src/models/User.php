<?php
Yaf_Loader::import("/home/q/php/ruc_sdk/ruclient.php");
class UserModel
{
    protected $client = null;
    protected static $_ins = null;

    public static function ins() {
        if(self::$_ins == null) {
            self::$_ins = new static();
        }
        return self::$_ins;
    }

    public function __construct() {
        $this->client = RUClient::getInstance("bag");
    }

    public function getSimpleInfoById($rid) {
        $info = $this->client->getAvatarOrNickByRids(array($rid));
        if($info && $info['errno'] == 0) {
            $data = array_values($info['data'])[0];
            return $data;
        } else {
            XLogKit::ins('ruc')->error('[error from remote ruc] '.$info['errno'].' '.$info['errmsg']);
            return array();
        }
    }

    public function auth() {
        $data = $this->client->auth();
        if($data['data'] && $data['data']['rid']) {
            return $data;
        } else {
            return array();
        }
    }

    // 增加改名次数
    public function incrModifyNicknameTimes($uid, $app, $uuid, $num) {
        $retry = 1;
        do {
            $info = $this->client->incrModifyNicknameTimes($uid, $app, $uuid, $num);
            XLogKit::ins('ruc')->info("[remote ruc retry:$retry] ".$info['errno'].' '.$info['errmsg'].' '.$info['data']);
            if($info && $info['errno'] == 0 && $info["data"] !== false) {
                return true;
            }
            $retry++;
            sleep(1);
        } while($retry <= 3);

        XLogKit::ins('ruc')->error('[err from remote ruc] '. json_encode($info));
        return false;
    }
}
