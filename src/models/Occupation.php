<?php
class OccupationModel
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
        $host = "{$user_prefix}sharingan.pdtv.io";

        $timeout = 2;
        $caller = 'bag';
        $this->logger = XLogKit::logger('badge');
        $this->httpClient = new GHttpClient($host, $this->logger, null, 8360, $timeout,$caller);
    }

    public function addByCate($args){

        $callerstrace = Context::get('callerstrace');

        $params['rid'] = $rid = $args['rid'];
        $params['hostid'] = $hostid = $args['hostid'];
        $params['roomid'] = $roomid = $args['roomid'];
        $params['cate'] = $cate = $args['cate'];
        $params['cost'] = $cost = $args['cost'];
        $params['level'] = $level = $args['level'];
        $params['howlong'] = $howlong = $args['howlong'];
        $params['banner'] = $banner = $args['banner'];
        $params['luckbag'] = $luckbag = $args['luckbag'];
        $params['source'] = 'bag';
        $params['callerstrace'] = $callerstrace;

        $room_type = Context::get("room_type");
        $room_type == isset($room_type) ? $room_type : 1;
        $params['room_type'] = $room_type;

        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , rid=$rid&hostid=$hostid&roomid=$roomid&cate=$cate&cost=$cost&level=$level&howlong=$howlong&banner=$banner&luckbag=$luckbag&source=bag&room_type=$room_type&callerstrace=$callerstrace");

        $ret = $this->httpClient->post('/member/paySuccess', $params);
        $data = json_decode($ret,true);

        if(!isset($data['errno']) || $data['errno']!=0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , rid=$rid&hostid=$hostid&roomid=$roomid&cate=$cate&cost=$cost&level=$level&howlong=$howlong&banner=$banner&luckbag=$luckbag&source=bag&room_type=$room_type,getUserOccupation ret:".$ret);

            return false;

        }

        return $data;
    }
    public function getOccInfo($args){

        $params['rid'] = $rid = $args['rid'];

        $this->logger->info("[".__CLASS__."::".__FUNCTION__."] , rid=$rid");

        $ret = $this->httpClient->get('/member/getUserOccupation?rids='.$params['rid']);

        $data = json_decode($ret,true);
        if(!isset($data['errno']) || $data['errno']!=0){

            $this->logger->error("[".__CLASS__."::".__FUNCTION__."] , rid=$rid,getUserOccupation ret:".$ret);

            return false;
        }


        return $data['data'];
    }


}


