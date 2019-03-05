<?php
class DvaModel
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
        $this->token = "425e758ec21f87a6aa8eae14d164c47e";
        if ($env === 'online' ) {
            $user_prefix = '';
            $this->token = "78939a42a2f48edd270dea5833fe2ec2";
        }
        $host = "{$user_prefix}dva.pdtv.io";

        $timeout = 1;
        $caller = 'bag';
        $this->logger = XLogKit::logger('dva');
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
    }



    //房间是否开通粉丝勋章
    public function fansMedal($anchorid){


        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , anchorid==$anchorid");

        $path = " /fans/medal_text?anchorid=" . $anchorid;

        $ret = $this->httpClient->get($path);

        $data = json_decode($ret,true);

        if(!isset($data['errno']) || $data['errno'] != 0){
            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , metal_text ret false , anchorid==$anchorid");
        }

        $ret = isset($data['data']) ? $data['data'] : array();

        return $ret;
    }

    //增加粉丝经验
    public function addFansExp($exp,$anchorid,$rid){


        $token = $this->token;
        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , anchorid==${anchorid}&rid=${rid}&exp=${exp}&token=${token}");

        $callerstrace = Context::get("callerstrace");

        $params = [
            'exp' => $exp,
            'anchorid' => $anchorid,
            'rid' => $rid,
            'token' => $token,
            'callerstrace' => $callerstrace,
        ];

        $ret = $this->httpClient->post("/chlv/add",$params);

        $data = json_decode($ret,true);
        if(!isset($data['errno']) || $data['errno'] != 0){
            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , /chlv/add ret false , anchorid==${anchorid}&rid=${rid}&exp=${exp}&token=${token}");
        }

        $ret = isset($data['data']) ? $data['data'] : array();

        return $ret;
    }

}


