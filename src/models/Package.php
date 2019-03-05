<?php

Class PackageModel {


    private static $ins;
    private $ini = array();

    private $cms_file = 'first_charge.json';
    public static function ins() {
        if(!(self::$ins instanceof self)) {
            self::$ins = new self();
        }
        return self::$ins;
    }
    public function __construct(){
        $this->ini = $this->get_cms_data($this->cms_file);
    }
    // 获取礼包, 一期写死
    public function get($type, $lev = 0) {
        //$this->formatData($comb);
        if($lev == 0){
            return $this->ini[$type];
        }
        return $this->ini[$type][$lev];
    }

    //组合首充实验

    public function formatData($comb = 0){

        $data = $this->getCombGoods($comb);
        if(!empty($data)){
            array_walk($this->ini["v2"], function (&$v,$k) use($data){
                foreach($v['list'] as &$val){
                    if(isset($val['goods_id']) && $val['goods_id'] == 132){
                        $val = $data[$k];
                        break;
                    }
                }
            });
        }
        return $this->ini;
    }


    function getCombGoods($comb, $level = 0){

        $goods = array(
            //组合实验
            1 => array(
                //首充级别
                1=>
                      array(
                          "goods_id"     => 132,
                          "name" => "粉丝经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                      ),
                 2=>
                      array(
                          "goods_id"     => 132,
                          "name" => "粉丝经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
                3=>
                      array(
                          "goods_id"     => 132,
                          "name" => "粉丝经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
            ),
            //组合实验
            2 => array(
                //首充级别
                1=>
                      array(
                          "goods_id"     => 132,
                          "name" => "英雄体验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                      ),
                 2=>
                      array(
                          "goods_id"     => 132,
                          "name" => "英雄体验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
                3=>
                      array(
                          "goods_id"     => 132,
                          "name" => "英雄体验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
            ),
            //组合实验
            3 => array(
                //首充级别
                1=>
                      array(
                          "goods_id"     => 132,
                          "name" => "经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                      ),
                 2=>
                      array(
                          "goods_id"     => 132,
                          "name" => "经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
                3=>
                      array(
                          "goods_id"     => 132,
                          "name" => "经验卡",
                          "instroduction" => "解锁高等级弹幕/价值599猫币",
                          "pictrue" => array(
                              "pc" => "https://i.ssl.pdim.gs/f8503da4d2d870d0cc0535242ff94f56.png",

                          ),
                          "price"         => '599',
                          "num"           => 1,
                         ),
            ),

        );
        if($level == 0){
            return isset($goods[$comb]) ? $goods[$comb] : array();
        }else{
            return isset($goods[$comb][$level]) ? $goods[$comb][$level] : array();
        }
    }

    public function get_first_mobile_data($lev){

        $file = 'first_charge_mobile.json';
        $data = $this->get_cms_data($file);
        if(is_array($data) && !empty($data)){
           return $data;
        }
        return array();

    }

    public function get_cms_data($file){
        $path = '/home/q/cmstpl/bag/'  . $file;
        if(is_file($path)){
            $data = file_get_contents($path);
            if(!trim($data)){
                $data = [];
            }else{
                $data = json_decode($data, true);
                if(json_last_error() != JSON_ERROR_NONE || !is_array($data)) {
                    $data = [];
                }
            }
        }
        return $data;
    }
}
