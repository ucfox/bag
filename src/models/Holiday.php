<?php
class HolidayModel
{
    protected $logger;
    private $httpClient;
    private $caller;

    public static function ins(){
        return new static();
    }

    public function __construct($env=''){
        if ($env === '') {
            $env = $_SERVER['ENV'];
        }
        $user_prefix = 'beta.';
        if ($env === 'online' ) {
            $user_prefix = '';
        }
        $host = "{$user_prefix}holiday.pdtv.io";

        $timeout = 2;
        $caller = 'bag';
        $this->logger = XLogKit::logger('holiday');
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
    }

    public function getBoxInfo($hostid){


        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , hostid=$hostid");

        $ret = $this->httpClient->get('/ajax_get_level_info?hostid='.$hostid);

        $data = json_decode($ret,true);
        if(!isset($data['errno']) || $data['errno']!=0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , hostid=$hostid,getBoxInfo ret:".$ret);

            return false;
        }


        return $data['data'];
    }

    //获取礼炮Pk时状态
    public function getSalvoInfo($hostid){

        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , hostid=$hostid");

        $ret = $this->httpClient->get('/game_pk_check_wildcard?hostid='.$hostid);

        $data = json_decode($ret,true);


        if(!isset($data['errno']) || $data['errno']!=0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , hostid=$hostid,getSalvoInfo ret:".$ret);

            return false;
        }

        return $data['data'];
    }

    public function salvoHot($hostid, $rid){

        $callerstrace = Context::get("callerstrace");

        $ret = $this->httpClient->get('/game_pk_use_wildcard?hostid='.$hostid . '&rid=' . $rid . '&callerstrace=' . $callerstrace);

        $data = json_decode($ret,true);


        if(!isset($data['errno']) || $data['errno']!=0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , hostid=${hostid}&rid=${rid},getSalvoInfo ret:" . $ret);

            return false;
        }

        if(!isset($data['data']) || !$data['data']){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , hostid=${hostid}&rid=${rid},getSalvoInfo ret:" . $ret);

            return false;

        }

        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , hostid=${hostid}&rid=${rid},getSalvoInfo ret:" . $ret);

        return true;
    }
}


