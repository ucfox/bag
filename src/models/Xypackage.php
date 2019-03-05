<?php
class XypackageModel
{
    protected $logger;
    private $caller;
    private $httpClient;

    /**
     * http://wiki.xingyan.pandatv.com:8090/pages/viewpage.action?pageId=5570812
     */
    public static function ins()
    {
        return new static();
    }
    public function __construct ()
    {
        $timeout = 2;

        if ($_SERVER['ENV'] != 'online' ){
            $caller = 'test';
            $host = "beta.package.xingyan.pdtv.io";
            $this->sign = 'test';
        }else{
            $caller = 'pandatv';
            $host = "package.xingyan.pdtv.io";
            $this->sign = '0029fd9273a234530941c884682e7763';
        }

        $this->logger = XLogKit::logger(__CLASS__);
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
        $this->httpClient->setDefinedOptions(array('Content-Type: application/json'));
    }

    public function addunused($rid,$goodsId,$num=1,$type='gift',$effectiveDate=0,$effectTime=0)
    {
        $time = time();
        if($effectiveDate!=0 && $effectiveDate<$time){
            $effectiveDate += $time;
        }
        $callerstrace = Context::get("callerstrace");
        $params = [
            'rid' => $rid,
            'type' => $type,
            'pid' => $goodsId,
            'number' => $num,
            'effective_date' => $effectiveDate,
            'callerstrace' => $callerstrace,
            'effect_time' => $effectTime,
        ];
        $params = http_build_query($params);
        $this->logger->info(__FUNCTION__." init: ".$params);
        $ret = $this->httpClient->get("/package/addunused?_sign=".$this->sign."&".$params);
        $this->logger->info(__FUNCTION__." Return: ".$ret." jsonReturn: ".json_encode($ret));
        $data = json_decode($ret,true);

        return $data;
    }

}

