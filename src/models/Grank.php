<?php
class GrankModel
{
    protected $logger;
    private $httpClient;
    private $caller;
    private $token;

    public static function ins(){
        return new static();
    }

    public function __construct($env=''){
        if ($env === '') {
            $env = $_SERVER['ENV'];
        }
        $user_prefix = 'beta.';
        $this->token = "M4ROh7PD2oI8Rwvqf94F213nLy140a5B";
        if ($env === 'online' ) {
            $user_prefix = '';
            $this->token = "S8Nk8R673tLk6pbws3iwD4SvdsmzkrAM";
        }
        $host = "{$user_prefix}grank.pdtv.io";

        $timeout = 1;
        $caller = 'bag';
        $this->logger = XLogKit::logger('grank');
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
    }



    //房间是否开通粉丝勋章
    public function blackList($anchorid){


        $token = $this->token;
        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , anchor_id==$anchorid");

        $path = "/blacklist/get_room?anchor_id=" . $anchorid . "&token=" . $token;

        $ret = $this->httpClient->get($path);

        $data = json_decode($ret,true);

        if(!isset($data['errno']) || $data['errno'] != 0){
            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , blacklist ret false , anchorid=={$anchorid}");
            return false;
        }

        $ret = isset($data['data']) ? $data['data'] : false;

        return $ret;
    }

    //添加黑名单
    public function addBlackList($anchorid,$rid){

        $callerstrace = Context::get("callerstrace");

        $token = $this->token;
        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , anchor_id=={$anchorid}&uid={$rid}&token={$token}");

        $params = [
            'anchor_id' => $anchorid,
            'uid' => $rid,
            'callerstrace' => $callerstrace,
            'token' => $token
        ];

        $ret = $this->httpClient->post("/blacklist/add_to_room",$params);

        $data = json_decode($ret,true);
        if(!isset($data['errno']) || $data['errno'] != 0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , /blacklisk/add_to_room ret false , anchorid=={$anchorid}&rid={$rid}&exp={$exp}&token={$token}");

            return false;
        }

        $ret = isset($data['data']) ? json_decode($data['data'],true) : false;

        return $ret;
    }
    //移除黑名单
    public function remBlackList($anchorid,$rid){


        $token = $this->token;
        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , anchor_id=={$anchorid}&uid={$rid}&token={$token}");

        $params = [
            'anchor_id' => $anchorid,
            'uid' => $rid,
            'token' => $token
        ];

        $ret = $this->httpClient->post("/blacklist/remove_from_room",$params);

        $data = json_decode($ret,true);
        if(!isset($data['errno']) || $data['errno'] != 0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , /blacklist/remove_from_room ret false , anchorid=={$anchorid}&rid={$rid}&exp={$exp}&token={$token}");

            return false;
        }

        $ret = isset($data['data']) ? json_decode($data['data'],true) : false;

        return $ret;
    }

}


