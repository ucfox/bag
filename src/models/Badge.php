<?php
class BadgeModel
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
        $host = "{$user_prefix}badge.pdtv.io";

        $timeout = 2;
        $caller = 'bag';
        $this->logger = XLogKit::logger('badge');
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
    }

    public function addByBcode($rid, $bcode, $endTime){

        $params['rid'] = $rid;
        $params['bcode'] = $bcode;
        $params['endTime'] = $endTime;

        $callerstrace = Context::get('callerstrace');

        $params['callerstrace'] = $callerstrace;

        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , rid=$rid&bcode=$bcode&endTime=$endTime");

        $ret = $this->httpClient->post('/user/add_badge_by_bcode', $params);
        $data = json_decode($ret,true);


        return $data;
    }
    public function getBadges($rid, $isDetail= true){

        $path =  '/user/badges?rid=' . $rid;
        //$params['rid'] = $rid;

        if($isDetail){
           // $params['includedetail'] = true;
            $path = $path . '&includedetail=true';
        }
        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , rid=$rid&includedetail=true");

        $ret = $this->httpClient->get($path);
        $data = json_decode($ret,true);


        return $data;
    }


}


